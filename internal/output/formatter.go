package output

import (
	"io"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	FormatIncidentList(w io.Writer, incidents *models.IncidentList) error
	FormatIncident(w io.Writer, incident *models.Incident) error
	FormatSampleList(w io.Writer, samples *models.SampleList) error
	FormatSample(w io.Writer, sample *models.Sample) error
	FormatApps(w io.Writer, apps *models.AppList) error
	FormatMessage(w io.Writer, message string) error
}

// Format represents the output format type
type Format string

const (
	FormatHuman   Format = "human"
	FormatJSON    Format = "json"
	FormatCompact Format = "compact"
)

// Options for formatter creation
type Options struct {
	NoColor bool
	Verbose bool
}

// New creates a new formatter based on the format type
func New(format Format, opts Options) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	case FormatCompact:
		return &CompactFormatter{}
	default:
		return &HumanFormatter{NoColor: opts.NoColor, Verbose: opts.Verbose}
	}
}
