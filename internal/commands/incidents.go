package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robzolkos/appsignal-cli/internal/api"
	"github.com/robzolkos/appsignal-cli/internal/models"
	"github.com/urfave/cli/v2"
)

// IncidentsCommand returns the incidents command
func IncidentsCommand() *cli.Command {
	return &cli.Command{
		Name:  "incidents",
		Usage: "Manage error incidents",
		Subcommands: []*cli.Command{
			incidentsListCommand(),
			incidentsGetCommand(),
			incidentsCloseCommand(),
			incidentsReopenCommand(),
			incidentsExportCommand(),
		},
		Action: func(c *cli.Context) error {
			// Default to list if no subcommand
			return incidentsList(c)
		},
	}
}

func incidentsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List error incidents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "state",
				Usage: "Filter by state: open, closed, wip",
				Value: "open",
			},
			&cli.StringFlag{
				Name:  "namespace",
				Usage: "Filter by namespace (web, background, frontend)",
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "Max results",
				Value: 25,
			},
			&cli.IntFlag{
				Name:  "offset",
				Usage: "Pagination offset",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "since",
				Usage: "Filter incidents after date (ISO 8601, e.g. 2024-01-15 or 2024-01-15T14:30:00Z)",
			},
			&cli.IntFlag{
				Name:  "min-occurrences",
				Usage: "Filter by minimum occurrence count",
				Value: 0,
			},
		},
		Action: incidentsList,
	}
}

func incidentsList(c *cli.Context) error {
	client, err := getClient(c)
	if err != nil {
		return err
	}

	// Parse --since flag
	var sinceTime time.Time
	if since := c.String("since"); since != "" {
		// Try various formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, since); err == nil {
				sinceTime = t
				break
			}
		}
		if sinceTime.IsZero() {
			return fmt.Errorf("invalid date format for --since: %s (use ISO 8601, e.g. 2024-01-15)", since)
		}
	}

	minOccurrences := c.Int("min-occurrences")

	// If filtering, fetch more results to ensure we have enough after filtering
	limit := c.Int("limit")
	fetchLimit := limit
	if !sinceTime.IsZero() || minOccurrences > 0 {
		fetchLimit = limit * 4 // Fetch extra to account for filtering
		if fetchLimit < 100 {
			fetchLimit = 100
		}
	}

	opts := api.ListIncidentsOptions{
		State:  c.String("state"),
		Limit:  fetchLimit,
		Offset: c.Int("offset"),
	}

	if ns := c.String("namespace"); ns != "" {
		opts.Namespaces = []string{ns}
	}

	incidents, err := client.ListIncidents(opts)
	if err != nil {
		return handleError(err)
	}

	// Apply client-side filters
	filtered := make([]models.Incident, 0)
	for _, inc := range incidents.Incidents {
		// Filter by --since
		if !sinceTime.IsZero() && inc.LastOccurredAt.Before(sinceTime) {
			continue
		}
		// Filter by --min-occurrences
		if minOccurrences > 0 && inc.Count < minOccurrences {
			continue
		}
		filtered = append(filtered, inc)
		if len(filtered) >= limit {
			break
		}
	}

	result := &models.IncidentList{
		Incidents:   filtered,
		HasNextPage: len(filtered) >= limit || incidents.HasNextPage,
	}

	formatter := getFormatter(c)
	return formatter.FormatIncidentList(os.Stdout, result)
}

func incidentsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get incident details with sample",
		ArgsUsage: "<number>",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("incident number required")
			}

			number, err := strconv.Atoi(c.Args().First())
			if err != nil {
				return fmt.Errorf("invalid incident number: %s", c.Args().First())
			}

			client, err := getClient(c)
			if err != nil {
				return err
			}

			incident, err := client.GetIncident(number)
			if err != nil {
				return handleError(err)
			}

			formatter := getFormatter(c)
			return formatter.FormatIncident(os.Stdout, incident)
		},
	}
}

func incidentsCloseCommand() *cli.Command {
	return &cli.Command{
		Name:      "close",
		Usage:     "Close/resolve an incident",
		ArgsUsage: "<number>",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("incident number required")
			}

			number, err := strconv.Atoi(c.Args().First())
			if err != nil {
				return fmt.Errorf("invalid incident number: %s", c.Args().First())
			}

			client, err := getClient(c)
			if err != nil {
				return err
			}

			updated, err := client.CloseIncident(number)
			if err != nil {
				return handleError(err)
			}

			formatter := getFormatter(c)
			if c.Bool("json") {
				return formatter.FormatIncident(os.Stdout, updated)
			}
			return formatter.FormatMessage(os.Stdout, fmt.Sprintf("Incident #%d closed.", number))
		},
	}
}

func incidentsReopenCommand() *cli.Command {
	return &cli.Command{
		Name:      "reopen",
		Usage:     "Reopen a closed incident",
		ArgsUsage: "<number>",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("incident number required")
			}

			number, err := strconv.Atoi(c.Args().First())
			if err != nil {
				return fmt.Errorf("invalid incident number: %s", c.Args().First())
			}

			client, err := getClient(c)
			if err != nil {
				return err
			}

			updated, err := client.ReopenIncident(number)
			if err != nil {
				return handleError(err)
			}

			formatter := getFormatter(c)
			if c.Bool("json") {
				return formatter.FormatIncident(os.Stdout, updated)
			}
			return formatter.FormatMessage(os.Stdout, fmt.Sprintf("Incident #%d reopened.", number))
		},
	}
}

func incidentsExportCommand() *cli.Command {
	return &cli.Command{
		Name:      "export",
		Usage:     "Export incident to markdown file",
		ArgsUsage: "<number>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("incident number required")
			}

			number, err := strconv.Atoi(c.Args().First())
			if err != nil {
				return fmt.Errorf("invalid incident number: %s", c.Args().First())
			}

			client, err := getClient(c)
			if err != nil {
				return err
			}

			incident, err := client.GetIncident(number)
			if err != nil {
				return handleError(err)
			}

			markdown := generateMarkdown(incident)

			outputPath := c.String("output")
			if outputPath == "" {
				outputPath = fmt.Sprintf("incident-%d.md", number)
			}

			if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			if !c.Bool("quiet") {
				fmt.Printf("Exported incident #%d to %s\n", number, outputPath)
			}
			return nil
		},
	}
}

func generateMarkdown(incident *models.Incident) string {
	var sb strings.Builder

	action := ""
	if len(incident.ActionNames) > 0 {
		action = " in " + incident.ActionNames[0]
	}

	sb.WriteString(fmt.Sprintf("# Bug Report: %s%s\n\n", incident.ExceptionName, action))

	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("- **Incident**: #%d\n", incident.Number))
	sb.WriteString(fmt.Sprintf("- **Exception**: %s\n", incident.ExceptionName))
	if len(incident.ActionNames) > 0 {
		sb.WriteString(fmt.Sprintf("- **Action**: %s\n", strings.Join(incident.ActionNames, ", ")))
	}
	sb.WriteString(fmt.Sprintf("- **Namespace**: %s\n", incident.Namespace))
	if incident.Count > 0 {
		sb.WriteString(fmt.Sprintf("- **Occurrences**: %d\n", incident.Count))
	}
	sb.WriteString(fmt.Sprintf("- **Last Seen**: %s\n", incident.LastOccurredAt.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString("\n")

	if incident.Sample != nil && incident.Sample.Exception != nil {
		sb.WriteString("## Error Message\n")
		sb.WriteString("```\n")
		sb.WriteString(incident.Sample.Exception.Message)
		sb.WriteString("\n```\n\n")

		if len(incident.Sample.Exception.Backtrace) > 0 {
			sb.WriteString("## Backtrace\n")
			sb.WriteString("```\n")
			for _, line := range incident.Sample.Exception.Backtrace {
				method := line.Method
				if method == "" {
					method = "<unknown>"
				}
				sb.WriteString(fmt.Sprintf("%s:%d in `%s`\n", line.Path, line.Line, method))
			}
			sb.WriteString("```\n\n")
		}

		if incident.Sample.Path != "" || incident.Sample.Method != "" {
			sb.WriteString("## Request Context\n")
			if incident.Sample.Path != "" {
				sb.WriteString(fmt.Sprintf("- **Path**: %s\n", incident.Sample.Path))
			}
			if incident.Sample.Method != "" {
				sb.WriteString(fmt.Sprintf("- **Method**: %s\n", incident.Sample.Method))
			}
			sb.WriteString("\n")
		}

		if incident.Sample.Hostname != "" || incident.Sample.Revision != "" {
			sb.WriteString("## Environment\n")
			if incident.Sample.Hostname != "" {
				sb.WriteString(fmt.Sprintf("- **Hostname**: %s\n", incident.Sample.Hostname))
			}
			if incident.Sample.Revision != "" {
				sb.WriteString(fmt.Sprintf("- **Revision**: %s\n", incident.Sample.Revision))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
