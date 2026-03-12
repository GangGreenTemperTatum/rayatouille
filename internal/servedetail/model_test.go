package servedetail

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func testApp() ray.ApplicationDetails {
	pid1 := 1234
	pid2 := 5678
	nodeIP1 := "10.0.0.1"
	nodeIP2 := "10.0.0.2"
	route := "/api"
	docs := "/docs"

	return ray.ApplicationDetails{
		Name:              "my-app",
		RoutePrefix:       &route,
		DocsPath:          &docs,
		Status:            "RUNNING",
		Message:           "Deployed successfully",
		LastDeployedTimeS: 1700000000,
		Deployments: map[string]ray.DeploymentDetails{
			"frontend": {
				Name:              "frontend",
				Status:            "HEALTHY",
				Message:           "",
				TargetNumReplicas: 2,
				Replicas: []ray.ReplicaDetails{
					{ReplicaID: "replica-abc123456789", State: "RUNNING", PID: &pid1, NodeIP: &nodeIP1, StartTimeS: 1700000000},
					{ReplicaID: "replica-def456", State: "RUNNING", PID: &pid2, NodeIP: &nodeIP2, StartTimeS: 1700000100},
				},
			},
			"backend": {
				Name:              "backend",
				Status:            "HEALTHY",
				Message:           "All replicas running",
				TargetNumReplicas: 3,
				Replicas: []ray.ReplicaDetails{
					{ReplicaID: "replica-ghi789", State: "RUNNING", PID: &pid1, NodeIP: &nodeIP1, StartTimeS: 1700000200},
					{ReplicaID: "replica-jkl012", State: "STARTING", PID: nil, NodeIP: nil, StartTimeS: 0},
				},
			},
		},
	}
}

func TestNew_DeploymentCount(t *testing.T) {
	app := testApp()
	m := New("my-app", app)

	assert.Equal(t, "my-app", m.appName)
	assert.Equal(t, SectionDeployments, m.section)
	assert.Len(t, m.sortedDeploymentNames, 2)
	// Sorted alphabetically: backend, frontend.
	assert.Equal(t, "backend", m.sortedDeploymentNames[0])
	assert.Equal(t, "frontend", m.sortedDeploymentNames[1])
}

func TestNew_EmptyDeployments(t *testing.T) {
	app := ray.ApplicationDetails{
		Status:      "DEPLOYING",
		Deployments: map[string]ray.DeploymentDetails{},
	}
	m := New("empty-app", app)

	assert.Len(t, m.sortedDeploymentNames, 0)
	assert.Contains(t, m.View(), "No deployments")
}

func TestMetadata_RendersNameAndStatus(t *testing.T) {
	app := testApp()
	m := New("my-app", app)

	view := m.View()
	assert.True(t, strings.Contains(view, "my-app"), "should contain app name")
	assert.True(t, strings.Contains(view, "RUNNING"), "should contain status")
}

func TestDeploymentToRow(t *testing.T) {
	app := testApp()
	d := app.Deployments["frontend"]
	row := DeploymentToRow("frontend", d)

	assert.Equal(t, "frontend", row[0])
	assert.Equal(t, "HEALTHY", row[1])
	assert.Equal(t, "2", row[2]) // target
	assert.Equal(t, "2", row[3]) // running (both RUNNING)
	assert.Equal(t, "-", row[4]) // empty message
}

func TestDeploymentToRow_WithMessage(t *testing.T) {
	app := testApp()
	d := app.Deployments["backend"]
	row := DeploymentToRow("backend", d)

	assert.Equal(t, "backend", row[0])
	assert.Equal(t, "1", row[3]) // only 1 RUNNING out of 2
	assert.Equal(t, "All replicas running", row[4])
}

func TestReplicaToRow(t *testing.T) {
	pid := 1234
	nodeIP := "10.0.0.1"
	r := ray.ReplicaDetails{
		ReplicaID:  "replica-abc123456789",
		State:      "RUNNING",
		PID:        &pid,
		NodeIP:     &nodeIP,
		StartTimeS: 1700000000,
	}
	row := ReplicaToRow(r)

	assert.Equal(t, "replica-abc1..", row[0]) // truncated
	assert.Equal(t, "RUNNING", row[1])
	assert.Equal(t, "1234", row[2])
	assert.Equal(t, "10.0.0.1", row[3])
	require.NotEqual(t, "-", row[4]) // should have a relative time
}

func TestReplicaToRow_NilFields(t *testing.T) {
	r := ray.ReplicaDetails{
		ReplicaID:  "short",
		State:      "STARTING",
		PID:        nil,
		NodeIP:     nil,
		StartTimeS: 0,
	}
	row := ReplicaToRow(r)

	assert.Equal(t, "short", row[0])
	assert.Equal(t, "STARTING", row[1])
	assert.Equal(t, "-", row[2])
	assert.Equal(t, "-", row[3])
	assert.Equal(t, "-", row[4])
}

func TestDeploymentColumns_WidthDistribution(t *testing.T) {
	cols := DeploymentColumns(100)
	assert.Len(t, cols, 5)

	totalWidth := 0
	for _, c := range cols {
		totalWidth += c.Width
	}
	assert.LessOrEqual(t, totalWidth, 100)
}

func TestReplicaColumns(t *testing.T) {
	cols := ReplicaColumns(100)
	assert.Len(t, cols, 5)
}
