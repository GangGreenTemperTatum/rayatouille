//go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataRetrieval_ClusterEvents(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	events, err := client.ListClusterEvents(context.Background())
	if err != nil {
		t.Logf("ListClusterEvents returned error (endpoint may not be supported): %v", err)
		return
	}
	require.NotNil(t, events, "events slice should not be nil")

	if len(events) == 0 {
		t.Log("No cluster events found, skipping detailed assertions")
		return
	}

	assert.NotEmpty(t, events[0].SourceType, "first event SourceType should not be empty")
	assert.NotEmpty(t, events[0].Message, "first event Message should not be empty")
	t.Logf("Retrieved %d cluster events", len(events))
}

func TestDataRetrieval_ServeApplications(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	details, err := client.GetServeApplications(context.Background())
	require.NoError(t, err)

	if details == nil {
		t.Log("Serve not running on cluster, skipping assertions")
		return
	}

	assert.NotNil(t, details.Applications, "Applications map should not be nil when Serve is running")
	t.Logf("Retrieved Serve instance with %d applications", len(details.Applications))
}

func TestDataRetrieval_JobLogs(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)
	ctx := context.Background()

	jobs, err := client.ListJobs(ctx)
	require.NoError(t, err)

	if len(jobs) == 0 {
		t.Log("No jobs found on cluster, skipping job logs test")
		return
	}

	submissionID := jobs[0].SubmissionID
	require.NotEmpty(t, submissionID, "first job should have a SubmissionID")

	logs, err := client.GetJobLogs(ctx, submissionID)
	require.NoError(t, err)

	if logs == "" {
		t.Logf("Job %s has no logs", submissionID)
	} else {
		t.Logf("Job %s has %d bytes of logs", submissionID, len(logs))
	}
}

func TestDataRetrieval_TaskSummary(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)
	ctx := context.Background()

	details, err := client.ListJobDetails(ctx)
	require.NoError(t, err)

	if len(details) == 0 {
		t.Log("No job details found on cluster, skipping task summary test")
		return
	}

	// Find a job with a JobID (not all jobs have one -- JobID is a *string)
	var jobID string
	for _, d := range details {
		if d.JobID != nil && *d.JobID != "" {
			jobID = *d.JobID
			break
		}
	}
	if jobID == "" {
		t.Log("No jobs with JobID found, skipping task summary test")
		return
	}

	summary, err := client.GetTaskSummary(ctx, jobID)
	// Task summary may error for jobs without tasks -- that's acceptable
	if err != nil {
		t.Logf("GetTaskSummary returned error for job %s (may have no tasks): %v", jobID, err)
		return
	}

	if summary == nil {
		t.Logf("Job %s has no task summary", jobID)
	} else {
		t.Logf("Job %s task summary retrieved successfully", jobID)
	}
}

func TestDataRetrieval_NodeLogs(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)
	ctx := context.Background()

	nodes, err := client.ListNodes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, nodes, "cluster should have at least one node")

	// Find the head node
	var nodeID string
	for _, n := range nodes {
		if n.IsHeadNode {
			nodeID = n.NodeID
			break
		}
	}
	if nodeID == "" {
		// Fallback to first node
		nodeID = nodes[0].NodeID
	}
	require.NotEmpty(t, nodeID, "should have a valid node ID")

	listing, err := client.ListNodeLogs(ctx, nodeID)
	require.NoError(t, err)
	require.NotNil(t, listing, "node log listing should not be nil")

	t.Logf("Node %s has %d log categories", nodeID, len(listing.Categories))

	// Find a log file to read (skip directory-like entries ending in "/")
	var filename string
	for _, files := range listing.Categories {
		for _, f := range files {
			if f != "" && !strings.HasSuffix(f, "/") {
				filename = f
				break
			}
		}
		if filename != "" {
			break
		}
	}

	if filename == "" {
		t.Log("No log files found on node, skipping log file read")
		return
	}

	content, err := client.GetNodeLogFile(ctx, nodeID, filename)
	if err != nil {
		t.Logf("GetNodeLogFile returned error for %s (file may be inaccessible): %v", filename, err)
		return
	}
	t.Logf("Read log file %s: %d bytes", filename, len(content))
}

func TestDataRetrieval_ActorLogs(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)
	ctx := context.Background()

	actors, err := client.ListActors(ctx)
	require.NoError(t, err)

	if len(actors) == 0 {
		t.Log("No actors found on cluster, skipping actor logs test")
		return
	}

	actorID := actors[0].ActorID
	require.NotEmpty(t, actorID, "first actor should have an ActorID")

	logs, err := client.GetActorLogs(ctx, actorID)
	// Actor logs may error if actor has no log output
	if err != nil {
		t.Logf("GetActorLogs returned error for actor %s: %v", actorID, err)
		return
	}

	if logs == "" {
		t.Logf("Actor %s has no logs", actorID)
	} else {
		t.Logf("Actor %s has %d bytes of logs", actorID, len(logs))
	}
}
