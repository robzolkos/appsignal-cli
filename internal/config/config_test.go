package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "appsignal-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `token: test-token
app_id: test-app-id
default_namespace: web
default_state: closed
output:
  format: json
  compact: true
  no_color: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Load the config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify values
	if cfg.Token != "test-token" {
		t.Errorf("Token = %q, want %q", cfg.Token, "test-token")
	}
	if cfg.AppID != "test-app-id" {
		t.Errorf("AppID = %q, want %q", cfg.AppID, "test-app-id")
	}
	if cfg.DefaultNamespace != "web" {
		t.Errorf("DefaultNamespace = %q, want %q", cfg.DefaultNamespace, "web")
	}
	if cfg.DefaultState != "closed" {
		t.Errorf("DefaultState = %q, want %q", cfg.DefaultState, "closed")
	}
	if cfg.Output.Format != "json" {
		t.Errorf("Output.Format = %q, want %q", cfg.Output.Format, "json")
	}
	if !cfg.Output.Compact {
		t.Error("Output.Compact = false, want true")
	}
	if !cfg.Output.NoColor {
		t.Error("Output.NoColor = false, want true")
	}
}

func TestLoadDefaults(t *testing.T) {
	// Load with empty path (non-existent)
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check defaults
	if cfg.DefaultState != "open" {
		t.Errorf("DefaultState = %q, want %q", cfg.DefaultState, "open")
	}
	if cfg.Output.Format != "human" {
		t.Errorf("Output.Format = %q, want %q", cfg.Output.Format, "human")
	}
}

func TestSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appsignal-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		Token:            "save-test-token",
		AppID:            "save-test-app",
		DefaultNamespace: "background",
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Token != cfg.Token {
		t.Errorf("Token = %q, want %q", loaded.Token, cfg.Token)
	}
	if loaded.AppID != cfg.AppID {
		t.Errorf("AppID = %q, want %q", loaded.AppID, cfg.AppID)
	}
}

func TestDefaultConfigContent(t *testing.T) {
	content := DefaultConfigContent()
	if content == "" {
		t.Error("DefaultConfigContent returned empty string")
	}
	if len(content) < 50 {
		t.Error("DefaultConfigContent seems too short")
	}
}
