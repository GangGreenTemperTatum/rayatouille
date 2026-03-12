package nodedetail

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
	actors     []ray.Actor
	actorsErr  error
	logListing *ray.NodeLogListing
	logListErr error
	logContent string
	logFileErr error
}

func (m *mockClient) Ping(_ context.Context) (*ray.VersionInfo, error)          { return nil, nil }
func (m *mockClient) ListJobs(_ context.Context) ([]ray.Job, error)             { return nil, nil }
func (m *mockClient) ListNodes(_ context.Context) ([]ray.Node, error)           { return nil, nil }
func (m *mockClient) ListActors(_ context.Context) ([]ray.Actor, error)         { return m.actors, m.actorsErr }
func (m *mockClient) ListJobDetails(_ context.Context) ([]ray.JobDetail, error) { return nil, nil }
func (m *mockClient) GetJobLogs(_ context.Context, _ string) (string, error)    { return "", nil }
func (m *mockClient) GetTaskSummary(_ context.Context, _ string) (*ray.TaskSummaryResponse, error) {
	return nil, nil
}
func (m *mockClient) GetActorLogs(_ context.Context, _ string) (string, error) { return "", nil }
func (m *mockClient) ListNodeLogs(_ context.Context, _ string) (*ray.NodeLogListing, error) {
	return m.logListing, m.logListErr
}
func (m *mockClient) GetNodeLogFile(_ context.Context, _, _ string) (string, error) {
	return m.logContent, m.logFileErr
}
func (m *mockClient) GetServeApplications(_ context.Context) (*ray.ServeInstanceDetails, error) {
	return nil, nil
}
func (m *mockClient) ListClusterEvents(_ context.Context) ([]ray.ClusterEvent, error) {
	return nil, nil
}

func strPtr(s string) *string { return &s }

func testNode() ray.Node {
	return ray.Node{
		State:       "ALIVE",
		NodeIP:      "10.0.0.1",
		NodeID:      "abcdef1234567890",
		NodeName:    "test-node",
		IsHeadNode:  false,
		StartTimeMs: 1710000000000,
		ResourcesTotal: map[string]float64{
			"CPU":                 4,
			"memory":              17179869184, // 16 GiB
			"object_store_memory": 8589934592,  // 8 GiB
		},
		ResourcesAvailable: map[string]float64{
			"CPU":                 2,
			"memory":              8589934592, // 8 GiB
			"object_store_memory": 4294967296, // 4 GiB
		},
	}
}

func testDeadNode() ray.Node {
	n := testNode()
	n.State = "DEAD"
	n.StateMessage = strPtr("Node lost contact")
	return n
}

func testActors() []ray.Actor {
	return []ray.Actor{
		{
			ActorID:   "actor001abc",
			ClassName: "TrainWorker",
			State:     "ALIVE",
			PID:       12345,
			NodeID:    "abcdef1234567890",
		},
		{
			ActorID:   "actor002def",
			ClassName: "EvalWorker",
			State:     "DEAD",
			PID:       12346,
			NodeID:    "abcdef1234567890",
		},
	}
}

func testLogListing() *ray.NodeLogListing {
	return &ray.NodeLogListing{
		Categories: map[string][]string{
			"worker": {"worker-1.log", "worker-2.log"},
			"gcs":    {"gcs_server.log"},
		},
	}
}

func TestNew_InitialState(t *testing.T) {
	client := &mockClient{}
	node := testNode()
	m := New(client, node)

	assert.Equal(t, SectionInfo, m.section)
	assert.Equal(t, node.NodeID, m.node.NodeID)
	assert.True(t, m.actorsLoading)
	assert.True(t, m.logListLoading)
	assert.Nil(t, m.actors)
	assert.Nil(t, m.actorsErr)
}

func TestNew_DeadNode(t *testing.T) {
	client := &mockClient{}
	node := testDeadNode()
	m := New(client, node)

	assert.True(t, m.actorsLoading)
	assert.False(t, m.logListLoading, "Dead node should not set logListLoading")
}

func TestUpdate_TabCycles3Sections(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())

	assert.Equal(t, SectionInfo, m.section)

	// Tab -> Actors
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionActors, m.section)

	// Tab -> Logs
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionLogs, m.section)

	// Tab -> Info (wraps around)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionInfo, m.section)
}

func TestUpdate_NodeActorsMsg_StoresActors(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())

	actors := testActors()
	m, _ = m.Update(NodeActorsMsg{Actors: actors, Err: nil})

	assert.False(t, m.actorsLoading)
	require.Len(t, m.actors, 2)
	assert.Equal(t, "TrainWorker", m.actors[0].ClassName)
	assert.Nil(t, m.actorsErr)
}

func TestUpdate_NodeActorsMsg_StoresError(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())

	testErr := errors.New("connection refused")
	m, _ = m.Update(NodeActorsMsg{Actors: nil, Err: testErr})

	assert.False(t, m.actorsLoading)
	assert.Nil(t, m.actors)
	assert.Equal(t, testErr, m.actorsErr)
}

func TestUpdate_NodeLogListMsg_FlattensFiles(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())

	listing := testLogListing()
	m, _ = m.Update(NodeLogListMsg{Listing: listing, Err: nil})

	assert.False(t, m.logListLoading)
	require.Len(t, m.logFiles, 3)
	// Categories sorted alphabetically: gcs, worker
	assert.Equal(t, "gcs", m.logFiles[0].Category)
	assert.Equal(t, "gcs_server.log", m.logFiles[0].Filename)
	assert.Equal(t, "worker", m.logFiles[1].Category)
	assert.Equal(t, "worker-1.log", m.logFiles[1].Filename)
	assert.Equal(t, "worker", m.logFiles[2].Category)
	assert.Equal(t, "worker-2.log", m.logFiles[2].Filename)
}

func TestUpdate_NodeLogContentMsg_SetsLogViewer(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())
	m.SetSize(80, 40)

	content := "line 1\nline 2\nline 3"
	m, _ = m.Update(NodeLogContentMsg{Content: content, Filename: "test.log", Err: nil})

	assert.False(t, m.logContentLoading)
	assert.Equal(t, content, m.logContent)
	assert.Nil(t, m.logContentErr)
	assert.True(t, m.logViewer.HasContent())
}

func TestRenderInfo_ContainsNodeFields(t *testing.T) {
	node := testNode()
	output := renderInfo(node, 80)

	assert.Contains(t, output, "abcdef1234567890")
	assert.Contains(t, output, "10.0.0.1")
	assert.Contains(t, output, "ALIVE")
	assert.Contains(t, output, "worker")
}

func TestRenderInfo_HeadNode(t *testing.T) {
	node := testNode()
	node.IsHeadNode = true
	output := renderInfo(node, 80)

	assert.Contains(t, output, "head")
}

func TestRenderInfo_DeadNode(t *testing.T) {
	node := testDeadNode()
	output := renderInfo(node, 80)

	assert.Contains(t, output, "DEAD")
	assert.Contains(t, output, "Node lost contact")
}

func TestRenderInfo_ResourcesAvailable(t *testing.T) {
	node := testNode()
	output := renderInfo(node, 80)

	// Should show used/total format
	assert.Contains(t, output, "2 / 4")    // CPU: 2 used / 4 total
	assert.Contains(t, output, "8.0 GiB")  // Memory values
	assert.Contains(t, output, "16.0 GiB") // Memory total
}

func TestRenderInfo_ResourcesUnavailable(t *testing.T) {
	node := testNode()
	node.ResourcesAvailable = nil
	output := renderInfo(node, 80)

	assert.Contains(t, output, "N/A")
}

func TestRenderActors_Loading(t *testing.T) {
	output := renderActors(nil, true, nil, 80)
	assert.Contains(t, output, "Fetching actors...")
}

func TestRenderActors_Error(t *testing.T) {
	err := errors.New("timeout reached")
	output := renderActors(nil, false, err, 80)
	assert.Contains(t, output, "timeout reached")
}

func TestRenderActors_NoActors(t *testing.T) {
	output := renderActors([]ray.Actor{}, false, nil, 80)
	assert.Contains(t, output, "No actors on this node")
}

func TestRenderActors_WithActors(t *testing.T) {
	actors := testActors()
	output := renderActors(actors, false, nil, 100)

	assert.Contains(t, output, "actor001abc")
	assert.Contains(t, output, "TrainWorker")
	assert.Contains(t, output, "ALIVE")
	assert.Contains(t, output, "DEAD")
	assert.Contains(t, output, "12345")
	assert.Contains(t, output, "Total: 2 actors")
}

func TestRenderLogFileList_Loading(t *testing.T) {
	output := renderLogFileList(nil, 0, true, nil, true, 80)
	assert.Contains(t, output, "Fetching log files...")
}

func TestRenderLogFileList_DeadNode(t *testing.T) {
	output := renderLogFileList(nil, 0, false, nil, false, 80)
	assert.Contains(t, output, "Logs unavailable for dead nodes")
}

func TestRenderLogFileList_Error(t *testing.T) {
	err := errors.New("network error")
	output := renderLogFileList(nil, 0, false, err, true, 80)
	assert.Contains(t, output, "network error")
}

func TestRenderLogFileList_WithFiles(t *testing.T) {
	files := []LogFile{
		{Category: "gcs", Filename: "gcs_server.log"},
		{Category: "worker", Filename: "worker-1.log"},
	}
	output := renderLogFileList(files, 0, false, nil, true, 80)

	assert.Contains(t, output, "gcs/")
	assert.Contains(t, output, "gcs_server.log")
	assert.Contains(t, output, "worker/")
	assert.Contains(t, output, "worker-1.log")
	assert.Contains(t, output, "2 files")
}

func TestSetSize_SetsReady(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())
	assert.False(t, m.ready)

	m.SetSize(120, 40)
	assert.True(t, m.ready)
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestView_ShowsTabBar(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())
	m.SetSize(80, 40)

	view := m.View()
	assert.Contains(t, view, "Info")
	assert.Contains(t, view, "Actors")
	assert.Contains(t, view, "Logs")
}

func TestFlattenLogFiles_SortedByCategoryAndFilename(t *testing.T) {
	categories := map[string][]string{
		"worker": {"worker-2.log", "worker-1.log"},
		"gcs":    {"gcs_server.log"},
	}
	files := flattenLogFiles(categories)

	require.Len(t, files, 3)
	assert.Equal(t, "gcs", files[0].Category)
	assert.Equal(t, "worker", files[1].Category)
	assert.Equal(t, "worker-1.log", files[1].Filename)
	assert.Equal(t, "worker-2.log", files[2].Filename)
}

func TestFlattenLogFiles_EmptyCategories(t *testing.T) {
	files := flattenLogFiles(map[string][]string{})
	assert.Nil(t, files)
}

func TestInit_ReturnsCommand(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())
	cmd := m.Init()
	require.NotNil(t, cmd, "Init should return a command for actors and log listing fetch")
}

func TestInit_DeadNode_SkipsLogFetch(t *testing.T) {
	client := &mockClient{}
	m := New(client, testDeadNode())
	cmd := m.Init()
	require.NotNil(t, cmd, "Init should return a command for actors fetch even for dead nodes")
}

func TestUpdate_LogListMsg_Error(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())

	testErr := errors.New("log fetch failed")
	m, _ = m.Update(NodeLogListMsg{Listing: nil, Err: testErr})

	assert.False(t, m.logListLoading)
	assert.Equal(t, testErr, m.logListErr)
}

func TestUpdate_LogContentMsg_Error(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())

	testErr := errors.New("file not found")
	m, _ = m.Update(NodeLogContentMsg{Content: "", Filename: "test.log", Err: testErr})

	assert.False(t, m.logContentLoading)
	assert.Equal(t, testErr, m.logContentErr)
	assert.False(t, m.logViewer.HasContent())
}

func TestUpdate_NavigateLogFiles(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())

	// Set up log files.
	m, _ = m.Update(NodeLogListMsg{
		Listing: testLogListing(),
		Err:     nil,
	})

	// Switch to Logs tab.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionLogs, m.section)

	// Navigate down with j.
	assert.Equal(t, 0, m.selectedLogFile)
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "j"})
	assert.Equal(t, 1, m.selectedLogFile)

	// Navigate up with k.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "k"})
	assert.Equal(t, 0, m.selectedLogFile)
}

func TestView_InfoSection(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())
	m.SetSize(80, 40)

	view := m.View()
	assert.Contains(t, view, "10.0.0.1")
}

func TestView_ActorsSection_Loading(t *testing.T) {
	client := &mockClient{}
	m := New(client, testNode())
	m.SetSize(80, 40)

	// Switch to Actors tab.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	view := m.View()
	assert.Contains(t, view, "Fetching actors...")
}
