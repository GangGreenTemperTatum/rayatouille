package ray

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client defines the Ray Dashboard API interface.
type Client interface {
	Ping(ctx context.Context) (*VersionInfo, error)
	ListJobs(ctx context.Context) ([]Job, error)
	ListNodes(ctx context.Context) ([]Node, error)
	ListActors(ctx context.Context) ([]Actor, error)
	ListJobDetails(ctx context.Context) ([]JobDetail, error)
	GetJobLogs(ctx context.Context, submissionID string) (string, error)
	GetTaskSummary(ctx context.Context, jobID string) (*TaskSummaryResponse, error)
	ListNodeLogs(ctx context.Context, nodeID string) (*NodeLogListing, error)
	GetNodeLogFile(ctx context.Context, nodeID, filename string) (string, error)
	GetActorLogs(ctx context.Context, actorID string) (string, error)
	GetServeApplications(ctx context.Context) (*ServeInstanceDetails, error)
	ListClusterEvents(ctx context.Context) ([]ClusterEvent, error)
}

// HTTPClient implements Client using the Ray Dashboard REST API.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Ray API client.
// The baseURL should be the Ray Dashboard URL (e.g., "http://localhost:8265").
// Any trailing slash on baseURL is stripped to avoid double-slash in URL construction.
func NewClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Ping checks connectivity to the Ray cluster by calling /api/version.
func (c *HTTPClient) Ping(ctx context.Context) (*VersionInfo, error) {
	url := c.baseURL + VersionEndpoint()

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("pinging Ray cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var info VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding version response: %w", err)
	}

	return &info, nil
}

// ListJobs fetches all jobs from the State API (/api/v0/jobs).
func (c *HTTPClient) ListJobs(ctx context.Context) ([]Job, error) {
	return doStateAPIRequest[Job](c, ctx, JobsStateEndpoint())
}

// ListNodes fetches all nodes from the State API (/api/v0/nodes).
func (c *HTTPClient) ListNodes(ctx context.Context) ([]Node, error) {
	return doStateAPIRequest[Node](c, ctx, NodesStateEndpoint())
}

// ListActors fetches all actors from the State API (/api/v0/actors).
func (c *HTTPClient) ListActors(ctx context.Context) ([]Actor, error) {
	return doStateAPIRequest[Actor](c, ctx, ActorsStateEndpoint())
}

// ListJobDetails fetches all jobs from the Jobs REST API (/api/jobs/).
// This endpoint returns a bare JSON array (not wrapped in StateAPIResponse).
func (c *HTTPClient) ListJobDetails(ctx context.Context) ([]JobDetail, error) {
	url := c.baseURL + JobsRESTEndpoint()

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching job details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var jobs []JobDetail
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, fmt.Errorf("decoding job details response: %w", err)
	}

	return jobs, nil
}

// GetJobLogs fetches the logs for a specific job by submission_id.
func (c *HTTPClient) GetJobLogs(ctx context.Context, submissionID string) (string, error) {
	url := c.baseURL + JobLogsEndpoint(submissionID)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetching job logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var result struct {
		Logs string `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding job logs response: %w", err)
	}

	return result.Logs, nil
}

// GetTaskSummary fetches the task summary for a specific job by job_id.
func (c *HTTPClient) GetTaskSummary(ctx context.Context, jobID string) (*TaskSummaryResponse, error) {
	url := c.baseURL + TasksSummarizeEndpoint(jobID)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching task summary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var apiResp struct {
		Result bool   `json:"result"`
		Msg    string `json:"msg"`
		Data   struct {
			Result struct {
				Result TaskSummaryResponse `json:"result"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding task summary response: %w", err)
	}

	if !apiResp.Result {
		return nil, fmt.Errorf("API error from task summary: %s", apiResp.Msg)
	}

	return &apiResp.Data.Result.Result, nil
}

// ListNodeLogs fetches the categorized log file listing for a specific node.
// The response format is {result: bool, msg: string, data: {result: map[string][]string}}.
func (c *HTTPClient) ListNodeLogs(ctx context.Context, nodeID string) (*NodeLogListing, error) {
	url := c.baseURL + NodeLogsEndpoint(nodeID)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching node logs listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var apiResp struct {
		Result bool   `json:"result"`
		Msg    string `json:"msg"`
		Data   struct {
			Result map[string][]string `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding node logs response: %w", err)
	}

	if !apiResp.Result {
		return nil, fmt.Errorf("API error from node logs: %s", apiResp.Msg)
	}

	return &NodeLogListing{Categories: apiResp.Data.Result}, nil
}

// GetNodeLogFile fetches the raw content of a specific log file from a node.
// The response is raw text (Content-Type: text/plain), not JSON.
func (c *HTTPClient) GetNodeLogFile(ctx context.Context, nodeID, filename string) (string, error) {
	url := c.baseURL + NodeLogFileEndpoint(nodeID, filename)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetching node log file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading node log file body: %w", err)
	}

	return string(body), nil
}

// GetActorLogs fetches the stdout logs for a specific actor by actor_id.
// The response is raw text (Content-Type: text/plain), not JSON.
func (c *HTTPClient) GetActorLogs(ctx context.Context, actorID string) (string, error) {
	url := c.baseURL + ActorLogsEndpoint(actorID)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetching actor logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading actor logs body: %w", err)
	}

	return string(body), nil
}

// GetServeApplications fetches the Serve instance details from /api/serve/applications/.
// Returns nil (not error) when Serve is not running or not deployed.
func (c *HTTPClient) GetServeApplications(ctx context.Context) (*ServeInstanceDetails, error) {
	url := c.baseURL + ServeApplicationsEndpoint()

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, nil // Serve likely not running -- NOT an error
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Serve not deployed
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var details ServeInstanceDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("decoding serve applications response: %w", err)
	}

	return &details, nil
}

// ListClusterEvents fetches all cluster events from the State API (/api/v0/cluster_events).
func (c *HTTPClient) ListClusterEvents(ctx context.Context) ([]ClusterEvent, error) {
	return doStateAPIRequest[ClusterEvent](c, ctx, ClusterEventsEndpoint())
}

// doStateAPIRequest performs a GET request to a State API endpoint and unwraps
// the StateAPIResponse wrapper, returning the inner result slice.
func doStateAPIRequest[T any](c *HTTPClient, ctx context.Context, endpoint string) ([]T, error) {
	url := c.baseURL + endpoint

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var apiResp StateAPIResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding %s response: %w", endpoint, err)
	}

	if !apiResp.Result {
		return nil, fmt.Errorf("API error from %s: %s", endpoint, apiResp.Msg)
	}

	return apiResp.Data.Result.Result, nil
}

// doRequest performs a GET request with context support.
func (c *HTTPClient) doRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", url, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
