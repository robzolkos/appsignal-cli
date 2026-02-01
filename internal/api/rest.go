package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/robzolkos/appsignal-cli/internal/models"
)

const restBaseURL = "https://appsignal.com/api"

// ListSamplesOptions contains options for listing samples
type ListSamplesOptions struct {
	Limit  int
	Since  time.Time
	Before time.Time
}

// ListSamples lists error samples via REST API
func (c *Client) ListSamples(opts ListSamplesOptions) (*models.SampleList, error) {
	endpoint := fmt.Sprintf("%s/%s/samples/errors.json", restBaseURL, c.appID)

	params := url.Values{}
	params.Set("token", c.token)
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if !opts.Since.IsZero() {
		params.Set("since", opts.Since.Format(time.RFC3339))
	}
	if !opts.Before.IsZero() {
		params.Set("before", opts.Before.Format(time.RFC3339))
	}

	reqURL := endpoint + "?" + params.Encode()
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, &AuthError{Message: "authentication failed: invalid or expired token"}
	}

	if resp.StatusCode == 404 {
		return nil, &NotFoundError{Resource: "app", ID: c.appID}
	}

	if resp.StatusCode != 200 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var response struct {
		Count      int            `json:"count"`
		LogEntries []restLogEntry `json:"log_entries"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse samples: %w", err)
	}

	result := &models.SampleList{
		Samples: make([]models.Sample, 0, len(response.LogEntries)),
	}
	for _, entry := range response.LogEntries {
		result.Samples = append(result.Samples, entry.toSample())
	}

	return result, nil
}

// GetSample gets a specific sample by ID via REST API
func (c *Client) GetSample(sampleID string) (*models.Sample, error) {
	endpoint := fmt.Sprintf("%s/%s/samples/%s.json", restBaseURL, c.appID, sampleID)

	params := url.Values{}
	params.Set("token", c.token)

	reqURL := endpoint + "?" + params.Encode()
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, &AuthError{Message: "authentication failed: invalid or expired token"}
	}

	if resp.StatusCode == 404 {
		return nil, &NotFoundError{Resource: "sample", ID: sampleID}
	}

	if resp.StatusCode != 200 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var s restSampleDetail
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, fmt.Errorf("failed to parse sample: %w", err)
	}

	sample := s.toSample()
	return &sample, nil
}

// restLogEntry represents a sample entry from the list API
type restLogEntry struct {
	ID          string `json:"id"`
	Action      string `json:"action"`
	Path        string `json:"path"`
	Duration    *int   `json:"duration"`
	Status      *int   `json:"status"`
	Time        int64  `json:"time"`
	IsException bool   `json:"is_exception"`
	Exception   *struct {
		Name string `json:"name"`
	} `json:"exception"`
}

func (e restLogEntry) toSample() models.Sample {
	sample := models.Sample{
		ID:     e.ID,
		Time:   time.Unix(e.Time, 0),
		Action: e.Action,
		Path:   e.Path,
	}

	if e.Exception != nil {
		sample.Exception = &models.ExceptionDetails{
			Name: e.Exception.Name,
		}
	}

	return sample
}

// restSampleDetail represents a detailed sample from the get API
type restSampleDetail struct {
	ID        string         `json:"id"`
	Time      int64          `json:"time"`
	Action    string         `json:"action"`
	Path      string         `json:"path"`
	Method    string         `json:"request_method"`
	Hostname  string         `json:"hostname"`
	Revision  string         `json:"revision"`
	Params    map[string]any `json:"params"`
	Session   map[string]any `json:"session_data"`
	Exception *struct {
		Name      string   `json:"name"`
		Message   string   `json:"message"`
		Backtrace []string `json:"backtrace"`
	} `json:"exception"`
	Environment map[string]string `json:"environment"`
}

func (s restSampleDetail) toSample() models.Sample {
	sample := models.Sample{
		ID:          s.ID,
		Time:        time.Unix(s.Time, 0),
		Action:      s.Action,
		Path:        s.Path,
		Method:      s.Method,
		Hostname:    s.Hostname,
		Revision:    s.Revision,
		Params:      s.Params,
		SessionData: s.Session,
	}

	for k, v := range s.Environment {
		sample.Environment = append(sample.Environment, models.EnvVar{
			Key:   k,
			Value: v,
		})
	}

	if s.Exception != nil {
		sample.Exception = &models.ExceptionDetails{
			Name:    s.Exception.Name,
			Message: s.Exception.Message,
		}
		for _, bt := range s.Exception.Backtrace {
			// Parse backtrace string format: "method@file:line:column" or just "file:line"
			line := models.BacktraceLine{Path: bt}
			if atIdx := strings.LastIndex(bt, "@"); atIdx != -1 {
				line.Method = bt[:atIdx]
				line.Path = bt[atIdx+1:]
			}
			// Try to extract line number from path like "file.js:19:265686"
			if colonIdx := strings.LastIndex(line.Path, ":"); colonIdx != -1 {
				if secondColonIdx := strings.LastIndex(line.Path[:colonIdx], ":"); secondColonIdx != -1 {
					if lineNum, err := strconv.Atoi(line.Path[secondColonIdx+1 : colonIdx]); err == nil {
						line.Line = lineNum
						line.Path = line.Path[:secondColonIdx]
					}
				}
			}
			sample.Exception.Backtrace = append(sample.Exception.Backtrace, line)
		}
	}

	return sample
}
