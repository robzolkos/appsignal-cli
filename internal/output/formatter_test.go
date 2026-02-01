package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

func sampleIncident() *models.Incident {
	return &models.Incident{
		ID:             "test-id",
		Number:         123,
		State:          "open",
		ExceptionName:  "TestError",
		ActionNames:    []string{"TestController#action"},
		Namespace:      "web",
		LastOccurredAt: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		Count:          42,
		Sample: &models.Sample{
			ID:     "sample-id",
			Time:   time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
			Action: "TestController#action",
			Exception: &models.ExceptionDetails{
				Name:    "TestError",
				Message: "Something went wrong",
				Backtrace: []models.BacktraceLine{
					{Line: 10, Path: "app/test.rb", Method: "test_method"},
				},
			},
		},
	}
}

func sampleIncidentList() *models.IncidentList {
	return &models.IncidentList{
		Incidents:   []models.Incident{*sampleIncident()},
		HasNextPage: false,
	}
}

func TestHumanFormatterIncidentList(t *testing.T) {
	f := &HumanFormatter{NoColor: true}
	var buf bytes.Buffer

	err := f.FormatIncidentList(&buf, sampleIncidentList())
	if err != nil {
		t.Fatalf("FormatIncidentList failed: %v", err)
	}

	output := buf.String()
	checks := []string{
		"INCIDENT #123",
		"OPEN",
		"TestError",
		"TestController#action",
		"web",
		"42",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Output missing %q", check)
		}
	}
}

func TestHumanFormatterEmptyList(t *testing.T) {
	f := &HumanFormatter{NoColor: true}
	var buf bytes.Buffer

	err := f.FormatIncidentList(&buf, &models.IncidentList{})
	if err != nil {
		t.Fatalf("FormatIncidentList failed: %v", err)
	}

	if !strings.Contains(buf.String(), "No incidents found") {
		t.Error("Expected 'No incidents found' message")
	}
}

func TestHumanFormatterIncident(t *testing.T) {
	f := &HumanFormatter{NoColor: true}
	var buf bytes.Buffer

	err := f.FormatIncident(&buf, sampleIncident())
	if err != nil {
		t.Fatalf("FormatIncident failed: %v", err)
	}

	output := buf.String()
	checks := []string{
		"INCIDENT #123",
		"TestError",
		"Something went wrong",
		"app/test.rb:10",
		"test_method",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Output missing %q", check)
		}
	}
}

func TestHumanFormatterNoColor(t *testing.T) {
	f := &HumanFormatter{NoColor: true}
	var buf bytes.Buffer

	f.FormatIncident(&buf, sampleIncident())
	output := buf.String()

	// Should not contain ANSI escape codes
	if strings.Contains(output, "\033[") {
		t.Error("NoColor output contains ANSI escape codes")
	}
}

func TestJSONFormatterIncident(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	err := f.FormatIncident(&buf, sampleIncident())
	if err != nil {
		t.Fatalf("FormatIncident failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if result["number"].(float64) != 123 {
		t.Errorf("number = %v, want 123", result["number"])
	}
	if result["state"] != "open" {
		t.Errorf("state = %v, want open", result["state"])
	}
	if result["exception_name"] != "TestError" {
		t.Errorf("exception_name = %v, want TestError", result["exception_name"])
	}
}

func TestJSONFormatterIncidentList(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	err := f.FormatIncidentList(&buf, sampleIncidentList())
	if err != nil {
		t.Fatalf("FormatIncidentList failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	incidents := result["incidents"].([]any)
	if len(incidents) != 1 {
		t.Errorf("Expected 1 incident, got %d", len(incidents))
	}
}

func TestCompactFormatterIncident(t *testing.T) {
	f := &CompactFormatter{}
	var buf bytes.Buffer

	err := f.FormatIncident(&buf, sampleIncident())
	if err != nil {
		t.Fatalf("FormatIncident failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#123") {
		t.Error("Compact output missing incident number")
	}
	if !strings.Contains(output, "TestError") {
		t.Error("Compact output missing exception name")
	}
	if !strings.Contains(output, "Something went wrong") {
		t.Error("Compact output missing error message")
	}
}

func TestCompactFormatterIncidentList(t *testing.T) {
	f := &CompactFormatter{}
	var buf bytes.Buffer

	err := f.FormatIncidentList(&buf, sampleIncidentList())
	if err != nil {
		t.Fatalf("FormatIncidentList failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#123") {
		t.Error("Compact output missing incident number")
	}
}

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name      string
		formatter Formatter
	}{
		{"Human", &HumanFormatter{}},
		{"JSON", &JSONFormatter{}},
		{"Compact", &CompactFormatter{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.formatter.FormatMessage(&buf, "test message")
			if err != nil {
				t.Errorf("FormatMessage failed: %v", err)
			}
			if !strings.Contains(buf.String(), "test message") {
				t.Error("Output missing message")
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format Format
	}{
		{FormatHuman},
		{FormatJSON},
		{FormatCompact},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			f := New(tt.format, Options{})
			if f == nil {
				t.Error("New returned nil")
			}
		})
	}
}
