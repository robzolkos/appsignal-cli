package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

const updateIncidentMutation = `
mutation UpdateIncident($appId: String!, $number: Int!, $state: IncidentStateEnum!) {
  updateIncident(appId: $appId, number: $number, state: $state) {
    ... on ExceptionIncident {
      id
      number
      state
    }
  }
}
`

// CloseIncident closes an incident
func (c *Client) CloseIncident(number int) (*models.Incident, error) {
	return c.updateIncidentState(number, "CLOSED")
}

// ReopenIncident reopens a closed incident
func (c *Client) ReopenIncident(number int) (*models.Incident, error) {
	return c.updateIncidentState(number, "OPEN")
}

func (c *Client) updateIncidentState(number int, state string) (*models.Incident, error) {
	variables := map[string]any{
		"appId":  c.appID,
		"number": number,
		"state":  strings.ToUpper(state),
	}

	resp, err := c.doGraphQL(GraphQLRequest{
		Query:     updateIncidentMutation,
		Variables: variables,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		UpdateIncident *struct {
			ID     string `json:"id"`
			Number int    `json:"number"`
			State  string `json:"state"`
		} `json:"updateIncident"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.UpdateIncident == nil {
		return nil, fmt.Errorf("incident not found")
	}

	return &models.Incident{
		ID:     result.UpdateIncident.ID,
		Number: result.UpdateIncident.Number,
		State:  result.UpdateIncident.State,
	}, nil
}
