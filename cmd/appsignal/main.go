package main

import (
	"fmt"
	"os"

	"github.com/robzolkos/appsignal-cli/internal/commands"
	"github.com/robzolkos/appsignal-cli/internal/config"
	"github.com/urfave/cli/v2"
)

var version = "dev"

func main() {
	app := &cli.App{
		Name:    "appsignal",
		Usage:   "CLI for AppSignal error monitoring",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "json",
				Usage: "JSON output (machine-readable)",
			},
			&cli.BoolFlag{
				Name:  "compact",
				Usage: "Minimal output for LLMs (less tokens)",
			},
			&cli.StringFlag{
				Name:  "app-id",
				Usage: "Override app ID",
			},
			&cli.StringFlag{
				Name:  "token",
				Usage: "Override API token",
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "Use specific config file",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Disable colors",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress output except errors",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Verbose output",
			},
			&cli.IntFlag{
				Name:  "timeout",
				Usage: "HTTP request timeout in seconds",
				Value: 30,
			},
		},
		Before: func(c *cli.Context) error {
			cfg, err := config.Load(c.String("config"))
			if err != nil {
				return err
			}

			// Apply config file values, CLI flags override
			if c.String("app-id") == "" && cfg.AppID != "" {
				c.Set("app-id", cfg.AppID)
			}
			if c.String("token") == "" && cfg.Token != "" {
				c.Set("token", cfg.Token)
			}

			// Check environment variables
			if c.String("app-id") == "" {
				if envAppID := os.Getenv("APPSIGNAL_APP_ID"); envAppID != "" {
					c.Set("app-id", envAppID)
				}
			}
			if c.String("token") == "" {
				if envToken := os.Getenv("APPSIGNAL_TOKEN"); envToken != "" {
					c.Set("token", envToken)
				}
			}

			// Apply output settings from config
			if !c.Bool("json") && cfg.Output.Format == "json" {
				c.Set("json", "true")
			}
			if !c.Bool("compact") && (cfg.Output.Format == "compact" || cfg.Output.Compact) {
				c.Set("compact", "true")
			}
			if !c.Bool("no-color") && cfg.Output.NoColor {
				c.Set("no-color", "true")
			}

			return nil
		},
		Commands: []*cli.Command{
			commands.AppsCommand(),
			commands.IncidentsCommand(),
			commands.SamplesCommand(),
			commands.ConfigCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
