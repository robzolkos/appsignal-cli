package models

import "time"

// Sample represents an error sample from AppSignal
type Sample struct {
	ID          string            `json:"id"`
	Time        time.Time         `json:"time"`
	Action      string            `json:"action"`
	Params      map[string]any    `json:"params,omitempty"`
	SessionData map[string]any    `json:"session_data,omitempty"`
	Environment []EnvVar          `json:"environment,omitempty"`
	Exception   *ExceptionDetails `json:"exception,omitempty"`
	Hostname    string            `json:"hostname,omitempty"`
	Revision    string            `json:"revision,omitempty"`
	Path        string            `json:"path,omitempty"`
	Method      string            `json:"method,omitempty"`
}

// EnvVar represents an environment variable key-value pair
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExceptionDetails contains exception information
type ExceptionDetails struct {
	Name      string          `json:"name"`
	Message   string          `json:"message"`
	Backtrace []BacktraceLine `json:"backtrace"`
}

// BacktraceLine represents a single line in a backtrace
type BacktraceLine struct {
	Line   int    `json:"line"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

// SampleList represents a list of samples
type SampleList struct {
	Samples []Sample `json:"samples"`
}
