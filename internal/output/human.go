package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// HumanFormatter formats output for human readability
type HumanFormatter struct {
	NoColor bool
	Verbose bool
}

func (f *HumanFormatter) color(code, text string) string {
	if f.NoColor {
		return text
	}
	return code + text + colorReset
}

// FormatIncidentList formats a list of incidents for human reading
func (f *HumanFormatter) FormatIncidentList(w io.Writer, incidents *models.IncidentList) error {
	if len(incidents.Incidents) == 0 {
		fmt.Fprintln(w, "No incidents found.")
		return nil
	}

	for i, inc := range incidents.Incidents {
		if i > 0 {
			fmt.Fprintln(w)
		}
		stateColor := colorGreen
		if inc.State == "open" || inc.State == "OPEN" {
			stateColor = colorRed
		}

		reset := ""
		if !f.NoColor {
			reset = colorReset
		}
		fmt.Fprintf(w, "%s #%d%s [%s]\n",
			f.color(colorBold, "INCIDENT"),
			inc.Number,
			reset,
			f.color(stateColor, strings.ToUpper(inc.State)))

		fmt.Fprintf(w, "Exception: %s\n", f.color(colorYellow, inc.ExceptionName))

		if len(inc.ActionNames) > 0 {
			fmt.Fprintf(w, "Action: %s\n", strings.Join(inc.ActionNames, ", "))
		}

		fmt.Fprintf(w, "Namespace: %s\n", inc.Namespace)
		fmt.Fprintf(w, "Last occurred: %s\n", f.color(colorDim, inc.LastOccurredAt.Format("2006-01-02 15:04:05 UTC")))

		if inc.Count > 0 {
			fmt.Fprintf(w, "Occurrences: %s\n", f.color(colorCyan, fmt.Sprintf("%d", inc.Count)))
		}

		if inc.FirstBacktraceLine != "" {
			fmt.Fprintf(w, "Location: %s\n", f.color(colorDim, inc.FirstBacktraceLine))
		}
	}

	if incidents.HasNextPage {
		fmt.Fprintln(w)
		fmt.Fprintln(w, f.color(colorDim, "(more results available, use --offset to paginate)"))
	}

	return nil
}

// FormatIncident formats a single incident for human reading
func (f *HumanFormatter) FormatIncident(w io.Writer, incident *models.Incident) error {
	stateColor := colorGreen
	if incident.State == "open" || incident.State == "OPEN" {
		stateColor = colorRed
	}

	reset := ""
	if !f.NoColor {
		reset = colorReset
	}
	fmt.Fprintf(w, "%s #%d%s [%s]\n",
		f.color(colorBold, "INCIDENT"),
		incident.Number,
		reset,
		f.color(stateColor, strings.ToUpper(incident.State)))

	fmt.Fprintf(w, "Exception: %s\n", f.color(colorYellow, incident.ExceptionName))

	if len(incident.ActionNames) > 0 {
		fmt.Fprintf(w, "Action: %s\n", strings.Join(incident.ActionNames, ", "))
	}

	fmt.Fprintf(w, "Namespace: %s\n", incident.Namespace)
	fmt.Fprintf(w, "Last occurred: %s\n", f.color(colorDim, incident.LastOccurredAt.Format("2006-01-02 15:04:05 UTC")))

	if incident.Count > 0 {
		fmt.Fprintf(w, "Occurrences: %s\n", f.color(colorCyan, fmt.Sprintf("%d", incident.Count)))
	}

	if incident.Sample != nil && incident.Sample.Exception != nil {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", f.color(colorBold, "Message:"))
		fmt.Fprintf(w, "%s\n", incident.Sample.Exception.Message)

		if len(incident.Sample.Exception.Backtrace) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%s\n", f.color(colorBold, "Backtrace:"))
			for _, line := range incident.Sample.Exception.Backtrace {
				method := line.Method
				if method == "" {
					method = "<unknown>"
				}
				fmt.Fprintf(w, "  %s:%d in `%s`\n",
					f.color(colorCyan, line.Path),
					line.Line,
					method)
			}
		}

		if incident.Sample.Path != "" || incident.Sample.Method != "" {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%s\n", f.color(colorBold, "Request:"))
			if incident.Sample.Path != "" {
				fmt.Fprintf(w, "  Path: %s\n", incident.Sample.Path)
			}
			if incident.Sample.Method != "" {
				fmt.Fprintf(w, "  Method: %s\n", incident.Sample.Method)
			}
		}

		if incident.Sample.Hostname != "" || incident.Sample.Revision != "" {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%s\n", f.color(colorBold, "Environment:"))
			if incident.Sample.Hostname != "" {
				fmt.Fprintf(w, "  Hostname: %s\n", incident.Sample.Hostname)
			}
			if incident.Sample.Revision != "" {
				fmt.Fprintf(w, "  Revision: %s\n", incident.Sample.Revision)
			}
		}

		// Verbose output: params and session data
		if f.Verbose {
			f.formatParamsAndSession(w, incident.Sample)
		}
	}

	return nil
}

func (f *HumanFormatter) formatParamsAndSession(w io.Writer, sample *models.Sample) {
	if len(sample.Params) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", f.color(colorBold, "Params:"))
		f.formatMap(w, sample.Params, "  ")
	}

	if len(sample.SessionData) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", f.color(colorBold, "Session Data:"))
		f.formatMap(w, sample.SessionData, "  ")
	}

	if len(sample.Environment) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", f.color(colorBold, "Environment Variables:"))
		for _, env := range sample.Environment {
			fmt.Fprintf(w, "  %s: %s\n", f.color(colorCyan, env.Key), env.Value)
		}
	}
}

func (f *HumanFormatter) formatMap(w io.Writer, m map[string]any, indent string) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			fmt.Fprintf(w, "%s%s:\n", indent, f.color(colorCyan, k))
			f.formatMap(w, val, indent+"  ")
		case []any:
			data, _ := json.Marshal(val)
			fmt.Fprintf(w, "%s%s: %s\n", indent, f.color(colorCyan, k), string(data))
		default:
			fmt.Fprintf(w, "%s%s: %v\n", indent, f.color(colorCyan, k), v)
		}
	}
}

// FormatSampleList formats a list of samples
func (f *HumanFormatter) FormatSampleList(w io.Writer, samples *models.SampleList) error {
	if len(samples.Samples) == 0 {
		fmt.Fprintln(w, "No samples found.")
		return nil
	}

	for i, sample := range samples.Samples {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s %s\n", f.color(colorBold, "SAMPLE"), sample.ID)
		fmt.Fprintf(w, "Time: %s\n", sample.Time.Format("2006-01-02 15:04:05 UTC"))
		if sample.Action != "" {
			fmt.Fprintf(w, "Action: %s\n", sample.Action)
		}
		if sample.Exception != nil {
			fmt.Fprintf(w, "Exception: %s\n", sample.Exception.Name)
		}
	}

	return nil
}

// FormatSample formats a single sample
func (f *HumanFormatter) FormatSample(w io.Writer, sample *models.Sample) error {
	fmt.Fprintf(w, "%s %s\n", f.color(colorBold, "SAMPLE"), sample.ID)
	fmt.Fprintf(w, "Time: %s\n", sample.Time.Format("2006-01-02 15:04:05 UTC"))

	if sample.Action != "" {
		fmt.Fprintf(w, "Action: %s\n", sample.Action)
	}

	if sample.Path != "" {
		fmt.Fprintf(w, "Path: %s\n", sample.Path)
	}
	if sample.Method != "" {
		fmt.Fprintf(w, "Method: %s\n", sample.Method)
	}
	if sample.Hostname != "" {
		fmt.Fprintf(w, "Hostname: %s\n", sample.Hostname)
	}

	if sample.Exception != nil {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", f.color(colorBold, "Exception:"))
		fmt.Fprintf(w, "Name: %s\n", sample.Exception.Name)
		fmt.Fprintf(w, "Message: %s\n", sample.Exception.Message)

		if len(sample.Exception.Backtrace) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%s\n", f.color(colorBold, "Backtrace:"))
			for _, line := range sample.Exception.Backtrace {
				fmt.Fprintf(w, "  %s:%d in `%s`\n", line.Path, line.Line, line.Method)
			}
		}
	}

	// Always show params and session for sample get
	if f.Verbose || true {
		f.formatParamsAndSession(w, sample)
	}

	return nil
}

// FormatApps formats a list of apps
func (f *HumanFormatter) FormatApps(w io.Writer, apps *models.AppList) error {
	if len(apps.Organizations) == 0 {
		fmt.Fprintln(w, "No organizations found.")
		return nil
	}

	for i, org := range apps.Organizations {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s %s\n", f.color(colorBold, "ORGANIZATION"), org.Name)
		fmt.Fprintf(w, "Slug: %s\n", org.Slug)
		fmt.Fprintf(w, "ID: %s\n", f.color(colorDim, org.ID))

		if len(org.Apps) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%s\n", f.color(colorBold, "Apps:"))
			for _, app := range org.Apps {
				fmt.Fprintf(w, "  %s\n", f.color(colorCyan, app.Name))
				fmt.Fprintf(w, "    ID: %s\n", f.color(colorDim, app.ID))
			}
		}
	}

	return nil
}

// FormatMessage formats a simple message
func (f *HumanFormatter) FormatMessage(w io.Writer, message string) error {
	fmt.Fprintln(w, message)
	return nil
}
