package langfuse

import (
	"bytes"
	"coda/internal/config"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	DefaultAPIURL = "https://cloud.langfuse.com"
)

// Client is a Langfuse API client
type Client struct {
	publicKey  string
	secretKey  string
	apiURL     string
	httpClient *http.Client
}

// NewClient creates a new Langfuse client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		publicKey:  cfg.LLM.Langfuse.PublicKey,
		secretKey:  cfg.LLM.Langfuse.PrivateKey,
		apiURL:     DefaultAPIURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SetAPIURL allows setting a custom API URL (useful for self-hosted instances)
func (c *Client) SetAPIURL(url string) {
	c.apiURL = url
}

// SetHTTPClient allows setting a custom HTTP client
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// Event represents a base Langfuse event
type Event struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Body      any    `json:"body"`
	Metadata  any    `json:"metadata,omitempty"`
}

// BatchRequest represents a batch request to the Langfuse ingestion API
type BatchRequest struct {
	Batch    []Event `json:"batch"`
	Metadata any     `json:"metadata,omitempty"`
}

// IngestionResponse represents the response from the Langfuse ingestion API
type IngestionResponse struct {
	Successes []struct {
		ID     string `json:"id"`
		Status int    `json:"status"`
	} `json:"successes"`
	Errors []struct {
		ID      string `json:"id"`
		Status  int    `json:"status"`
		Message string `json:"message,omitempty"`
		Error   any    `json:"error,omitempty"`
	} `json:"errors"`
}

// Ingest sends a batch of events to the Langfuse ingestion API
func (c *Client) Ingest(batch []Event) (*IngestionResponse, error) {
	req := BatchRequest{
		Batch: batch,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/api/public/ingestion", c.apiURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add Basic Auth header
	auth := base64.StdEncoding.EncodeToString([]byte(c.publicKey + ":" + c.secretKey))
	httpReq.Header.Add("Authorization", "Basic "+auth)
	httpReq.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ingestionResp IngestionResponse
	if err := json.NewDecoder(resp.Body).Decode(&ingestionResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &ingestionResp, nil
}

// Helper functions to create different types of events

// TraceBody contains the fields for a trace event
type TraceBody struct {
	ID          string   `json:"id"`
	Timestamp   string   `json:"timestamp,omitempty"`
	Name        string   `json:"name,omitempty"`
	UserID      string   `json:"userId,omitempty"`
	Input       any      `json:"input,omitempty"`
	Output      any      `json:"output,omitempty"`
	SessionID   string   `json:"sessionId,omitempty"`
	Release     string   `json:"release,omitempty"`
	Version     string   `json:"version,omitempty"`
	Metadata    any      `json:"metadata,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Environment string   `json:"environment,omitempty"`
	Public      bool     `json:"public,omitempty"`
}

// CreateTrace creates a trace creation event
func CreateTrace(id string, body TraceBody) Event {
	return Event{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "trace-create",
		Body:      body,
	}
}

// GenerationBody contains the fields for a generation event
type GenerationBody struct {
	ID                  string             `json:"id"`
	TraceID             string             `json:"traceId,omitempty"`
	Name                string             `json:"name,omitempty"`
	StartTime           string             `json:"startTime,omitempty"`
	EndTime             string             `json:"endTime,omitempty"`
	CompletionStartTime string             `json:"completionStartTime,omitempty"`
	Model               string             `json:"model,omitempty"`
	ModelParameters     map[string]any     `json:"modelParameters,omitempty"`
	Input               any                `json:"input,omitempty"`
	Output              any                `json:"output,omitempty"`
	Usage               any                `json:"usage,omitempty"`
	UsageDetails        map[string]int     `json:"usageDetails,omitempty"`
	CostDetails         map[string]float64 `json:"costDetails,omitempty"`
	Level               string             `json:"level,omitempty"`
	StatusMessage       string             `json:"statusMessage,omitempty"`
	ParentObservationID string             `json:"parentObservationId,omitempty"`
	Version             string             `json:"version,omitempty"`
	Metadata            any                `json:"metadata,omitempty"`
	Environment         string             `json:"environment,omitempty"`
}

// CreateGeneration creates a generation creation event
func CreateGeneration(id string, body GenerationBody) Event {
	return Event{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "generation-create",
		Body:      body,
	}
}

// UpdateGeneration creates a generation update event
func UpdateGeneration(id string, body GenerationBody) Event {
	return Event{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "generation-update",
		Body:      body,
	}
}

// SpanBody contains the fields for a span event
type SpanBody struct {
	ID                  string `json:"id"`
	TraceID             string `json:"traceId,omitempty"`
	Name                string `json:"name,omitempty"`
	StartTime           string `json:"startTime,omitempty"`
	EndTime             string `json:"endTime,omitempty"`
	Input               any    `json:"input,omitempty"`
	Output              any    `json:"output,omitempty"`
	Level               string `json:"level,omitempty"`
	StatusMessage       string `json:"statusMessage,omitempty"`
	ParentObservationID string `json:"parentObservationId,omitempty"`
	Version             string `json:"version,omitempty"`
	Metadata            any    `json:"metadata,omitempty"`
	Environment         string `json:"environment,omitempty"`
}

// CreateSpan creates a span creation event
func CreateSpan(id string, body SpanBody) Event {
	return Event{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "span-create",
		Body:      body,
	}
}

// UpdateSpan creates a span update event
func UpdateSpan(id string, body SpanBody) Event {
	return Event{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "span-update",
		Body:      body,
	}
}

// ScoreBody contains the fields for a score event
type ScoreBody struct {
	ID            string `json:"id,omitempty"`
	TraceID       string `json:"traceId"`
	Name          string `json:"name"`
	Value         any    `json:"value"`
	ObservationID string `json:"observationId,omitempty"`
	Comment       string `json:"comment,omitempty"`
	DataType      string `json:"dataType,omitempty"`
	ConfigID      string `json:"configId,omitempty"`
	Environment   string `json:"environment,omitempty"`
}

// CreateScore creates a score creation event
func CreateScore(id string, body ScoreBody) Event {
	return Event{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "score-create",
		Body:      body,
	}
}

// EventBody contains the fields for an event observation
type EventBody struct {
	ID                  string `json:"id,omitempty"`
	TraceID             string `json:"traceId,omitempty"`
	Name                string `json:"name,omitempty"`
	StartTime           string `json:"startTime,omitempty"`
	Input               any    `json:"input,omitempty"`
	Output              any    `json:"output,omitempty"`
	Level               string `json:"level,omitempty"`
	StatusMessage       string `json:"statusMessage,omitempty"`
	ParentObservationID string `json:"parentObservationId,omitempty"`
	Version             string `json:"version,omitempty"`
	Metadata            any    `json:"metadata,omitempty"`
	Environment         string `json:"environment,omitempty"`
}

// CreateEvent creates an event creation event
func CreateEvent(id string, body EventBody) Event {
	return Event{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "event-create",
		Body:      body,
	}
}
