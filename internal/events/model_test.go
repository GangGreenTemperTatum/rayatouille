package events

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func testEvents() []ray.ClusterEvent {
	return []ray.ClusterEvent{
		{
			Severity:   "INFO",
			Time:       "2024-01-15T10:00:00Z",
			SourceType: "GCS",
			Message:    "Node added to cluster",
			EventID:    "evt-001",
		},
		{
			Severity:   "INFO",
			Time:       "2024-01-15T10:05:00Z",
			SourceType: "RAYLET",
			Message:    "Worker process started",
			EventID:    "evt-002",
		},
		{
			Severity:   "WARNING",
			Time:       "2024-01-15T10:10:00Z",
			SourceType: "AUTOSCALER",
			Message:    "Scaling up: insufficient resources",
			EventID:    "evt-003",
		},
		{
			Severity:   "ERROR",
			Time:       "2024-01-15T10:15:00Z",
			SourceType: "GCS",
			Message:    "Node failed heartbeat check",
			EventID:    "evt-004",
		},
	}
}

func TestNew_InitialState(t *testing.T) {
	m := New()
	assert.Equal(t, SortByTime, m.sortField)
	assert.Equal(t, SortDesc, m.sortOrder)
	assert.Equal(t, SeverityAll, m.severityFilter)
	assert.False(t, m.ready)
	assert.Empty(t, m.allEvents)
	assert.Empty(t, m.filteredEvents)
}

func TestSetEvents_PopulatesTable(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	assert.Len(t, m.filteredEvents, 4)
	assert.Len(t, m.allEvents, 4)
}

func TestSeverityFilter_Error(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	m.severityFilter = SeverityError
	m.applyFilters()

	assert.Len(t, m.filteredEvents, 1)
	assert.Equal(t, "ERROR", m.filteredEvents[0].Severity)
}

func TestSeverityFilter_Warning(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	m.severityFilter = SeverityWarning
	m.applyFilters()

	assert.Len(t, m.filteredEvents, 1)
	assert.Equal(t, "WARNING", m.filteredEvents[0].Severity)
}

func TestSeverityFilter_Info(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	m.severityFilter = SeverityInfo
	m.applyFilters()

	assert.Len(t, m.filteredEvents, 2)
	for _, e := range m.filteredEvents {
		assert.Equal(t, "INFO", e.Severity)
	}
}

func TestTextFilter_BySourceType(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	m.filter.SetValueForTest("GCS")
	m.applyFilters()

	assert.Len(t, m.filteredEvents, 2)
	for _, e := range m.filteredEvents {
		assert.Equal(t, "GCS", e.SourceType)
	}
}

func TestTextFilter_ByMessage(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	m.filter.SetValueForTest("heartbeat")
	m.applyFilters()

	assert.Len(t, m.filteredEvents, 1)
	assert.Equal(t, "evt-004", m.filteredEvents[0].EventID)
}

func TestSortByTime_NewestFirst(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	// Default sort is SortByTime, SortDesc (newest first).
	assert.Len(t, m.filteredEvents, 4)
	assert.Equal(t, "evt-004", m.filteredEvents[0].EventID) // latest
	assert.Equal(t, "evt-001", m.filteredEvents[3].EventID) // earliest
}

func TestSortToggle(t *testing.T) {
	m := New()
	m.SetEvents(testEvents())

	assert.Equal(t, SortByTime, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortBySeverity, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortBySource, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByTime, m.sortField)
}

func TestFilterActive(t *testing.T) {
	m := New()
	assert.False(t, m.FilterActive())

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	assert.True(t, m.FilterActive())
}

func TestEventToRow(t *testing.T) {
	e := testEvents()[0]
	row := EventToRow(e)

	assert.Equal(t, "2024-01-15 10:00:00", row[0]) // parsed ISO timestamp
	assert.Equal(t, "INFO", row[1])
	assert.Equal(t, "GCS", row[2])
	assert.Equal(t, "Node added to cluster", row[3])
}

func TestColumns_WidthDistribution(t *testing.T) {
	cols := Columns(120)
	assert.Len(t, cols, 4)

	totalWidth := 0
	for _, c := range cols {
		totalWidth += c.Width
	}
	assert.LessOrEqual(t, totalWidth, 120)
}

func TestEmptyEvents_View(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	view := m.View()
	assert.Contains(t, view, "No cluster events")
}
