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

	binaryPath = filepath.Join(tmpDir, "appsignal-cli")
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

// =============================================================================
// Version and Help
// =============================================================================

func TestVersion(t *testing.T) {
	stdout, _, err := runCLI("--version")
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	if !strings.Contains(stdout, "appsignal-cli version") {
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
		"apps",
		"--json",
		"--compact",
		"--verbose",
		"--no-color",
	}

	for _, check := range checks {
		if !strings.Contains(stdout, check) {
			t.Errorf("Help missing %q", check)
		}
	}
}

// =============================================================================
// Apps Command
// =============================================================================

func TestApps(t *testing.T) {
	stdout, stderr, err := runCLI("apps")
	if err != nil {
		t.Fatalf("apps failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "ORGANIZATION") {
		t.Error("apps output missing ORGANIZATION")
	}
	if !strings.Contains(stdout, "Apps:") {
		t.Error("apps output missing Apps section")
	}
}

func TestAppsJSON(t *testing.T) {
	stdout, stderr, err := runCLI("--json", "apps")
	if err != nil {
		t.Fatalf("apps --json failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if _, ok := result["organizations"]; !ok {
		t.Error("JSON output missing 'organizations' field")
	}
}

func TestAppsCompact(t *testing.T) {
	stdout, stderr, err := runCLI("--compact", "apps")
	if err != nil {
		t.Fatalf("apps --compact failed: %v\nstderr: %s", err, stderr)
	}

	// Compact format shows [slug] Name
	if !strings.Contains(stdout, "[") || !strings.Contains(stdout, "]") {
		t.Error("Compact output doesn't match expected format")
	}
}

// =============================================================================
// Incidents List
// =============================================================================

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

func TestIncidentsListNoColor(t *testing.T) {
	stdout, stderr, err := runCLI("--no-color", "incidents", "list", "--limit", "3")
	if err != nil {
		t.Fatalf("incidents list --no-color failed: %v\nstderr: %s", err, stderr)
	}

	// Should not contain ANSI escape codes
	if strings.Contains(stdout, "\033[") {
		t.Error("--no-color output contains ANSI escape codes")
	}
}

func TestIncidentsListStateFilter(t *testing.T) {
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

func TestIncidentsListNamespaceFilter(t *testing.T) {
	stdout, stderr, err := runCLI("incidents", "list", "--namespace", "frontend", "--limit", "3")
	if err != nil {
		t.Fatalf("incidents list with namespace filter failed: %v\nstderr: %s", err, stderr)
	}
	_ = stdout
}

func TestIncidentsListPagination(t *testing.T) {
	// Get first page
	stdout1, _, err := runCLI("--json", "incidents", "list", "--limit", "2", "--offset", "0")
	if err != nil {
		t.Fatalf("incidents list page 1 failed: %v", err)
	}

	// Get second page
	stdout2, _, err := runCLI("--json", "incidents", "list", "--limit", "2", "--offset", "2")
	if err != nil {
		t.Fatalf("incidents list page 2 failed: %v", err)
	}

	// Pages should be different
	if stdout1 == stdout2 {
		t.Error("Pagination returned same results for different offsets")
	}
}

// =============================================================================
// Incidents Get
// =============================================================================

func getFirstIncidentNumber(t *testing.T) string {
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

	return fmt.Sprintf("%d", result.Incidents[0].Number)
}

func TestIncidentsGet(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	stdout, stderr, err := runCLI("incidents", "get", incidentNumber)
	if err != nil {
		t.Fatalf("incidents get failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "INCIDENT") {
		t.Error("incidents get output doesn't contain expected content")
	}
	if !strings.Contains(stdout, "Exception:") {
		t.Error("incidents get output missing Exception")
	}
}

func TestIncidentsGetJSON(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	stdout, stderr, err := runCLI("--json", "incidents", "get", incidentNumber)
	if err != nil {
		t.Fatalf("incidents get --json failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if _, ok := result["number"]; !ok {
		t.Error("JSON output missing 'number' field")
	}
}

func TestIncidentsGetCompact(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	stdout, stderr, err := runCLI("--compact", "incidents", "get", incidentNumber)
	if err != nil {
		t.Fatalf("incidents get --compact failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "#") {
		t.Error("Compact output missing incident number")
	}
}

func TestIncidentsGetVerbose(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	stdout, stderr, err := runCLI("--verbose", "incidents", "get", incidentNumber)
	if err != nil {
		t.Fatalf("incidents get --verbose failed: %v\nstderr: %s", err, stderr)
	}

	// Verbose should include environment variables section
	if !strings.Contains(stdout, "Environment") {
		t.Error("Verbose output missing Environment section")
	}
}

// =============================================================================
// Incidents Close/Reopen
// =============================================================================

func TestIncidentsCloseAndReopen(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	// Close the incident
	stdout, stderr, err := runCLI("incidents", "close", incidentNumber)
	if err != nil {
		t.Fatalf("incidents close failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "closed") {
		t.Error("Close output doesn't confirm closure")
	}

	// Reopen the incident
	stdout, stderr, err = runCLI("incidents", "reopen", incidentNumber)
	if err != nil {
		t.Fatalf("incidents reopen failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "reopened") {
		t.Error("Reopen output doesn't confirm reopening")
	}
}

func TestIncidentsCloseJSON(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	// Close with JSON output
	stdout, stderr, err := runCLI("--json", "incidents", "close", incidentNumber)
	if err != nil {
		t.Fatalf("incidents close --json failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Reopen to restore state
	runCLI("incidents", "reopen", incidentNumber)
}

// =============================================================================
// Incidents Export
// =============================================================================

func TestIncidentsExport(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	tmpDir, err := os.MkdirTemp("", "appsignal-export-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	exportPath := filepath.Join(tmpDir, "incident.md")

	stdout, stderr, err := runCLI("incidents", "export", "-o", exportPath, incidentNumber)
	if err != nil {
		t.Fatalf("incidents export failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Exported") {
		t.Error("Export output doesn't confirm export")
	}

	// Verify file was created
	content, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	// Check markdown content
	checks := []string{
		"# Bug Report:",
		"## Summary",
		"## Error Message",
		"## Backtrace",
	}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Errorf("Exported markdown missing %q", check)
		}
	}
}

func TestIncidentsExportDefaultFilename(t *testing.T) {
	incidentNumber := getFirstIncidentNumber(t)

	tmpDir, err := os.MkdirTemp("", "appsignal-export-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	_, stderr, err := runCLI("incidents", "export", incidentNumber)
	if err != nil {
		t.Fatalf("incidents export failed: %v\nstderr: %s", err, stderr)
	}

	// Should create incident-<number>.md
	expectedFile := fmt.Sprintf("incident-%s.md", incidentNumber)
	if _, err := os.Stat(filepath.Join(tmpDir, expectedFile)); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", expectedFile)
	}
}

// =============================================================================
// Samples
// =============================================================================

func TestSamplesList(t *testing.T) {
	stdout, stderr, err := runCLI("samples", "list", "--limit", "3")
	if err != nil {
		t.Fatalf("samples list failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "SAMPLE") {
		t.Error("samples list output missing SAMPLE")
	}
}

func TestSamplesListJSON(t *testing.T) {
	stdout, stderr, err := runCLI("--json", "samples", "list", "--limit", "3")
	if err != nil {
		t.Fatalf("samples list --json failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if _, ok := result["samples"]; !ok {
		t.Error("JSON output missing 'samples' field")
	}
}

func TestSamplesGet(t *testing.T) {
	// First get a sample ID
	stdout, _, err := runCLI("--json", "samples", "list", "--limit", "1")
	if err != nil {
		t.Skip("Could not list samples")
	}

	var result struct {
		Samples []struct {
			ID string `json:"id"`
		} `json:"samples"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || len(result.Samples) == 0 {
		t.Skip("No samples available for testing")
	}

	sampleID := result.Samples[0].ID

	stdout, stderr, err := runCLI("samples", "get", sampleID)
	if err != nil {
		t.Fatalf("samples get failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "SAMPLE") {
		t.Error("samples get output missing SAMPLE")
	}
}

// =============================================================================
// Config
// =============================================================================

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

func TestConfigInitAlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appsignal-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create config first time
	runCLI("config", "init")

	// Try to create again - should fail
	_, stderr, err := runCLI("config", "init")
	if err == nil {
		t.Error("config init should fail when file already exists")
	}

	if !strings.Contains(stderr, "already exists") {
		t.Errorf("Expected 'already exists' error, got: %s", stderr)
	}
}

func TestConfigShow(t *testing.T) {
	stdout, _, _ := runCLI("config", "show")

	// Should show something about environment variables at minimum
	if !strings.Contains(stdout, "APPSIGNAL") && !strings.Contains(stdout, "config") {
		t.Log("config show output:", stdout)
	}
}

func TestConfigSet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appsignal-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Set a config value (this will create the file)
	stdout, stderr, err := runCLI("config", "set", "default_namespace", "web")
	if err != nil {
		t.Fatalf("config set failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Set") {
		t.Error("config set output doesn't confirm setting")
	}

	// Verify the file was created with the value
	configPath := filepath.Join(tmpDir, ".appsignal-cli.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if !strings.Contains(string(content), "default_namespace") {
		t.Error("Config file missing set value")
	}
}

func TestConfigSetInvalidKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appsignal-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	_, stderr, err := runCLI("config", "set", "invalid_key", "value")
	if err == nil {
		t.Error("config set should fail for invalid key")
	}

	if !strings.Contains(stderr, "unknown config key") {
		t.Errorf("Expected 'unknown config key' error, got: %s", stderr)
	}
}

// =============================================================================
// Error Cases
// =============================================================================

func TestIncidentsGetNotFound(t *testing.T) {
	_, stderr, err := runCLI("incidents", "get", "999999999")
	if err == nil {
		t.Error("incidents get should fail for non-existent incident")
	}

	if !strings.Contains(stderr, "not found") {
		t.Logf("Expected 'not found' error, got: %s", stderr)
	}
}

func TestIncidentsGetInvalidNumber(t *testing.T) {
	_, stderr, err := runCLI("incidents", "get", "notanumber")
	if err == nil {
		t.Error("incidents get should fail for invalid number")
	}

	if !strings.Contains(stderr, "invalid") {
		t.Errorf("Expected 'invalid' error, got: %s", stderr)
	}
}

func TestMissingRequiredArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"incidents get no number", []string{"incidents", "get"}},
		{"incidents close no number", []string{"incidents", "close"}},
		{"incidents reopen no number", []string{"incidents", "reopen"}},
		{"incidents export no number", []string{"incidents", "export"}},
		{"samples get no id", []string{"samples", "get"}},
		{"config set no args", []string{"config", "set"}},
		{"config set one arg", []string{"config", "set", "key"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, err := runCLI(tt.args...)
			if err == nil {
				t.Errorf("%s should fail", tt.name)
			}
			_ = stderr
		})
	}
}
