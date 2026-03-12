package dashboard

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// testModel constructs a Model with pre-populated health and jobs data.
func testModel(nodes []ray.Node, jobs []ray.JobDetail) Model {
	barOpts := []progress.Option{
		progress.WithWidth(30),
		progress.WithDefaultBlend(),
		progress.WithoutPercentage(),
	}
	m := Model{
		width:   80,
		height:  24,
		cpuBar:  progress.New(barOpts...),
		gpuBar:  progress.New(barOpts...),
		memBar:  progress.New(barOpts...),
		diskBar: progress.New(barOpts...),
	}
	m.nodes = nodes
	m.jobDetails = jobs
	m.health = AggregateClusterHealth(nodes)
	m.jobs = AggregateJobSummary(jobs)
	return m
}

func makeNode(state string, cpu, gpu, memory, objStore float64) ray.Node {
	res := map[string]float64{
		"CPU":    cpu,
		"memory": memory,
	}
	if gpu > 0 {
		res["GPU"] = gpu
	}
	if objStore > 0 {
		res["object_store_memory"] = objStore
	}
	return ray.Node{
		State:          state,
		ResourcesTotal: res,
	}
}

// makeNodeWithAvailable creates a node with both resources_total and resources_available.
func makeNodeWithAvailable(state string, cpuTotal, cpuAvail, gpuTotal, gpuAvail, memTotal, memAvail, objTotal, objAvail float64) ray.Node {
	total := map[string]float64{
		"CPU":    cpuTotal,
		"memory": memTotal,
	}
	avail := map[string]float64{
		"CPU":    cpuAvail,
		"memory": memAvail,
	}
	if gpuTotal > 0 {
		total["GPU"] = gpuTotal
		avail["GPU"] = gpuAvail
	}
	if objTotal > 0 {
		total["object_store_memory"] = objTotal
		avail["object_store_memory"] = objAvail
	}
	return ray.Node{
		State:              state,
		ResourcesTotal:     total,
		ResourcesAvailable: avail,
	}
}

func makeJob(submissionID, status string, startTime, endTime int64) ray.JobDetail {
	return ray.JobDetail{
		SubmissionID: submissionID,
		Status:       status,
		StartTime:    startTime,
		EndTime:      endTime,
	}
}

func TestDashboardView_HealthyCluster(t *testing.T) {
	nodes := []ray.Node{
		makeNode("ALIVE", 8, 1, 16*1073741824, 0),
		makeNode("ALIVE", 8, 1, 16*1073741824, 0),
		makeNode("ALIVE", 8, 1, 16*1073741824, 0),
	}

	now := time.Now().UnixMilli()
	jobs := []ray.JobDetail{
		makeJob("job-1", "RUNNING", now-60000, 0),
		makeJob("job-2", "RUNNING", now-120000, 0),
		makeJob("job-3", "PENDING", now-30000, 0),
		makeJob("job-4", "FAILED", now-300000, now-200000),
		makeJob("job-5", "SUCCEEDED", now-600000, now-500000),
	}

	m := testModel(nodes, jobs)
	view := m.View()

	assert.Contains(t, view, "Cluster Overview")
	assert.Contains(t, view, "3/3")
	assert.Contains(t, view, "24")
	assert.Contains(t, view, "3")
	assert.Contains(t, view, "48.0 GiB")
	assert.Contains(t, view, "healthy")
	assert.Contains(t, view, "2 running")
	assert.Contains(t, view, "1 pending")
	assert.Contains(t, view, "1 failed")
}

func TestDashboardView_DegradedCluster(t *testing.T) {
	nodes := []ray.Node{
		makeNode("ALIVE", 8, 1, 16*1073741824, 0),
		makeNode("ALIVE", 8, 1, 16*1073741824, 0),
		makeNode("DEAD", 8, 1, 16*1073741824, 0),
	}

	m := testModel(nodes, nil)
	view := m.View()

	assert.Contains(t, view, "2/3")
	assert.Contains(t, view, "degraded")
}

func TestDashboardView_NoData(t *testing.T) {
	m := testModel(nil, nil)
	view := m.View()

	assert.Contains(t, view, "Waiting for data")
	assert.Contains(t, view, "No recent jobs")
}

func TestDashboardView_NoGPU(t *testing.T) {
	nodes := []ray.Node{
		makeNode("ALIVE", 4, 0, 8*1073741824, 0),
		makeNode("ALIVE", 4, 0, 8*1073741824, 0),
	}

	m := testModel(nodes, nil)
	view := m.View()

	assert.Contains(t, view, "N/A")
}

func TestDashboardView_DiskBar(t *testing.T) {
	t.Run("with object store memory", func(t *testing.T) {
		nodes := []ray.Node{
			makeNode("ALIVE", 4, 0, 8*1073741824, 4*1073741824),
			makeNode("ALIVE", 4, 0, 8*1073741824, 4*1073741824),
		}

		m := testModel(nodes, nil)
		view := m.View()

		assert.Contains(t, view, "DISK")
		assert.Contains(t, view, "8.0 GiB")
	})

	t.Run("without object store memory", func(t *testing.T) {
		nodes := []ray.Node{
			makeNode("ALIVE", 4, 0, 8*1073741824, 0),
		}

		m := testModel(nodes, nil)
		view := m.View()

		assert.Contains(t, view, "DISK")
		assert.Contains(t, view, "N/A")
	})
}

func TestDashboardView_RecentJobs(t *testing.T) {
	now := time.Now().UnixMilli()

	jobs := []ray.JobDetail{
		makeJob("oldest-job-01", "SUCCEEDED", now-700000, now-600000),
		makeJob("old-job-0002", "SUCCEEDED", now-500000, now-400000),
		makeJob("job-recent-3", "RUNNING", now-60000, 0),
		makeJob("job-recent-4", "PENDING", now-50000, 0),
		makeJob("job-recent-5", "FAILED", now-300000, now-200000),
		makeJob("job-recent-6", "RUNNING", now-40000, 0),
		makeJob("job-recent-7", "SUCCEEDED", now-100000, now-80000),
	}

	nodes := []ray.Node{makeNode("ALIVE", 4, 0, 8*1073741824, 0)}
	m := testModel(nodes, jobs)
	view := m.View()

	assert.Contains(t, view, "Recent Jobs")

	// The 5 most recent jobs should appear (by timestamp).
	// Most recent first: job-recent-6 (40s), job-recent-4 (50s), job-recent-3 (60s),
	// job-recent-7 (80s end), job-recent-5 (200s end)
	assert.Contains(t, view, "job-recent-")

	// The 2 oldest should NOT appear.
	assert.NotContains(t, view, "oldest-job-")
	assert.NotContains(t, view, "old-job-000")
}

func TestDashboardView_RecentJobs_Empty(t *testing.T) {
	nodes := []ray.Node{makeNode("ALIVE", 4, 0, 8*1073741824, 0)}
	m := testModel(nodes, nil)
	view := m.View()

	assert.Contains(t, view, "No recent jobs")
}

func TestDashboardView_ErrorState(t *testing.T) {
	m := testModel(nil, nil)
	m.lastErr = errors.New("connection refused")
	view := m.View()

	assert.Contains(t, view, "connection refused")
}

func TestDashboardView_NarrowTerminal(t *testing.T) {
	nodes := []ray.Node{
		makeNode("ALIVE", 8, 1, 16*1073741824, 4*1073741824),
	}
	now := time.Now().UnixMilli()
	jobs := []ray.JobDetail{
		makeJob("job-1", "RUNNING", now-60000, 0),
	}

	m := testModel(nodes, jobs)
	m.width = 40
	view := m.View()

	assert.NotEmpty(t, view)
	// Verify no panic occurred and output was produced.
	assert.Contains(t, view, "Cluster Overview")
}

func TestDashboardSetSize(t *testing.T) {
	m := testModel(nil, nil)
	m.SetSize(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestDashboardView_JobStatusColors(t *testing.T) {
	now := time.Now().UnixMilli()
	jobs := []ray.JobDetail{
		makeJob("running-job1", "RUNNING", now-10000, 0),
		makeJob("pending-job1", "PENDING", now-20000, 0),
		makeJob("failed-job01", "FAILED", now-30000, now-25000),
		makeJob("success-job1", "SUCCEEDED", now-40000, now-35000),
		makeJob("stopped-job1", "STOPPED", now-50000, now-45000),
	}

	nodes := []ray.Node{makeNode("ALIVE", 4, 0, 8*1073741824, 0)}
	m := testModel(nodes, jobs)
	view := m.View()

	// Status strings appear in recent jobs
	assert.Contains(t, view, "RUNNING")
	assert.Contains(t, view, "PENDING")
	assert.Contains(t, view, "FAILED")
	assert.Contains(t, view, "SUCCEEDED")
	assert.Contains(t, view, "STOPPED")
}

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		ms       int64
		contains string
	}{
		{"zero", 0, "unknown"},
		{"seconds ago", time.Now().Add(-30 * time.Second).UnixMilli(), "s ago"},
		{"minutes ago", time.Now().Add(-5 * time.Minute).UnixMilli(), "m ago"},
		{"hours ago", time.Now().Add(-3 * time.Hour).UnixMilli(), "h ago"},
		{"days ago", time.Now().Add(-48 * time.Hour).UnixMilli(), "d ago"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatRelativeTime(tc.ms)
			assert.Contains(t, result, tc.contains)
		})
	}
}

func TestColorizeStatus(t *testing.T) {
	// Verify each status produces non-empty styled output.
	statuses := []string{"RUNNING", "PENDING", "FAILED", "SUCCEEDED", "STOPPED", "UNKNOWN"}
	for _, s := range statuses {
		result := colorizeStatus(s)
		assert.Contains(t, result, s)
	}
}

func TestJobTimestamp(t *testing.T) {
	t.Run("prefers EndTime", func(t *testing.T) {
		j := ray.JobDetail{StartTime: 100, EndTime: 200}
		assert.Equal(t, int64(200), jobTimestamp(j))
	})
	t.Run("falls back to StartTime", func(t *testing.T) {
		j := ray.JobDetail{StartTime: 100, EndTime: 0}
		assert.Equal(t, int64(100), jobTimestamp(j))
	})
	t.Run("both zero", func(t *testing.T) {
		j := ray.JobDetail{}
		assert.Equal(t, int64(0), jobTimestamp(j))
	})
}

func TestDashboardView_SubmissionIDTruncation(t *testing.T) {
	now := time.Now().UnixMilli()
	longID := "very-long-submission-id-that-exceeds-12-chars"
	jobs := []ray.JobDetail{
		makeJob(longID, "RUNNING", now-10000, 0),
	}

	nodes := []ray.Node{makeNode("ALIVE", 4, 0, 8*1073741824, 0)}
	m := testModel(nodes, jobs)
	view := m.View()

	// The full ID should NOT appear; only first 12 chars.
	assert.NotContains(t, view, longID)
	assert.Contains(t, view, longID[:12])
}

func TestDashboardView_ResourceCapacitySection_NoAvailable(t *testing.T) {
	// Without resources_available, bars show 0% with total only.
	nodes := []ray.Node{
		makeNode("ALIVE", 8, 2, 16*1073741824, 4*1073741824),
	}

	m := testModel(nodes, nil)
	view := m.View()

	assert.Contains(t, view, "Resource Utilization")
	assert.Contains(t, view, "CPU")
	assert.Contains(t, view, "8 cores")
	assert.Contains(t, view, "GPU")
	assert.Contains(t, view, "2 units")
	assert.Contains(t, view, "MEM")
	assert.Contains(t, view, "16.0 GiB")
	assert.Contains(t, view, "DISK")
	assert.Contains(t, view, "4.0 GiB")
}

func TestDashboardView_ResourceUtilization_WithAvailable(t *testing.T) {
	// With resources_available, bars show used/total ratios.
	nodes := []ray.Node{
		// Node 1: 8 CPU total, 2 available (6 used), 2 GPU total, 1 available (1 used)
		// 16 GiB memory total, 4 GiB available (12 used), 4 GiB obj store, 1 GiB available (3 used)
		makeNodeWithAvailable("ALIVE",
			8, 2, // CPU total, avail
			2, 1, // GPU total, avail
			16*1073741824, 4*1073741824, // mem total, avail
			4*1073741824, 1*1073741824, // obj store total, avail
		),
	}

	m := testModel(nodes, nil)
	view := m.View()

	assert.Contains(t, view, "Resource Utilization")
	// Should show used/total format
	assert.Contains(t, view, "6/8 cores")
	assert.Contains(t, view, "1/2 units")
	assert.Contains(t, view, "12.0/16.0 GiB")
	assert.Contains(t, view, "3.0/4.0 GiB")

	// Verify progress bars rendered (they use block characters).
	assert.True(t, strings.ContainsAny(view, "▌█░"))
}

// --- Polling and data handling tests ---

func TestDashboardNew_NoArgs(t *testing.T) {
	m := New("")
	assert.Equal(t, 0, m.width)
	assert.Equal(t, 0, m.height)
	assert.Nil(t, m.lastErr)
}

func TestDashboardInit_ReturnsSpinnerTick(t *testing.T) {
	m := New("")
	cmd := m.Init()
	assert.NotNil(t, cmd, "Init() should return spinner tick command")
}

func TestDashboardUpdate_ClusterDataMsg_Success(t *testing.T) {
	m := New("")
	m.refreshing = true

	now := time.Now().UnixMilli()
	msg := ClusterDataMsg{
		Nodes: []ray.Node{
			makeNode("ALIVE", 4, 0, 8*1073741824, 0),
			makeNode("ALIVE", 8, 1, 16*1073741824, 0),
		},
		Jobs: []ray.JobDetail{
			makeJob("j1", "RUNNING", now-10000, 0),
			makeJob("j2", "FAILED", now-20000, now-15000),
			makeJob("j3", "SUCCEEDED", now-30000, now-25000),
		},
		FetchErr: nil,
		Latency:  50 * time.Millisecond,
	}

	updated, cmd := m.Update(msg)

	assert.False(t, updated.refreshing, "refreshing should be false after data received")
	assert.Nil(t, cmd, "no follow-up command needed")
	assert.Equal(t, 2, updated.health.NodeCount)
	assert.Equal(t, 2, updated.health.AliveNodes)
	assert.Equal(t, "healthy", updated.health.Status)
	assert.Equal(t, 1, updated.jobs.Running)
	assert.Equal(t, 1, updated.jobs.Failed)
	assert.Equal(t, 50*time.Millisecond, updated.lastLatency)
	assert.Nil(t, updated.lastErr)
	assert.WithinDuration(t, time.Now(), updated.lastUpdate, time.Second)
}

func TestDashboardUpdate_ClusterDataMsg_Error(t *testing.T) {
	m := New("")
	m.refreshing = true
	// Set some previous data that should be preserved on error.
	m.health = ClusterHealth{NodeCount: 3, AliveNodes: 3, Status: "healthy"}
	m.jobs = JobSummary{Running: 2, Total: 5}

	msg := ClusterDataMsg{
		FetchErr: fmt.Errorf("connection refused"),
		Latency:  100 * time.Millisecond,
	}

	updated, _ := m.Update(msg)

	assert.False(t, updated.refreshing)
	require.NotNil(t, updated.lastErr)
	assert.Contains(t, updated.lastErr.Error(), "connection refused")
	// Previous health/jobs data should be preserved.
	assert.Equal(t, 3, updated.health.NodeCount)
	assert.Equal(t, "healthy", updated.health.Status)
	assert.Equal(t, 2, updated.jobs.Running)
	assert.Equal(t, 100*time.Millisecond, updated.lastLatency)
	// lastUpdate should NOT be updated on error.
	assert.True(t, updated.lastUpdate.IsZero(), "lastUpdate should not be set on error")
}

// Note: Status bar tests moved to app package (status bar is now global, rendered by root model).

func TestDashboardView_IncludesHeatmap(t *testing.T) {
	nodes := []ray.Node{
		{
			State:              "ALIVE",
			NodeIP:             "10.0.0.1",
			IsHeadNode:         true,
			ResourcesTotal:     map[string]float64{"CPU": 4, "memory": 8 * 1073741824},
			ResourcesAvailable: map[string]float64{"CPU": 2, "memory": 4 * 1073741824},
		},
		{
			State:              "ALIVE",
			NodeIP:             "10.0.0.2",
			IsHeadNode:         false,
			ResourcesTotal:     map[string]float64{"CPU": 8, "memory": 16 * 1073741824},
			ResourcesAvailable: map[string]float64{"CPU": 1, "memory": 2 * 1073741824},
		},
	}

	m := testModel(nodes, nil)
	view := m.View()

	assert.Contains(t, view, "Usage")
}

func TestDashboardUpdate_HeatmapCycle(t *testing.T) {
	m := New("")
	assert.Equal(t, HeatCPU, m.heatmapResource)

	// h -> HeatMemory
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "h"})
	assert.Equal(t, HeatMemory, m.heatmapResource)

	// h -> HeatGPU
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "h"})
	assert.Equal(t, HeatGPU, m.heatmapResource)

	// h -> back to HeatCPU
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "h"})
	assert.Equal(t, HeatCPU, m.heatmapResource)
}

func TestDonutChart_Empty(t *testing.T) {
	result := renderDonutChart(JobSummary{}, ClusterHealth{})
	assert.Equal(t, "", result)
}

func TestDonutChart_WithJobs(t *testing.T) {
	jobs := JobSummary{Running: 5, Pending: 2, Failed: 1, Succeeded: 2, Total: 10}
	result := renderDonutChart(jobs, ClusterHealth{NodeCount: 2, AliveNodes: 2, Status: "healthy"})
	assert.NotEmpty(t, result)
	// Should contain node health in the center.
	assert.Contains(t, result, "2/2N")
	// Should contain legend dots.
	assert.Contains(t, result, "●")
}

func TestDonutChart_SingleStatus(t *testing.T) {
	jobs := JobSummary{Running: 3, Total: 3}
	result := renderDonutChart(jobs, ClusterHealth{NodeCount: 2, AliveNodes: 2, Status: "healthy"})
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "2/2N")
}

func TestDashboardView_WithDonut(t *testing.T) {
	nodes := []ray.Node{
		makeNode("ALIVE", 4, 0, 8*1073741824, 0),
		makeNode("ALIVE", 4, 0, 8*1073741824, 0),
	}
	jobs := []ray.JobDetail{
		{Status: "RUNNING", StartTime: time.Now().Add(-5 * time.Minute).UnixMilli()},
		{Status: "SUCCEEDED", StartTime: time.Now().Add(-10 * time.Minute).UnixMilli(), EndTime: time.Now().Add(-8 * time.Minute).UnixMilli()},
	}
	m := testModel(nodes, jobs)
	view := m.View()

	// Donut chart should be rendered alongside health section.
	assert.Contains(t, view, "Cluster Health")
	// Job status bar should be present.
	assert.Contains(t, view, "Job Status")
}
