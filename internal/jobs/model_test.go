package jobs

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func mockJobs() []ray.JobDetail {
	now := time.Now()
	return []ray.JobDetail{
		{
			SubmissionID: "raysubmit_001",
			Status:       "RUNNING",
			Entrypoint:   "python train.py",
			StartTime:    now.Add(-10 * time.Minute).UnixMilli(),
			EndTime:      0,
		},
		{
			SubmissionID: "raysubmit_002",
			Status:       "FAILED",
			Entrypoint:   "python evaluate.py --model=bert",
			StartTime:    now.Add(-30 * time.Minute).UnixMilli(),
			EndTime:      now.Add(-25 * time.Minute).UnixMilli(),
		},
		{
			SubmissionID: "raysubmit_003",
			Status:       "PENDING",
			Entrypoint:   "python preprocess.py",
			StartTime:    now.Add(-5 * time.Minute).UnixMilli(),
			EndTime:      0,
		},
		{
			SubmissionID: "raysubmit_004_very_long_id",
			Status:       "RUNNING",
			Entrypoint:   "python serve.py --port=8080",
			StartTime:    now.Add(-60 * time.Minute).UnixMilli(),
			EndTime:      0,
		},
	}
}

func TestNew_InitialState(t *testing.T) {
	m := New()
	assert.Equal(t, StatusRunning, m.statusFilter)
	assert.Equal(t, SortByAge, m.sortField)
	assert.False(t, m.ready)
	assert.Nil(t, m.allJobs)
	assert.Nil(t, m.filteredJobs)
}

func TestSetJobs_PopulatesFilteredJobs(t *testing.T) {
	m := New()
	m.statusFilter = StatusAll
	jobs := mockJobs()
	m.SetJobs(jobs)

	assert.Equal(t, len(jobs), len(m.allJobs))
	assert.Equal(t, len(jobs), len(m.filteredJobs))
	assert.Equal(t, len(jobs), len(m.table.Rows()))
}

func TestStatusFilter_Running(t *testing.T) {
	m := New()
	m.SetJobs(mockJobs())

	m.statusFilter = StatusRunning
	m.applyFilters()

	assert.Equal(t, 2, len(m.filteredJobs))
	for _, j := range m.filteredJobs {
		assert.Equal(t, "RUNNING", j.Status)
	}
}

func TestStatusFilter_Failed(t *testing.T) {
	m := New()
	m.SetJobs(mockJobs())

	m.statusFilter = StatusFailed
	m.applyFilters()

	assert.Equal(t, 1, len(m.filteredJobs))
	assert.Equal(t, "FAILED", m.filteredJobs[0].Status)
}

func TestStatusFilter_Pending(t *testing.T) {
	m := New()
	m.SetJobs(mockJobs())

	m.statusFilter = StatusPending
	m.applyFilters()

	assert.Equal(t, 1, len(m.filteredJobs))
	assert.Equal(t, "PENDING", m.filteredJobs[0].Status)
}

func TestTextFilter_MatchesSubmissionID(t *testing.T) {
	m := New()
	m.statusFilter = StatusAll
	m.SetJobs(mockJobs())
	m.filter.SetValueForTest("003")
	m.applyFilters()

	assert.Equal(t, 1, len(m.filteredJobs))
	assert.Equal(t, "raysubmit_003", m.filteredJobs[0].SubmissionID)
}

func TestTextFilter_MatchesEntrypoint(t *testing.T) {
	m := New()
	m.statusFilter = StatusAll
	m.SetJobs(mockJobs())
	m.filter.SetValueForTest("train")
	m.applyFilters()

	assert.Equal(t, 1, len(m.filteredJobs))
	assert.Equal(t, "python train.py", m.filteredJobs[0].Entrypoint)
}

func TestCombinedStatusAndTextFilter(t *testing.T) {
	m := New()
	m.SetJobs(mockJobs())
	m.statusFilter = StatusRunning
	m.filter.SetValueForTest("serve")
	m.applyFilters()

	assert.Equal(t, 1, len(m.filteredJobs))
	assert.Equal(t, "RUNNING", m.filteredJobs[0].Status)
	assert.Contains(t, m.filteredJobs[0].Entrypoint, "serve")
}

func TestSortByAge_DefaultDescending(t *testing.T) {
	m := New()
	m.SetJobs(mockJobs())

	// Default: SortByAge + SortDesc = most recent first (highest StartTime first).
	require.True(t, len(m.filteredJobs) >= 2)
	assert.True(t, m.filteredJobs[0].StartTime >= m.filteredJobs[1].StartTime,
		"expected most recent first")
}

func TestSortCycling(t *testing.T) {
	m := New()
	assert.Equal(t, SortByAge, m.sortField)

	// Simulate pressing 's' four times to cycle through all fields.
	for _, expected := range []SortField{SortByStatus, SortByEntrypoint, SortByDuration, SortByAge} {
		m, _ = m.Update(tea.KeyPressMsg{Code: 's'})
		assert.Equal(t, expected, m.sortField)
	}
}

func TestCursorPreservation(t *testing.T) {
	m := New()
	m.statusFilter = StatusAll
	m.SetSize(80, 40)
	m.SetJobs(mockJobs())

	// Move cursor to row 2.
	m.table.SetCursor(2)
	assert.Equal(t, 2, m.table.Cursor())

	// Re-set jobs with same data -- cursor should be preserved.
	m.SetJobs(mockJobs())
	assert.Equal(t, 2, m.table.Cursor())
}

func TestCursorClampedOnSmallerList(t *testing.T) {
	m := New()
	m.statusFilter = StatusAll
	m.SetSize(80, 40)
	m.SetJobs(mockJobs())

	// Move cursor to last row.
	m.table.SetCursor(3)
	assert.Equal(t, 3, m.table.Cursor())

	// Filter to fewer jobs.
	m.statusFilter = StatusPending
	m.applyFilters()

	// Cursor should be clamped to valid range.
	assert.True(t, m.table.Cursor() < len(m.filteredJobs))
}

func TestSelectedJob(t *testing.T) {
	m := New()
	m.SetJobs(mockJobs())

	j := m.SelectedJob()
	require.NotNil(t, j)
	// Should return the job at cursor position 0.
	assert.Equal(t, m.filteredJobs[0].SubmissionID, j.SubmissionID)
}

func TestSelectedJob_EmptyList(t *testing.T) {
	m := New()
	j := m.SelectedJob()
	assert.Nil(t, j)
}

func TestJobToRow_BasicConversion(t *testing.T) {
	now := time.Now()
	j := ray.JobDetail{
		SubmissionID: "raysubmit_abc",
		Status:       "RUNNING",
		Entrypoint:   "python main.py",
		StartTime:    now.Add(-5 * time.Minute).UnixMilli(),
		EndTime:      0,
	}

	row := JobToRow(j)
	// "raysubmit_abc" is 13 chars > 12, so it gets truncated.
	assert.Equal(t, "raysubmit_ab..", row[0])
	assert.Equal(t, "RUNNING", row[1])
	assert.Equal(t, "python main.py", row[2])
	assert.NotEmpty(t, row[3]) // duration
	assert.Contains(t, row[4], "ago")
}

func TestJobToRow_LongSubmissionID(t *testing.T) {
	j := ray.JobDetail{
		SubmissionID: "raysubmit_very_long_id_here",
		Status:       "PENDING",
		Entrypoint:   "python train.py",
		StartTime:    time.Now().Add(-1 * time.Minute).UnixMilli(),
	}

	row := JobToRow(j)
	assert.Equal(t, "raysubmit_ve..", row[0])
}

func TestJobToRow_NoDuration(t *testing.T) {
	j := ray.JobDetail{
		SubmissionID: "short",
		Status:       "PENDING",
		Entrypoint:   "python x.py",
		StartTime:    0,
		EndTime:      0,
	}

	row := JobToRow(j)
	assert.Equal(t, "-", row[3])
	assert.Equal(t, "-", row[4])
}

func TestSetSize_SetsReady(t *testing.T) {
	m := New()
	assert.False(t, m.ready)

	m.SetSize(120, 40)
	assert.True(t, m.ready)
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestView_NotReady(t *testing.T) {
	m := New()
	assert.Equal(t, "Loading...", m.View())
}

func TestView_Ready(t *testing.T) {
	m := New()
	m.SetSize(80, 40)
	m.SetJobs(mockJobs())

	view := m.View()
	assert.Contains(t, view, "Jobs")
	assert.Contains(t, view, "[RUNNING]")
	assert.Contains(t, view, "Showing 2 of 4 jobs")
	assert.Contains(t, view, "Sort: Age")
}

func TestStatusFilter_ViaKeypress(t *testing.T) {
	m := New()
	m.statusFilter = StatusAll
	m.SetJobs(mockJobs())

	// Press 'r' for running.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "r"})
	assert.Equal(t, StatusRunning, m.statusFilter)
	assert.Equal(t, 2, len(m.filteredJobs))

	// Press 'a' for all.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "a"})
	assert.Equal(t, StatusAll, m.statusFilter)
	assert.Equal(t, 4, len(m.filteredJobs))

	// Press 'f' for failed.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "f"})
	assert.Equal(t, StatusFailed, m.statusFilter)
	assert.Equal(t, 1, len(m.filteredJobs))

	// Press 'c' for completed/succeeded.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "c"})
	assert.Equal(t, StatusSucceeded, m.statusFilter)
}
