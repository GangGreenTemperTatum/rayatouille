package ray

import (
	"fmt"
	"net/url"
)

// Endpoint helpers construct the correct URL paths for Ray Dashboard API endpoints.
// CRITICAL: State API endpoints have NO trailing slash. Jobs REST API has trailing slash.
// Getting this wrong causes 404s. Validated against live Ray 2.50.0 cluster.

// JobsStateEndpoint returns the path for the State API jobs endpoint.
func JobsStateEndpoint() string { return "/api/v0/jobs" }

// NodesStateEndpoint returns the path for the State API nodes endpoint.
func NodesStateEndpoint() string { return "/api/v0/nodes" }

// ActorsStateEndpoint returns the path for the State API actors endpoint.
func ActorsStateEndpoint() string { return "/api/v0/actors" }

// JobsRESTEndpoint returns the path for the Jobs REST API endpoint.
// Note: this endpoint requires a trailing slash.
func JobsRESTEndpoint() string { return "/api/jobs/" }

// VersionEndpoint returns the path for the version/ping endpoint.
func VersionEndpoint() string { return "/api/version" }

// JobLogsEndpoint returns the path for fetching logs for a specific job.
// Uses submission_id (e.g., "raysubmit_XXX"), NOT the internal job_id.
func JobLogsEndpoint(submissionID string) string {
	return fmt.Sprintf("/api/jobs/%s/logs", submissionID)
}

// TasksSummarizeEndpoint returns the path for the task summary endpoint filtered by job_id.
// Uses the internal job_id (e.g., "02000000"), NOT submission_id.
func TasksSummarizeEndpoint(jobID string) string {
	return fmt.Sprintf("/api/v0/tasks/summarize?filter_keys=job_id&filter_predicates=%%3D&filter_values=%s", jobID)
}

// NodeLogsEndpoint returns the path for listing log files for a specific node.
func NodeLogsEndpoint(nodeID string) string {
	return fmt.Sprintf("/api/v0/logs?node_id=%s", nodeID)
}

// NodeLogFileEndpoint returns the path for fetching a specific log file from a node.
func NodeLogFileEndpoint(nodeID, filename string) string {
	return fmt.Sprintf("/api/v0/logs/file?node_id=%s&filename=%s", nodeID, url.QueryEscape(filename))
}

// ActorLogsEndpoint returns the path for fetching stdout logs for a specific actor.
func ActorLogsEndpoint(actorID string) string {
	return fmt.Sprintf("/api/v0/logs/file?actor_id=%s&suffix=out&lines=-1", actorID)
}

// ServeApplicationsEndpoint returns the path for the Serve applications endpoint.
// CRITICAL: This endpoint requires a trailing slash. Without it, you get a 404.
func ServeApplicationsEndpoint() string { return "/api/serve/applications/" }

// ClusterEventsEndpoint returns the path for the cluster events State API endpoint.
func ClusterEventsEndpoint() string { return "/api/v0/cluster_events" }
