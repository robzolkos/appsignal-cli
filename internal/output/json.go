package output

import (
	"encoding/json"
	"io"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// FormatIncidentList formats a list of incidents as JSON
func (f *JSONFormatter) FormatIncidentList(w io.Writer, incidents *models.IncidentList) error {
	return json.NewEncoder(w).Encode(incidents)
}

// FormatIncident formats a single incident as JSON
func (f *JSONFormatter) FormatIncident(w io.Writer, incident *models.Incident) error {
	return json.NewEncoder(w).Encode(incident)
}

// FormatSampleList formats a list of samples as JSON
func (f *JSONFormatter) FormatSampleList(w io.Writer, samples *models.SampleList) error {
	return json.NewEncoder(w).Encode(samples)
}

// FormatSample formats a single sample as JSON
func (f *JSONFormatter) FormatSample(w io.Writer, sample *models.Sample) error {
	return json.NewEncoder(w).Encode(sample)
}

// FormatMessage formats a message as JSON
func (f *JSONFormatter) FormatMessage(w io.Writer, message string) error {
	return json.NewEncoder(w).Encode(map[string]string{"message": message})
}
