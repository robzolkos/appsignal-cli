package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

// CompactFormatter formats output in a minimal format for LLMs
type CompactFormatter struct{}

// FormatIncidentList formats a list of incidents in compact format
func (f *CompactFormatter) FormatIncidentList(w io.Writer, incidents *models.IncidentList) error {
	for _, inc := range incidents.Incidents {
		action := ""
		if len(inc.ActionNames) > 0 {
			action = inc.ActionNames[0]
		}

		fmt.Fprintf(w, "#%d %s [%s] %s (%s) %d occurrences\n",
			inc.Number,
			inc.ExceptionName,
			inc.State,
			action,
			inc.Namespace,
			inc.Count)

		if inc.FirstBacktraceLine != "" {
			fmt.Fprintf(w, "  %s\n", inc.FirstBacktraceLine)
		}
	}
	return nil
}

// FormatIncident formats a single incident in compact format
func (f *CompactFormatter) FormatIncident(w io.Writer, incident *models.Incident) error {
	action := ""
	if len(incident.ActionNames) > 0 {
		action = incident.ActionNames[0]
	}

	message := ""
	if incident.Sample != nil && incident.Sample.Exception != nil {
		message = incident.Sample.Exception.Message
	}

	fmt.Fprintf(w, "#%d %s: %s\n", incident.Number, incident.ExceptionName, message)
	fmt.Fprintf(w, "  %s (%s) - %d occurrences\n", action, incident.Namespace, incident.Count)

	if incident.Sample != nil && incident.Sample.Exception != nil {
		for _, line := range incident.Sample.Exception.Backtrace {
			// Only show app code, skip framework/gem files
			if !strings.Contains(line.Path, "/gems/") &&
				!strings.Contains(line.Path, "/ruby/") &&
				!strings.Contains(line.Path, "/vendor/") {
				fmt.Fprintf(w, "  %s:%d\n", line.Path, line.Line)
			}
		}
	}

	return nil
}

// FormatSampleList formats a list of samples in compact format
func (f *CompactFormatter) FormatSampleList(w io.Writer, samples *models.SampleList) error {
	for _, sample := range samples.Samples {
		exName := ""
		if sample.Exception != nil {
			exName = sample.Exception.Name
		}
		fmt.Fprintf(w, "%s %s %s\n", sample.ID, sample.Action, exName)
	}
	return nil
}

// FormatSample formats a single sample in compact format
func (f *CompactFormatter) FormatSample(w io.Writer, sample *models.Sample) error {
	if sample.Exception != nil {
		fmt.Fprintf(w, "%s: %s\n", sample.Exception.Name, sample.Exception.Message)
		for _, line := range sample.Exception.Backtrace {
			if !strings.Contains(line.Path, "/gems/") &&
				!strings.Contains(line.Path, "/ruby/") &&
				!strings.Contains(line.Path, "/vendor/") {
				fmt.Fprintf(w, "  %s:%d\n", line.Path, line.Line)
			}
		}
	}
	return nil
}

// FormatApps formats a list of apps in compact format
func (f *CompactFormatter) FormatApps(w io.Writer, apps *models.AppList) error {
	for _, org := range apps.Organizations {
		fmt.Fprintf(w, "[%s] %s\n", org.Slug, org.Name)
		for _, app := range org.Apps {
			fmt.Fprintf(w, "  %s %s\n", app.ID, app.Name)
		}
	}
	return nil
}

// FormatMessage formats a message in compact format
func (f *CompactFormatter) FormatMessage(w io.Writer, message string) error {
	fmt.Fprintln(w, message)
	return nil
}
