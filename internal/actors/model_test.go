package actors

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func testActors() []ray.Actor {
	return []ray.Actor{
		{
			ActorID:   "abc123def456789",
			ClassName: "MyTrainer",
			Name:      "trainer-0",
			State:     "ALIVE",
			PID:       12345,
			JobID:     "01000000",
			NodeID:    "node1",
		},
		{
			ActorID:   "def456ghi789012",
			ClassName: "DataWorker",
			Name:      "",
			State:     "ALIVE",
			PID:       23456,
			JobID:     "01000000",
			NodeID:    "node2",
		},
		{
			ActorID:   "ghi789jkl012345",
			ClassName: "CheckpointSaver",
			Name:      "saver",
			State:     "DEAD",
			PID:       34567,
			JobID:     "02000000",
			NodeID:    "node1",
		},
		{
			ActorID:   "jkl012mno345678",
			ClassName: "Scheduler",
			Name:      "scheduler-main",
			State:     "PENDING_CREATION",
			PID:       0,
			JobID:     "01000000",
			NodeID:    "",
		},
	}
}

func TestNew_InitialState(t *testing.T) {
	m := New()
	assert.Equal(t, SortByClass, m.sortField)
	assert.Equal(t, SortAsc, m.sortOrder)
	assert.Equal(t, StatusAll, m.statusFilter)
	assert.False(t, m.ready)
	assert.Empty(t, m.allActors)
	assert.Empty(t, m.filteredActors)
}

func TestSetActors_PopulatesTable(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	assert.Len(t, m.filteredActors, 4)
	assert.Len(t, m.allActors, 4)
}

func TestStatusFilter_Alive(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	m.statusFilter = StatusAlive
	m.applyFilters()

	assert.Len(t, m.filteredActors, 2)
	for _, a := range m.filteredActors {
		assert.Equal(t, "ALIVE", a.State)
	}
}

func TestStatusFilter_Dead(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	m.statusFilter = StatusDead
	m.applyFilters()

	assert.Len(t, m.filteredActors, 1)
	assert.Equal(t, "DEAD", m.filteredActors[0].State)
}

func TestStatusFilter_Pending(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	m.statusFilter = StatusPending
	m.applyFilters()

	assert.Len(t, m.filteredActors, 1)
	assert.Equal(t, "PENDING_CREATION", m.filteredActors[0].State)
}

func TestTextFilter_ByClassName(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	m.filter.SetValueForTest("trainer")
	m.applyFilters()

	assert.Len(t, m.filteredActors, 1)
	assert.Equal(t, "MyTrainer", m.filteredActors[0].ClassName)
}

func TestTextFilter_ByActorID(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	m.filter.SetValueForTest("abc123")
	m.applyFilters()

	assert.Len(t, m.filteredActors, 1)
	assert.Equal(t, "abc123def456789", m.filteredActors[0].ActorID)
}

func TestTextFilter_ByName(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	m.filter.SetValueForTest("saver")
	m.applyFilters()

	assert.Len(t, m.filteredActors, 1)
	assert.Equal(t, "saver", m.filteredActors[0].Name)
}

func TestSelectedActor_ReturnsCorrectActor(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	selected := m.SelectedActor()
	require.NotNil(t, selected)
	// Should be the first actor after sorting by Class (ascending).
	// CheckpointSaver < DataWorker < MyTrainer < Scheduler
	assert.Equal(t, "CheckpointSaver", selected.ClassName)
}

func TestSelectedActor_EmptyList(t *testing.T) {
	m := New()
	selected := m.SelectedActor()
	assert.Nil(t, selected)
}

func TestSortToggle(t *testing.T) {
	m := New()
	m.SetActors(testActors())

	assert.Equal(t, SortByClass, m.sortField)

	// Press 's' to cycle sort field.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByState, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByPID, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByJobID, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByClass, m.sortField)
}

func TestFilterActive(t *testing.T) {
	m := New()
	assert.False(t, m.FilterActive())

	// Activate filter via key press.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	assert.True(t, m.FilterActive())
}

func TestActorToRow_WithName(t *testing.T) {
	actors := testActors()
	row := ActorToRow(actors[0]) // MyTrainer, trainer-0

	assert.Equal(t, "abc123def456..", row[0]) // Truncated actor ID
	assert.Equal(t, "ALIVE", row[1])
	assert.Equal(t, "MyTrainer", row[2])
	assert.Equal(t, "trainer-0", row[3])
	assert.Equal(t, "12345", row[4])
	assert.Equal(t, "01000000", row[5])
}

func TestActorToRow_EmptyName(t *testing.T) {
	actors := testActors()
	row := ActorToRow(actors[1]) // DataWorker, no name

	assert.Equal(t, "-", row[3]) // Name defaults to "-"
}

func TestActorToRow_ZeroPID(t *testing.T) {
	actors := testActors()
	row := ActorToRow(actors[3]) // Scheduler, PID 0

	assert.Equal(t, "-", row[4]) // PID shows "-" for 0
}

func TestColumns_WidthDistribution(t *testing.T) {
	cols := Columns(120)
	assert.Len(t, cols, 6)

	totalWidth := 0
	for _, c := range cols {
		totalWidth += c.Width
	}
	assert.LessOrEqual(t, totalWidth, 120)
}
