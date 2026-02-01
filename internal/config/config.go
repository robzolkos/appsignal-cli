package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	Token            string       `yaml:"token"`
	AppID            string       `yaml:"app_id"`
	DefaultNamespace string       `yaml:"default_namespace"`
	DefaultState     string       `yaml:"default_state"`
	Output           OutputConfig `yaml:"output"`
}

// OutputConfig represents output settings
type OutputConfig struct {
	Format  string `yaml:"format"` // human, json, compact
	Compact bool   `yaml:"compact"`
	NoColor bool   `yaml:"no_color"`
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	cfg := &Config{
		DefaultState: "open",
		Output: OutputConfig{
			Format: "human",
		},
	}

	// Find config file
	path := configPath
	if path == "" {
		path = Discover()
	}

	// Load from file if exists
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	return cfg, nil
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// DefaultConfigContent returns the default config file content
func DefaultConfigContent() string {
	return `# AppSignal CLI Configuration
# token: your-personal-api-token
# app_id: your-app-id
# default_namespace: web
# default_state: open
# output:
#   format: human  # human, json, compact
#   compact: false
#   no_color: false
`
}
