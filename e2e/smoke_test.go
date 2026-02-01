//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	binaryPath string
	token      string
	appID      string
)

func TestMain(m *testing.M) {
	// Get environment variables
	token = os.Getenv("APPSIGNAL_TOKEN")
	appID = os.Getenv("APPSIGNAL_APP_ID")

	if token == "" || appID == "" {
		fmt.Println("E2E tests require APPSIGNAL_TOKEN and APPSIGNAL_APP_ID environment variables")
		os.Exit(1)
	}

	// Build the binary
	tmpDir, err := os.MkdirTemp("", "appsignal-e2e")
	if err != nil {
		fmt.Println("Failed to create temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "appsignal")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/appsignal")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("Failed to build:", string(output))
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func runCLI(args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(),
		"APPSIGNAL_TOKEN="+token,
		"APPSIGNAL_APP_ID="+appID,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestVersion(t *testing.T) {
	stdout, _, err := runCLI("--version")
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	if !strings.Contains(stdout, "appsignal version") {
		t.Errorf("Unexpected version output: %s", stdout)
	}
}

func TestHelp(t *testing.T) {
	stdout, _, err := runCLI("--help")
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	checks := []string{
		"incidents",
		"samples",
		"config",
		"--json",
		"--compact",
	}

	for _, check := range checks {
		if !strings.Contains(stdout, check) {
			t.Errorf("Help missing %q", check)
		}
	}
}

func TestIncidentsList(t *testing.T) {
	stdout, stderr, err := runCLI("incidents", "list", "--limit", "5")
	if err != nil {
		t.Fatalf("incidents list failed: %v\nstderr: %s", err, stderr)
	}

	if len(stdout) == 0 {
		t.Error("incidents list returned empty output")
	}
}

func TestIncidentsListJSON(t *testing.T) {
	stdout, stderr, err := runCLI("--json", "incidents", "list", "--limit", "3")
	if err != nil {
		t.Fatalf("incidents list --json failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, stdout)
	}

	if _, ok := result["incidents"]; !ok {
		t.Error("JSON output missing 'incidents' field")
	}
}

func TestIncidentsListCompact(t *testing.T) {
	stdout, stderr, err := runCLI("--compact", "incidents", "list", "--limit", "3")
	if err != nil {
		t.Fatalf("incidents list --compact failed: %v\nstderr: %s", err, stderr)
	}

	if len(stdout) == 0 {
		t.Error("incidents list --compact returned empty output")
	}

	if !strings.Contains(stdout, "#") {
		t.Error("Compact output doesn't contain expected format")
	}
}

func TestIncidentsListFilters(t *testing.T) {
	stdout, stderr, err := runCLI("incidents", "list", "--state", "open", "--limit", "3")
	if err != nil {
		t.Fatalf("incidents list with state filter failed: %v\nstderr: %s", err, stderr)
	}

	if strings.Contains(stdout, "[CLOSED]") {
		t.Error("Open filter returned closed incidents")
	}
}

func TestIncidentsListSinceFilter(t *testing.T) {
	stdout, stderr, err := runCLI("incidents", "list", "--since", "2024-01-01", "--limit", "5")
	if err != nil {
		t.Fatalf("incidents list with since filter failed: %v\nstderr: %s", err, stderr)
	}
	_ = stdout
}

func TestIncidentsListMinOccurrences(t *testing.T) {
	stdout, stderr, err := runCLI("incidents", "list", "--min-occurrences", "2", "--limit", "5")
	if err != nil {
		t.Fatalf("incidents list with min-occurrences failed: %v\nstderr: %s", err, stderr)
	}
	_ = stdout
}

func TestConfigInit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appsignal-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	stdout, stderr, err := runCLI("config", "init")
	if err != nil {
		t.Fatalf("config init failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Created config file") {
		t.Errorf("Unexpected output: %s", stdout)
	}

	configPath := filepath.Join(tmpDir, ".appsignal-cli.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestIncidentsGet(t *testing.T) {
	// First get a list to find an incident number
	stdout, _, err := runCLI("--json", "incidents", "list", "--limit", "1")
	if err != nil {
		t.Skip("Could not list incidents")
	}

	var result struct {
		Incidents []struct {
			Number int `json:"number"`
		} `json:"incidents"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || len(result.Incidents) == 0 {
		t.Skip("No incidents available for testing")
	}

	incidentNumber := fmt.Sprintf("%d", result.Incidents[0].Number)

	stdout, stderr, err := runCLI("incidents", "get", incidentNumber)
	if err != nil {
		t.Fatalf("incidents get failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "INCIDENT") {
		t.Error("incidents get output doesn't contain expected content")
	}
}

func TestIncidentsGetJSON(t *testing.T) {
	// First get a list to find an incident number
	stdout, _, err := runCLI("--json", "incidents", "list", "--limit", "1")
	if err != nil {
		t.Skip("Could not list incidents")
	}

	var result struct {
		Incidents []struct {
			Number int `json:"number"`
		} `json:"incidents"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || len(result.Incidents) == 0 {
		t.Skip("No incidents available for testing")
	}

	incidentNumber := fmt.Sprintf("%d", result.Incidents[0].Number)

	stdout, stderr, err := runCLI("--json", "incidents", "get", incidentNumber)
	if err != nil {
		t.Fatalf("incidents get --json failed: %v\nstderr: %s", err, stderr)
	}

	var incidentResult map[string]any
	if err := json.Unmarshal([]byte(stdout), &incidentResult); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if _, ok := incidentResult["number"]; !ok {
		t.Error("JSON output missing 'number' field")
	}
}
