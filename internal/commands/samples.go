package commands

import (
	"fmt"
	"os"

	"github.com/robzolkos/appsignal-cli/internal/api"
	"github.com/urfave/cli/v2"
)

// SamplesCommand returns the samples command
func SamplesCommand() *cli.Command {
	return &cli.Command{
		Name:  "samples",
		Usage: "Manage error samples",
		Subcommands: []*cli.Command{
			samplesListCommand(),
			samplesGetCommand(),
		},
		Action: func(c *cli.Context) error {
			return samplesList(c)
		},
	}
}

func samplesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List error samples",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "limit",
				Usage: "Max results",
				Value: 25,
			},
		},
		Action: samplesList,
	}
}

func samplesList(c *cli.Context) error {
	client, err := getClient(c)
	if err != nil {
		return err
	}

	opts := api.ListSamplesOptions{
		Limit: c.Int("limit"),
	}

	samples, err := client.ListSamples(opts)
	if err != nil {
		return handleError(err)
	}

	formatter := getFormatter(c)
	return formatter.FormatSampleList(os.Stdout, samples)
}

func samplesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get sample details",
		ArgsUsage: "<id>",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("sample ID required")
			}

			sampleID := c.Args().First()

			client, err := getClient(c)
			if err != nil {
				return err
			}

			sample, err := client.GetSample(sampleID)
			if err != nil {
				return handleError(err)
			}

			formatter := getFormatter(c)
			return formatter.FormatSample(os.Stdout, sample)
		},
	}
}
