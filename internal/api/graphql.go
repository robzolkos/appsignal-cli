package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

const listIncidentsQuery = `
query ExceptionIncidentsQuery(
  $appId: String!
  $namespaces: [String!]
  $state: IncidentStateEnum
  $limit: Int
  $offset: Int
) {
  app(id: $appId) {
    exceptionIncidents(
      namespaces: $namespaces
      state: $state
      limit: $limit
      offset: $offset
    ) {
      id
      number
      state
      exceptionName
      actionNames
      namespace
      lastOccurredAt
      firstBacktraceLine
      count
    }
  }
}
`

const getIncidentQuery = `
query IncidentQuery($appId: String!, $incidentNumber: Int!) {
  app(id: $appId) {
    incident(incidentNumber: $incidentNumber) {
      ... on ExceptionIncident {
        id
        number
        state
        exceptionName
        actionNames
        namespace
        lastOccurredAt
        count
        sample {
          id
          time
          action
          params
          sessionData
          revision
          environment {
            key
            value
          }
          exception {
            name
            message
            backtrace {
              line
              path
              method
            }
          }
        }
      }
    }
  }
}
`

// ListIncidentsOptions contains options for listing incidents
type ListIncidentsOptions struct {
	State      string
	Namespaces []string
	Limit      int
	Offset     int
}

// ListIncidents lists exception incidents
func (c *Client) ListIncidents(opts ListIncidentsOptions) (*models.IncidentList, error) {
	variables := map[string]any{
		"appId": c.appID,
	}

	if opts.State != "" {
		variables["state"] = strings.ToUpper(opts.State)
	}
	if len(opts.Namespaces) > 0 {
		variables["namespaces"] = opts.Namespaces
	}
	if opts.Limit > 0 {
		variables["limit"] = opts.Limit
	}
	if opts.Offset > 0 {
		variables["offset"] = opts.Offset
	}

	resp, err := c.doGraphQL(GraphQLRequest{
		Query:     listIncidentsQuery,
		Variables: variables,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		App struct {
			ExceptionIncidents []incidentNode `json:"exceptionIncidents"`
		} `json:"app"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse incidents: %w", err)
	}

	incidents := make([]models.Incident, 0, len(result.App.ExceptionIncidents))
	for _, node := range result.App.ExceptionIncidents {
		incidents = append(incidents, node.toIncident())
	}

	// Check if there might be more results
	hasNextPage := len(incidents) >= opts.Limit && opts.Limit > 0

	return &models.IncidentList{
		Incidents:   incidents,
		HasNextPage: hasNextPage,
	}, nil
}

// GetIncident gets a single incident with its sample
func (c *Client) GetIncident(number int) (*models.Incident, error) {
	variables := map[string]any{
		"appId":          c.appID,
		"incidentNumber": number,
	}

	resp, err := c.doGraphQL(GraphQLRequest{
		Query:     getIncidentQuery,
		Variables: variables,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		App struct {
			Incident *incidentWithSampleNode `json:"incident"`
		} `json:"app"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse incident: %w", err)
	}

	if result.App.Incident == nil {
		return nil, &NotFoundError{Resource: "incident", ID: fmt.Sprintf("%d", number)}
	}

	incident := result.App.Incident.toIncident()
	return &incident, nil
}

// incidentNode represents the GraphQL incident node
type incidentNode struct {
	ID                 string   `json:"id"`
	Number             int      `json:"number"`
	State              string   `json:"state"`
	ExceptionName      string   `json:"exceptionName"`
	ActionNames        []string `json:"actionNames"`
	Namespace          string   `json:"namespace"`
	LastOccurredAt     string   `json:"lastOccurredAt"`
	FirstBacktraceLine string   `json:"firstBacktraceLine"`
	Count              int      `json:"count"`
}

// parseJSONField handles JSON fields that may be either a JSON string or a direct object
func parseJSONField(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}

	// Try parsing as a map directly
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err == nil {
		return result
	}

	// Try parsing as a JSON string containing JSON
	var jsonStr string
	if err := json.Unmarshal(raw, &jsonStr); err == nil && jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
			return result
		}
	}

	return nil
}

func (n incidentNode) toIncident() models.Incident {
	lastOccurred, _ := time.Parse(time.RFC3339, n.LastOccurredAt)
	return models.Incident{
		ID:                 n.ID,
		Number:             n.Number,
		State:              n.State,
		ExceptionName:      n.ExceptionName,
		ActionNames:        n.ActionNames,
		Namespace:          n.Namespace,
		LastOccurredAt:     lastOccurred,
		FirstBacktraceLine: n.FirstBacktraceLine,
		Count:              n.Count,
	}
}

// incidentWithSampleNode represents an incident with sample data
type incidentWithSampleNode struct {
	incidentNode
	Sample *sampleNode `json:"sample"`
}

func (n incidentWithSampleNode) toIncident() models.Incident {
	incident := n.incidentNode.toIncident()
	if n.Sample != nil {
		sample := n.Sample.toSample()
		incident.Sample = &sample
	}
	return incident
}

// sampleNode represents the GraphQL sample node
type sampleNode struct {
	ID          string          `json:"id"`
	Time        string          `json:"time"`
	Action      string          `json:"action"`
	Params      json.RawMessage `json:"params"`
	SessionData json.RawMessage `json:"sessionData"`
	Revision    string          `json:"revision"`
	Environment []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"environment"`
	Exception *struct {
		Name      string `json:"name"`
		Message   string `json:"message"`
		Backtrace []struct {
			Line   string `json:"line"`
			Path   string `json:"path"`
			Method string `json:"method"`
		} `json:"backtrace"`
	} `json:"exception"`
}

func (n sampleNode) toSample() models.Sample {
	sampleTime, _ := time.Parse(time.RFC3339, n.Time)

	sample := models.Sample{
		ID:          n.ID,
		Time:        sampleTime,
		Action:      n.Action,
		Params:      parseJSONField(n.Params),
		SessionData: parseJSONField(n.SessionData),
		Revision:    n.Revision,
	}

	for _, env := range n.Environment {
		sample.Environment = append(sample.Environment, models.EnvVar{
			Key:   env.Key,
			Value: env.Value,
		})
	}

	if n.Exception != nil {
		sample.Exception = &models.ExceptionDetails{
			Name:    n.Exception.Name,
			Message: n.Exception.Message,
		}
		for _, bt := range n.Exception.Backtrace {
			line, _ := strconv.Atoi(bt.Line)
			sample.Exception.Backtrace = append(sample.Exception.Backtrace, models.BacktraceLine{
				Line:   line,
				Path:   bt.Path,
				Method: bt.Method,
			})
		}
	}

	return sample
}

const listAppsQuery = `
query ViewerApps {
  viewer {
    organizations {
      id
      slug
      name
      apps {
        id
        name
      }
    }
  }
}
`

// ListApps lists all apps accessible to the current token
func (c *Client) ListApps() (*models.AppList, error) {
	resp, err := c.doGraphQL(GraphQLRequest{
		Query: listAppsQuery,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Viewer struct {
			Organizations []struct {
				ID   string `json:"id"`
				Slug string `json:"slug"`
				Name string `json:"name"`
				Apps []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"apps"`
			} `json:"organizations"`
		} `json:"viewer"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse apps: %w", err)
	}

	appList := &models.AppList{}
	for _, org := range result.Viewer.Organizations {
		organization := models.Organization{
			ID:   org.ID,
			Slug: org.Slug,
			Name: org.Name,
		}
		for _, app := range org.Apps {
			organization.Apps = append(organization.Apps, models.App{
				ID:   app.ID,
				Name: app.Name,
			})
		}
		appList.Organizations = append(appList.Organizations, organization)
	}

	return appList, nil
}
