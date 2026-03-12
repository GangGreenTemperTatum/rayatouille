//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataRetrieval_Jobs(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	jobs, err := client.ListJobs(context.Background())
	require.NoError(t, err)
	require.NotNil(t, jobs, "jobs slice should not be nil")

	if len(jobs) > 0 {
		assert.NotEmpty(t, jobs[0].SubmissionID, "first job SubmissionID should not be empty")
		assert.NotEmpty(t, jobs[0].Status, "first job Status should not be empty")
	}
	t.Logf("Retrieved %d jobs", len(jobs))
}

func TestDataRetrieval_JobDetails(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	details, err := client.ListJobDetails(context.Background())
	require.NoError(t, err)
	require.NotNil(t, details, "job details slice should not be nil")

	if len(details) > 0 {
		assert.NotEmpty(t, details[0].SubmissionID, "first job detail SubmissionID should not be empty")
		assert.NotEmpty(t, details[0].Status, "first job detail Status should not be empty")
		// REST API populates StartTime for non-pending jobs
		if details[0].Status != "PENDING" {
			assert.NotZero(t, details[0].StartTime, "StartTime should be populated for non-pending jobs")
		}
	}
	t.Logf("Retrieved %d job details", len(details))
}

func TestDataRetrieval_Nodes(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	nodes, err := client.ListNodes(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, nodes, "a running cluster should have at least 1 node")

	assert.NotEmpty(t, nodes[0].NodeID, "first node NodeID should not be empty")
	assert.NotEmpty(t, nodes[0].NodeIP, "first node NodeIP should not be empty")
	assert.NotEmpty(t, nodes[0].State, "first node State should not be empty")
	assert.NotEmpty(t, nodes[0].ResourcesTotal, "ResourcesTotal should have at least CPU and memory")

	// Verify head node exists
	hasHead := false
	for _, n := range nodes {
		if n.IsHeadNode {
			hasHead = true
			break
		}
	}
	assert.True(t, hasHead, "at least one node should be the head node")
	t.Logf("Retrieved %d nodes", len(nodes))
}

func TestDataRetrieval_Actors(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	actors, err := client.ListActors(context.Background())
	require.NoError(t, err)
	require.NotNil(t, actors, "actors slice should not be nil")

	if len(actors) > 0 {
		assert.NotEmpty(t, actors[0].ActorID, "first actor ActorID should not be empty")
		assert.NotEmpty(t, actors[0].State, "first actor State should not be empty")
	}
	t.Logf("Retrieved %d actors", len(actors))
}

func TestDataRetrieval_AllEndpoints(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)
	ctx := context.Background()

	version, err := client.Ping(ctx)
	require.NoError(t, err, "Ping should not error")

	jobs, err := client.ListJobs(ctx)
	require.NoError(t, err, "ListJobs should not error")

	nodes, err := client.ListNodes(ctx)
	require.NoError(t, err, "ListNodes should not error")

	actors, err := client.ListActors(ctx)
	require.NoError(t, err, "ListActors should not error")

	details, err := client.ListJobDetails(ctx)
	require.NoError(t, err, "ListJobDetails should not error")

	t.Logf("Ping OK (Ray %s), Jobs: %d, Nodes: %d, Actors: %d, JobDetails: %d",
		version.RayVersion, len(jobs), len(nodes), len(actors), len(details))
}
