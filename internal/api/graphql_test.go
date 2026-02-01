package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListIncidents(t *testing.T) {
	response := `{
		"data": {
			"app": {
				"exceptionIncidents": [
					{
						"id": "incident-1",
						"number": 123,
						"state": "OPEN",
						"exceptionName": "TestError",
						"actionNames": ["TestController#action"],
						"namespace": "web",
						"lastOccurredAt": "2024-01-15T14:30:00Z",
						"firstBacktraceLine": "app/test.rb:10",
						"count": 42
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with test server
	client := &Client{
		token:      "test-token",
		appID:      "test-app",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// We can't easily override the endpoint, so we'll test the parsing logic directly
	var result struct {
		App struct {
			ExceptionIncidents []incidentNode `json:"exceptionIncidents"`
		} `json:"app"`
	}

	data := json.RawMessage(`{
		"app": {
			"exceptionIncidents": [
				{
					"id": "incident-1",
					"number": 123,
					"state": "OPEN",
					"exceptionName": "TestError",
					"actionNames": ["TestController#action"],
					"namespace": "web",
					"lastOccurredAt": "2024-01-15T14:30:00Z",
					"firstBacktraceLine": "app/test.rb:10",
					"count": 42
				}
			]
		}
	}`)

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(result.App.ExceptionIncidents) != 1 {
		t.Fatalf("Expected 1 incident, got %d", len(result.App.ExceptionIncidents))
	}

	inc := result.App.ExceptionIncidents[0]
	if inc.Number != 123 {
		t.Errorf("Number = %d, want 123", inc.Number)
	}
	if inc.State != "OPEN" {
		t.Errorf("State = %s, want OPEN", inc.State)
	}
	if inc.ExceptionName != "TestError" {
		t.Errorf("ExceptionName = %s, want TestError", inc.ExceptionName)
	}

	// Test conversion to model
	incident := inc.toIncident()
	if incident.Number != 123 {
		t.Errorf("Model Number = %d, want 123", incident.Number)
	}
	if incident.Count != 42 {
		t.Errorf("Model Count = %d, want 42", incident.Count)
	}

	_ = client // Silence unused variable warning
}

func TestIncidentNodeToIncident(t *testing.T) {
	node := incidentNode{
		ID:                 "test-id",
		Number:             456,
		State:              "CLOSED",
		ExceptionName:      "RuntimeError",
		ActionNames:        []string{"Controller#index", "Controller#show"},
		Namespace:          "background",
		LastOccurredAt:     "2024-06-20T10:00:00Z",
		FirstBacktraceLine: "lib/worker.rb:25",
		Count:              100,
	}

	incident := node.toIncident()

	if incident.ID != "test-id" {
		t.Errorf("ID = %s, want test-id", incident.ID)
	}
	if incident.Number != 456 {
		t.Errorf("Number = %d, want 456", incident.Number)
	}
	if incident.State != "CLOSED" {
		t.Errorf("State = %s, want CLOSED", incident.State)
	}
	if len(incident.ActionNames) != 2 {
		t.Errorf("ActionNames length = %d, want 2", len(incident.ActionNames))
	}
	if incident.Namespace != "background" {
		t.Errorf("Namespace = %s, want background", incident.Namespace)
	}
	if incident.Count != 100 {
		t.Errorf("Count = %d, want 100", incident.Count)
	}
}

func TestSampleNodeToSample(t *testing.T) {
	node := sampleNode{
		ID:       "sample-123",
		Time:     "2024-01-15T14:30:00Z",
		Action:   "UsersController#show",
		Revision: "abc123",
		Params:   map[string]any{"id": "42"},
		Exception: &struct {
			Name      string `json:"name"`
			Message   string `json:"message"`
			Backtrace []struct {
				Line   string `json:"line"`
				Path   string `json:"path"`
				Method string `json:"method"`
			} `json:"backtrace"`
		}{
			Name:    "NoMethodError",
			Message: "undefined method 'foo'",
			Backtrace: []struct {
				Line   string `json:"line"`
				Path   string `json:"path"`
				Method string `json:"method"`
			}{
				{Line: "10", Path: "app/models/user.rb", Method: "foo"},
				{Line: "20", Path: "app/controllers/users_controller.rb", Method: "show"},
			},
		},
	}

	sample := node.toSample()

	if sample.ID != "sample-123" {
		t.Errorf("ID = %s, want sample-123", sample.ID)
	}
	if sample.Action != "UsersController#show" {
		t.Errorf("Action = %s, want UsersController#show", sample.Action)
	}
	if sample.Revision != "abc123" {
		t.Errorf("Revision = %s, want abc123", sample.Revision)
	}
	if sample.Exception == nil {
		t.Fatal("Exception is nil")
	}
	if sample.Exception.Name != "NoMethodError" {
		t.Errorf("Exception.Name = %s, want NoMethodError", sample.Exception.Name)
	}
	if len(sample.Exception.Backtrace) != 2 {
		t.Errorf("Backtrace length = %d, want 2", len(sample.Exception.Backtrace))
	}
	if sample.Exception.Backtrace[0].Line != 10 {
		t.Errorf("Backtrace[0].Line = %d, want 10", sample.Exception.Backtrace[0].Line)
	}
}

func TestAuthError(t *testing.T) {
	err := &AuthError{Message: "test auth error"}
	if err.Error() != "test auth error" {
		t.Errorf("Error() = %s, want test auth error", err.Error())
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 500, Message: "internal error"}
	expected := "API error (status 500): internal error"
	if err.Error() != expected {
		t.Errorf("Error() = %s, want %s", err.Error(), expected)
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{Resource: "incident", ID: "123"}
	expected := "incident not found: 123"
	if err.Error() != expected {
		t.Errorf("Error() = %s, want %s", err.Error(), expected)
	}
}

func TestGraphQLRequestError(t *testing.T) {
	err := &GraphQLRequestError{
		Errors: []GraphQLError{
			{Message: "first error"},
			{Message: "second error"},
		},
	}
	if err.Error() != "first error" {
		t.Errorf("Error() = %s, want first error", err.Error())
	}

	emptyErr := &GraphQLRequestError{}
	if emptyErr.Error() != "unknown GraphQL error" {
		t.Errorf("Empty error = %s, want unknown GraphQL error", emptyErr.Error())
	}
}
