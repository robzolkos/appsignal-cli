package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// AppsCommand returns the apps command
func AppsCommand() *cli.Command {
	return &cli.Command{
		Name:  "apps",
		Usage: "List applications (requires org-level access)",
		Action: func(c *cli.Context) error {
			// Apps list requires organization-level API access
			// which is not available with regular app tokens
			return fmt.Errorf("apps list requires organization-level API access, which is not yet implemented")
		},
	}
}
