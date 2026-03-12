package jobdetail

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// mockClient implements ray.Client for testing.
type mockClient struct {
	taskSummary *ray.TaskSummaryResponse
	taskErr     error
	logs        string
	logsErr     error
}

func (m *mockClient) Ping(_ context.Context) (*ray.VersionInfo, error)          { return nil, nil }
func (m *mockClient) ListJobs(_ context.Context) ([]ray.Job, error)             { return nil, nil }
func (m *mockClient) ListNodes(_ context.Context) ([]ray.Node, error)           { return nil, nil }
func (m *mockClient) ListActors(_ context.Context) ([]ray.Actor, error)         { return nil, nil }
func (m *mockClient) ListJobDetails(_ context.Context) ([]ray.JobDetail, error) { return nil, nil }
func (m *mockClient) GetJobLogs(_ context.Context, _ string) (string, error) {
	return m.logs, m.logsErr
}
func (m *mockClient) GetTaskSummary(_ context.Context, _ string) (*ray.TaskSummaryResponse, error) {
	return m.taskSummary, m.taskErr
}
func (m *mockClient) GetActorLogs(_ context.Context, _ string) (string, error) { return "", nil }
func (m *mockClient) ListNodeLogs(_ context.Context, _ string) (*ray.NodeLogListing, error) {
	return nil, nil
}
func (m *mockClient) GetNodeLogFile(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (m *mockClient) GetServeApplications(_ context.Context) (*ray.ServeInstanceDetails, error) {
	return nil, nil
}
func (m *mockClient) ListClusterEvents(_ context.Context) ([]ray.ClusterEvent, error) {
	return nil, nil
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func testJob() ray.JobDetail {
	return ray.JobDetail{
		SubmissionID:   "raysubmit_test123",
		JobID:          strPtr("01000000"),
		Status:         "RUNNING",
		Entrypoint:     "python train.py --epochs=10",
		StartTime:      time.Now().Add(-10 * time.Minute).UnixMilli(),
		EndTime:        0,
		DriverNodeID:   "abcdef123456789xyz",
		DriverExitCode: nil,
		Message:        "",
	}
}

func testTaskSummary() *ray.TaskSummaryResponse {
	return &ray.TaskSummaryResponse{
		NodeIDToSummary: map[string]ray.NodeTaskSummary{
			"cluster": {
				Summary: map[string]ray.TaskFuncSummary{
					"train_model": {
						FuncOrClassName: "train_model",
						Type:            "NORMAL_TASK",
						StateCounts:     map[string]int{"FINISHED": 3, "RUNNING": 1},
					},
					"evaluate": {
						FuncOrClassName: "evaluate",
						Type:            "NORMAL_TASK",
						StateCounts:     map[string]int{"PENDING": 2},
					},
				},
				TotalTasks:          6,
				TotalActorTasks:     0,
				TotalActorScheduled: 0,
			},
		},
	}
}

func TestNew_InitialState(t *testing.T) {
	client := &mockClient{}
	job := testJob()
	m := New(client, job)

	assert.Equal(t, SectionMetadata, m.section)
	assert.Equal(t, job.SubmissionID, m.job.SubmissionID)
	assert.True(t, m.loading)
	assert.True(t, m.logsLoading)
	assert.Nil(t, m.taskSummary)
	assert.Nil(t, m.taskErr)
}

func TestUpdate_TabCycles3Sections(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())

	assert.Equal(t, SectionMetadata, m.section)

	// Tab -> Tasks
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionTasks, m.section)

	// Tab -> Logs
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionLogs, m.section)

	// Tab -> Metadata (wraps around)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionMetadata, m.section)
}

func TestUpdate_TaskSummaryMsg_StoresSummary(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())

	summary := testTaskSummary()
	m, _ = m.Update(TaskSummaryMsg{Summary: summary, Err: nil})

	assert.False(t, m.loading)
	require.NotNil(t, m.taskSummary)
	assert.Contains(t, m.taskSummary.NodeIDToSummary, "cluster")
	assert.Nil(t, m.taskErr)
}

func TestUpdate_TaskSummaryMsg_StoresError(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())

	testErr := errors.New("connection refused")
	m, _ = m.Update(TaskSummaryMsg{Summary: nil, Err: testErr})

	assert.False(t, m.loading)
	assert.Nil(t, m.taskSummary)
	assert.Equal(t, testErr, m.taskErr)
}

func TestUpdate_JobLogsMsg_StoresLogs(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	m.SetSize(80, 40)

	logContent := "line 1\nline 2\nline 3"
	m, _ = m.Update(JobLogsMsg{Logs: logContent, Err: nil})

	assert.False(t, m.logsLoading)
	assert.Equal(t, logContent, m.logs)
	assert.Nil(t, m.logsErr)
	assert.True(t, m.logViewer.HasContent())
}

func TestUpdate_JobLogsMsg_StoresError(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())

	testErr := errors.New("timeout")
	m, _ = m.Update(JobLogsMsg{Logs: "", Err: testErr})

	assert.False(t, m.logsLoading)
	assert.Equal(t, "", m.logs)
	assert.Equal(t, testErr, m.logsErr)
	assert.False(t, m.logViewer.HasContent())
}

func TestRenderMetadata_ContainsKeyFields(t *testing.T) {
	job := testJob()
	output := renderMetadata(job, 80)

	assert.Contains(t, output, "raysubmit_test123")
	assert.Contains(t, output, "RUNNING")
	assert.Contains(t, output, "python train.py --epochs=10")
	assert.Contains(t, output, "01000000")
	assert.Contains(t, output, "abcdef123456...")
}

func TestRenderMetadata_NilJobID_ShowsNA(t *testing.T) {
	job := testJob()
	job.JobID = nil
	output := renderMetadata(job, 80)

	assert.Contains(t, output, "N/A")
}

func TestRenderMetadata_WithExitCode(t *testing.T) {
	job := testJob()
	job.DriverExitCode = intPtr(0)
	job.Status = "SUCCEEDED"
	job.EndTime = time.Now().UnixMilli()
	output := renderMetadata(job, 80)

	assert.Contains(t, output, "SUCCEEDED")
	// Exit code 0 should appear.
	assert.Contains(t, output, "0")
}

func TestRenderMetadata_WithMessage(t *testing.T) {
	job := testJob()
	job.Message = "Job completed successfully"
	output := renderMetadata(job, 80)

	assert.Contains(t, output, "Job completed successfully")
}

func TestRenderTasks_Loading(t *testing.T) {
	output := renderTasks(nil, true, nil, 80)
	assert.Contains(t, output, "Fetching task summary...")
}

func TestRenderTasks_Error(t *testing.T) {
	err := errors.New("timeout reached")
	output := renderTasks(nil, false, err, 80)
	assert.Contains(t, output, "timeout reached")
}

func TestRenderTasks_NilSummary(t *testing.T) {
	output := renderTasks(nil, false, nil, 80)
	assert.Contains(t, output, "No tasks found")
}

func TestRenderTasks_ValidSummary(t *testing.T) {
	summary := testTaskSummary()
	output := renderTasks(summary, false, nil, 100)

	assert.Contains(t, output, "train_model")
	assert.Contains(t, output, "evaluate")
	assert.Contains(t, output, "NORMAL_TASK")
	assert.Contains(t, output, "FINISHED")
	assert.Contains(t, output, "RUNNING")
	assert.Contains(t, output, "PENDING")
	assert.Contains(t, output, "Total Tasks: 6")
}

func TestRenderTasks_MissingClusterKey(t *testing.T) {
	summary := &ray.TaskSummaryResponse{
		NodeIDToSummary: map[string]ray.NodeTaskSummary{
			"some-node-id": {},
		},
	}
	output := renderTasks(summary, false, nil, 80)
	assert.Contains(t, output, "No tasks found")
}

func TestInit_NilJobID_SkipsFetch(t *testing.T) {
	client := &mockClient{}
	job := testJob()
	job.JobID = nil
	m := New(client, job)

	cmd := m.Init()
	require.NotNil(t, cmd)
}

func TestInit_ReturnsCommand(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())

	cmd := m.Init()
	require.NotNil(t, cmd, "Init should return a non-nil command for both logs and tasks fetch")
}

func TestSetSize_SetsReady(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	assert.False(t, m.ready)

	m.SetSize(120, 40)
	assert.True(t, m.ready)
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestView_ShowsTabBar(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	m.SetSize(80, 40)

	view := m.View()
	assert.Contains(t, view, "Metadata")
	assert.Contains(t, view, "Tasks")
	assert.Contains(t, view, "Logs")
}

func TestView_MetadataSection(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	m.SetSize(80, 40)

	// Default section is Metadata.
	view := m.View()
	assert.Contains(t, view, "raysubmit_test123")
}

func TestView_TasksSection(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	m.SetSize(80, 40)

	// Switch to Tasks section.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	view := m.View()
	// Should show loading since we haven't received TaskSummaryMsg yet.
	assert.Contains(t, view, "Fetching task summary...")
}

func TestView_LogsSection_Loading(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	m.SetSize(80, 40)

	// Switch to Logs section (2 tabs).
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, SectionLogs, m.section)

	view := m.View()
	assert.Contains(t, view, "Fetching logs...")
}

func TestView_LogsSection_Error(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	m.SetSize(80, 40)

	// Receive error logs.
	m, _ = m.Update(JobLogsMsg{Logs: "", Err: errors.New("network timeout")})

	// Switch to Logs section.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	view := m.View()
	assert.Contains(t, view, "network timeout")
}

func TestView_LogsSection_NoLogs(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())
	m.SetSize(80, 40)

	// Receive empty logs.
	m, _ = m.Update(JobLogsMsg{Logs: "", Err: nil})

	// Switch to Logs section.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	view := m.View()
	assert.Contains(t, view, "No logs available")
}

func TestRefresh_ReturnsBatchCommand(t *testing.T) {
	client := &mockClient{}
	m := New(client, testJob())

	// Receive initial data.
	m, _ = m.Update(TaskSummaryMsg{Summary: testTaskSummary(), Err: nil})
	m, _ = m.Update(JobLogsMsg{Logs: "some logs", Err: nil})

	// Press r to refresh.
	m, cmd := m.Update(tea.KeyPressMsg{Code: -1, Text: "r"})
	require.NotNil(t, cmd, "Refresh should return a batch command for logs and tasks")
	assert.True(t, m.loading)
	assert.True(t, m.logsLoading)
}

func TestFormatStateCounts(t *testing.T) {
	counts := map[string]int{
		"FINISHED": 5,
		"RUNNING":  2,
	}
	output := formatStateCounts(counts)
	assert.Contains(t, output, "FINISHED: 5")
	assert.Contains(t, output, "RUNNING: 2")
}

func TestFormatStateCounts_Empty(t *testing.T) {
	output := formatStateCounts(map[string]int{})
	assert.Equal(t, "-", output)
}

func TestPtrOrNA(t *testing.T) {
	s := "hello"
	assert.Equal(t, "hello", ptrOrNA(&s))
	assert.Equal(t, "N/A", ptrOrNA(nil))
}

func TestTruncateNode(t *testing.T) {
	assert.Equal(t, "N/A", truncateNode(""))
	assert.Equal(t, "short", truncateNode("short"))
	assert.Equal(t, "abcdef123456...", truncateNode("abcdef123456789xyz"))
}

func TestFormatExitCode(t *testing.T) {
	assert.Equal(t, "N/A", formatExitCode(nil))
	code := 0
	assert.Equal(t, "0", formatExitCode(&code))
	code = 1
	assert.Equal(t, "1", formatExitCode(&code))
}

func TestFormatTimestamp(t *testing.T) {
	assert.Equal(t, "-", formatTimestamp(0))
	assert.Equal(t, "-", formatTimestamp(-1))

	ts := time.Date(2024, 6, 15, 14, 30, 0, 0, time.Local).UnixMilli()
	output := formatTimestamp(ts)
	assert.Contains(t, output, "2024-06-15")
	assert.Contains(t, output, "14:30:00")
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "-", formatDuration(0, 0))

	now := time.Now()
	start := now.Add(-5 * time.Minute).UnixMilli()
	end := now.UnixMilli()
	d := formatDuration(start, end)
	assert.Contains(t, d, "5m")
}
