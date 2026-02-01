package models

import "time"

// Incident represents an AppSignal exception incident
type Incident struct {
	ID                 string    `json:"id"`
	Number             int       `json:"number"`
	State              string    `json:"state"`
	ExceptionName      string    `json:"exception_name"`
	ActionNames        []string  `json:"action_names"`
	Namespace          string    `json:"namespace"`
	LastOccurredAt     time.Time `json:"last_occurred_at"`
	FirstBacktraceLine string    `json:"first_backtrace_line,omitempty"`
	Count              int       `json:"count,omitempty"`
	Sample             *Sample   `json:"sample,omitempty"`
}

// IncidentList represents a paginated list of incidents
type IncidentList struct {
	Incidents   []Incident `json:"incidents"`
	HasNextPage bool       `json:"has_next_page"`
}
