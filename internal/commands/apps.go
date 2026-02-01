package commands

import (
	"os"

	"github.com/urfave/cli/v2"
)

// AppsCommand returns the apps command
func AppsCommand() *cli.Command {
	return &cli.Command{
		Name:  "apps",
		Usage: "List applications",
		Action: func(c *cli.Context) error {
			client, err := getClient(c)
			if err != nil {
				return err
			}

			apps, err := client.ListApps()
			if err != nil {
				return handleError(err)
			}

			formatter := getFormatter(c)
			return formatter.FormatApps(os.Stdout, apps)
		},
	}
}
