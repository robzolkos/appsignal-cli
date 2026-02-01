package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/robzolkos/appsignal-cli/internal/api"
	"github.com/robzolkos/appsignal-cli/internal/output"
	"github.com/urfave/cli/v2"
)

// Exit codes
const (
	ExitSuccess       = 0
	ExitGeneralError  = 1
	ExitAuthError     = 2
	ExitNotFoundError = 3
	ExitAPIError      = 4
)

// getClient creates an API client from the CLI context
func getClient(c *cli.Context) (*api.Client, error) {
	token := c.String("token")
	if token == "" {
		return nil, fmt.Errorf("API token required: set APPSIGNAL_TOKEN, use --token flag, or configure in .appsignal-cli.yaml")
	}

	appID := c.String("app-id")
	if appID == "" {
		return nil, fmt.Errorf("app ID required: set APPSIGNAL_APP_ID, use --app-id flag, or configure in .appsignal-cli.yaml")
	}

	timeout := time.Duration(c.Int("timeout")) * time.Second
	return api.NewClient(token, appID, timeout), nil
}

// getFormatter returns the appropriate formatter based on flags
func getFormatter(c *cli.Context) output.Formatter {
	format := output.FormatHuman
	if c.Bool("json") {
		format = output.FormatJSON
	} else if c.Bool("compact") {
		format = output.FormatCompact
	}
	return output.New(format, output.Options{
		NoColor: c.Bool("no-color"),
		Verbose: c.Bool("verbose"),
	})
}

// handleError handles errors and returns the appropriate exit code
func handleError(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case *api.AuthError:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitAuthError)
	case *api.NotFoundError:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitNotFoundError)
	case *api.APIError, *api.GraphQLRequestError:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitAPIError)
	default:
		return err
	}
	return nil
}
