package commands

import (
	"fmt"
	"os"

	"github.com/robzolkos/appsignal-cli/internal/config"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// ConfigCommand returns the config command
func ConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration",
		Subcommands: []*cli.Command{
			configInitCommand(),
			configShowCommand(),
			configSetCommand(),
		},
		Action: func(c *cli.Context) error {
			return configShow(c)
		},
	}
}

func configInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize config file in current directory",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "global",
				Usage: "Create in XDG config directory instead",
			},
		},
		Action: func(c *cli.Context) error {
			var path string
			if c.Bool("global") {
				path = config.XDGConfigPath()
				// Create parent directory if needed
				dir := path[:len(path)-len("/config.yaml")]
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create config directory: %w", err)
				}
			} else {
				path = config.InitConfigPath()
			}

			// Check if file already exists
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config file already exists: %s", path)
			}

			content := config.DefaultConfigContent()
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			if !c.Bool("quiet") {
				fmt.Printf("Created config file: %s\n", path)
			}
			return nil
		},
	}
}

func configShowCommand() *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show current configuration",
		Action: configShow,
	}
}

func configShow(c *cli.Context) error {
	configPath := c.String("config")
	if configPath == "" {
		configPath = config.Discover()
	}

	if configPath == "" {
		fmt.Println("No config file found.")
		fmt.Println()
		fmt.Println("Config locations checked:")
		fmt.Println("  - .appsignal-cli.yaml (in current directory or parent directories)")
		fmt.Printf("  - %s\n", config.XDGConfigPath())
		fmt.Println()
		fmt.Println("Environment variables:")
		if token := os.Getenv("APPSIGNAL_TOKEN"); token != "" {
			fmt.Println("  APPSIGNAL_TOKEN: (set)")
		} else {
			fmt.Println("  APPSIGNAL_TOKEN: (not set)")
		}
		if appID := os.Getenv("APPSIGNAL_APP_ID"); appID != "" {
			fmt.Printf("  APPSIGNAL_APP_ID: %s\n", appID)
		} else {
			fmt.Println("  APPSIGNAL_APP_ID: (not set)")
		}
		return nil
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Config file: %s\n\n", configPath)

	if c.Bool("json") {
		data, _ := yaml.Marshal(cfg)
		fmt.Print(string(data))
	} else {
		if cfg.Token != "" {
			fmt.Println("token: (set)")
		}
		if cfg.AppID != "" {
			fmt.Printf("app_id: %s\n", cfg.AppID)
		}
		if cfg.DefaultNamespace != "" {
			fmt.Printf("default_namespace: %s\n", cfg.DefaultNamespace)
		}
		if cfg.DefaultState != "" {
			fmt.Printf("default_state: %s\n", cfg.DefaultState)
		}
		if cfg.Output.Format != "" && cfg.Output.Format != "human" {
			fmt.Printf("output.format: %s\n", cfg.Output.Format)
		}
		if cfg.Output.Compact {
			fmt.Println("output.compact: true")
		}
		if cfg.Output.NoColor {
			fmt.Println("output.no_color: true")
		}
	}

	return nil
}

func configSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set a config value",
		ArgsUsage: "<key> <value>",
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				return fmt.Errorf("usage: config set <key> <value>")
			}

			key := c.Args().Get(0)
			value := c.Args().Get(1)

			// Find or create config file
			configPath := config.Discover()
			if configPath == "" {
				configPath = config.InitConfigPath()
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				cfg = &config.Config{}
			}

			switch key {
			case "token":
				cfg.Token = value
			case "app_id":
				cfg.AppID = value
			case "default_namespace":
				cfg.DefaultNamespace = value
			case "default_state":
				cfg.DefaultState = value
			case "output.format":
				cfg.Output.Format = value
			case "output.compact":
				cfg.Output.Compact = value == "true"
			case "output.no_color":
				cfg.Output.NoColor = value == "true"
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			if err := cfg.Save(configPath); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if !c.Bool("quiet") {
				fmt.Printf("Set %s in %s\n", key, configPath)
			}
			return nil
		},
	}
}
