package actordetail

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// mockClient implements ray.Client for testing.
type mockClient struct {
	actorLogs    string
	actorLogsErr error
}

func (m *mockClient) Ping(_ context.Context) (*ray.VersionInfo, error)          { return nil, nil }
func (m *mockClient) ListJobs(_ context.Context) ([]ray.Job, error)             { return nil, nil }
func (m *mockClient) ListNodes(_ context.Context) ([]ray.Node, error)           { return nil, nil }
func (m *mockClient) ListActors(_ context.Context) ([]ray.Actor, error)         { return nil, nil }
func (m *mockClient) ListJobDetails(_ context.Context) ([]ray.JobDetail, error) { return nil, nil }
func (m *mockClient) GetJobLogs(_ context.Context, _ string) (string, error)    { return "", nil }
func (m *mockClient) GetTaskSummary(_ context.Context, _ string) (*ray.TaskSummaryResponse, error) {
	return nil, nil
}
func (m *mockClient) ListNodeLogs(_ context.Context, _ string) (*ray.NodeLogListing, error) {
	return nil, nil
}
func (m *mockClient) GetNodeLogFile(_ context.Context, _, _ string) (string, error) { return "", nil }
func (m *mockClient) GetActorLogs(_ context.Context, _ string) (string, error) {
	return m.actorLogs, m.actorLogsErr
}
func (m *mockClient) GetServeApplications(_ context.Context) (*ray.ServeInstanceDetails, error) {
	return nil, nil
}
func (m *mockClient) ListClusterEvents(_ context.Context) ([]ray.ClusterEvent, error) {
	return nil, nil
}

func testActor() ray.Actor {
	return ray.Actor{
		ActorID:      "abc123def456",
		Name:         "my-actor",
		ClassName:    "TrainWorker",
		State:        "ALIVE",
		PID:          12345,
		JobID:        "01000000",
		NodeID:       "node-abc-123",
		RayNamespace: "default",
		IsDetached:   false,
		NumRestarts:  "0",
	}
}

func testDeadActor() ray.Actor {
	a := testActor()
	a.State = "DEAD"
	a.DeathCause = &ray.DeathCause{
		ActorDiedErrorContext: &ray.ActorDiedErrorContext{
			Reason:       "CREATION_TASK_ERROR",
			ErrorMessage: "RuntimeError: training failed",
		},
	}
	return a
}

func TestNewModel(t *testing.T) {
	client := &mockClient{}
	actor := testActor()
	m := New(client, actor)

	assert.Equal(t, SectionInfo, m.section)
	assert.True(t, m.logLoading)
	assert.Equal(t, actor.ActorID, m.actor.ActorID)
}

func TestActorLogMsg_Success(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	m.SetSize(80, 40)

	content := "line 1\nline 2\nline 3"
	m, _ = m.Update(ActorLogMsg{Content: content, Err: nil})

	assert.False(t, m.logLoading)
	assert.Equal(t, content, m.logContent)
	assert.Nil(t, m.logErr)
	assert.True(t, m.logViewer.HasContent())
}

func TestActorLogMsg_Error(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())

	testErr := errors.New("connection refused")
	m, _ = m.Update(ActorLogMsg{Content: "", Err: testErr})

	assert.False(t, m.logLoading)
	assert.Equal(t, testErr, m.logErr)
	assert.False(t, m.logViewer.HasContent())
}

func TestTabCycling(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())

	assert.Equal(t, SectionInfo, m.section)

	// Tab -> Logs
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionLogs, m.section)

	// Tab -> Info (wraps around)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionInfo, m.section)
}

func TestRenderInfo_BasicActor(t *testing.T) {
	actor := testActor()
	output := renderInfo(actor, 80)

	assert.Contains(t, output, "abc123def456")
	assert.Contains(t, output, "TrainWorker")
	assert.Contains(t, output, "ALIVE")
	assert.Contains(t, output, "my-actor")
	assert.Contains(t, output, "01000000")
	assert.Contains(t, output, "node-abc-123")
	assert.Contains(t, output, "default")
}

func TestRenderInfo_DeadActor(t *testing.T) {
	actor := testDeadActor()
	output := renderInfo(actor, 80)

	assert.Contains(t, output, "DEAD")
	assert.Contains(t, output, "Death Cause")
	assert.Contains(t, output, "CREATION_TASK_ERROR")
	assert.Contains(t, output, "RuntimeError: training failed")
}

func TestGoToJobMsg(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())

	// "J" key emits GoToJobMsg.
	m2, cmd := m.Update(tea.KeyPressMsg{Code: -1, Text: "J"})
	require.NotNil(t, cmd)
	_ = m2

	msg := cmd()
	goToJob, ok := msg.(GoToJobMsg)
	require.True(t, ok, "expected GoToJobMsg, got %T", msg)
	assert.Equal(t, "01000000", goToJob.JobID)
}

func TestGoToNodeMsg(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	// Must be on Info section for O to trigger GoToNode.
	assert.Equal(t, SectionInfo, m.section)

	// "O" key emits GoToNodeMsg when on Info section.
	m2, cmd := m.Update(tea.KeyPressMsg{Code: -1, Text: "O"})
	require.NotNil(t, cmd)
	_ = m2

	msg := cmd()
	goToNode, ok := msg.(GoToNodeMsg)
	require.True(t, ok, "expected GoToNodeMsg, got %T", msg)
	assert.Equal(t, "node-abc-123", goToNode.NodeID)
}

func TestInit_ReturnsCommand(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	cmd := m.Init()
	require.NotNil(t, cmd, "Init should return a command for log fetch")
}

func TestSetSize(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	assert.False(t, m.ready)

	m.SetSize(120, 40)
	assert.True(t, m.ready)
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestView_ShowsTabBar(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	m.SetSize(80, 40)

	view := m.View()
	assert.Contains(t, view, "Info")
	assert.Contains(t, view, "Logs")
}

func TestRenderLogs_Loading(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	output := renderLogs(m)
	assert.Contains(t, output, "Loading actor logs...")
}

func TestRenderLogs_Error(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	m.logLoading = false
	m.logErr = errors.New("timeout")
	output := renderLogs(m)
	assert.Contains(t, output, "Error: timeout")
}

func TestRenderLogs_Empty(t *testing.T) {
	client := &mockClient{}
	m := New(client, testActor())
	m.logLoading = false
	output := renderLogs(m)
	assert.Contains(t, output, "No logs available for this actor.")
}

func TestRenderInfo_RequiredResources(t *testing.T) {
	actor := testActor()
	actor.RequiredResources = map[string]any{
		"CPU": 1.0,
	}
	output := renderInfo(actor, 80)
	assert.Contains(t, output, "Required Resources")
	assert.Contains(t, output, "CPU")
}

func TestActorStateStyled(t *testing.T) {
	// Just verify it returns non-empty strings with the state name.
	assert.Contains(t, actorStateStyled("ALIVE"), "ALIVE")
	assert.Contains(t, actorStateStyled("DEAD"), "DEAD")
	assert.Contains(t, actorStateStyled("PENDING_CREATION"), "PENDING_CREATION")
	assert.Contains(t, actorStateStyled("UNKNOWN"), "UNKNOWN")
}
