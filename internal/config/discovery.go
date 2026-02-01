package config

import (
	"os"
	"path/filepath"
)

const configFileName = ".appsignal-cli.yaml"

// Discover finds the configuration file by walking up directories
func Discover() string {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return tryXDGConfig()
	}

	// Walk up looking for config file, stop at git root or filesystem root
	for {
		configPath := filepath.Join(dir, configFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		// Check if we're at a git root
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			// We're at git root, stop looking up
			break
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached filesystem root
			break
		}
		dir = parent
	}

	// Try XDG config directory
	return tryXDGConfig()
}

// tryXDGConfig checks for config in XDG_CONFIG_HOME
func tryXDGConfig() string {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		xdgConfig = filepath.Join(home, ".config")
	}

	configPath := filepath.Join(xdgConfig, "appsignal-cli", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

// InitConfigPath returns the path where config init should create the file
func InitConfigPath() string {
	return configFileName
}

// XDGConfigPath returns the XDG config path
func XDGConfigPath() string {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		xdgConfig = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgConfig, "appsignal-cli", "config.yaml")
}
