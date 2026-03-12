package nodes

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func testNodes() []ray.Node {
	return []ray.Node{
		{
			NodeID:     "head123456789",
			NodeName:   "head-node",
			NodeIP:     "10.0.0.1",
			State:      "ALIVE",
			IsHeadNode: true,
			ResourcesTotal: map[string]float64{
				"CPU":                 0,
				"GPU":                 0,
				"memory":              17179869184, // 16 GiB
				"object_store_memory": 8589934592,  // 8 GiB
			},
		},
		{
			NodeID:     "worker1abc",
			NodeName:   "worker-1",
			NodeIP:     "10.0.0.2",
			State:      "ALIVE",
			IsHeadNode: false,
			ResourcesTotal: map[string]float64{
				"CPU":                 4,
				"GPU":                 1,
				"memory":              17179869184, // 16 GiB
				"object_store_memory": 8589934592,  // 8 GiB
			},
		},
		{
			NodeID:     "worker2def",
			NodeName:   "worker-2",
			NodeIP:     "10.0.0.3",
			State:      "DEAD",
			IsHeadNode: false,
			ResourcesTotal: map[string]float64{
				"CPU":                 4,
				"GPU":                 0,
				"memory":              17179869184, // 16 GiB
				"object_store_memory": 8589934592,  // 8 GiB
			},
		},
	}
}

func TestNew_InitialState(t *testing.T) {
	m := New()
	assert.Equal(t, SortByIP, m.sortField)
	assert.Equal(t, SortAsc, m.sortOrder)
	assert.Equal(t, StatusAll, m.statusFilter)
	assert.False(t, m.ready)
	assert.Empty(t, m.allNodes)
	assert.Empty(t, m.filteredNodes)
}

func TestSetNodes_PopulatesTable(t *testing.T) {
	m := New()
	nodes := testNodes()
	m.SetNodes(nodes)

	assert.Len(t, m.filteredNodes, 3)
	assert.Len(t, m.allNodes, 3)
}

func TestSetNodes_AppliesFilters(t *testing.T) {
	m := New()
	m.statusFilter = StatusAlive
	m.SetNodes(testNodes())

	assert.Len(t, m.filteredNodes, 2)
}

func TestStatusFilter_Alive(t *testing.T) {
	m := New()
	m.SetNodes(testNodes())

	m.statusFilter = StatusAlive
	m.applyFilters()

	assert.Len(t, m.filteredNodes, 2)
	for _, n := range m.filteredNodes {
		assert.Equal(t, "ALIVE", n.State)
	}
}

func TestStatusFilter_Dead(t *testing.T) {
	m := New()
	m.SetNodes(testNodes())

	m.statusFilter = StatusDead
	m.applyFilters()

	assert.Len(t, m.filteredNodes, 1)
	assert.Equal(t, "DEAD", m.filteredNodes[0].State)
}

func TestTextFilter_ByIP(t *testing.T) {
	m := New()
	m.SetNodes(testNodes())

	m.filter.SetValueForTest("10.0.0.2")
	m.applyFilters()

	assert.Len(t, m.filteredNodes, 1)
	assert.Equal(t, "10.0.0.2", m.filteredNodes[0].NodeIP)
}

func TestSortToggle(t *testing.T) {
	m := New()
	m.SetNodes(testNodes())

	assert.Equal(t, SortByIP, m.sortField)

	// Press 's' to cycle sort field.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByStatus, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByCPU, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByMemory, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByIP, m.sortField)
}

func TestSelectedNode_ReturnsCorrectNode(t *testing.T) {
	m := New()
	m.SetNodes(testNodes())

	selected := m.SelectedNode()
	require.NotNil(t, selected)
	// Should be the first node after sorting by IP (ascending).
	assert.Equal(t, "10.0.0.1", selected.NodeIP)
}

func TestSelectedNode_EmptyList(t *testing.T) {
	m := New()
	selected := m.SelectedNode()
	assert.Nil(t, selected)
}

func TestNodeToRow_WorkerNode(t *testing.T) {
	nodes := testNodes()
	worker := nodes[1] // worker1abc
	row := NodeToRow(worker)

	assert.Equal(t, "worker1abc", row[0]) // Node ID (short enough, no truncation)
	assert.Equal(t, "ALIVE", row[1])      // Status
	assert.Equal(t, "worker", row[2])     // Role
	assert.Equal(t, "10.0.0.2", row[3])   // IP
	assert.Equal(t, "4", row[4])          // CPU
	assert.Equal(t, "1", row[5])          // GPU
	assert.Equal(t, "16.0 GiB", row[6])   // Memory
	assert.Equal(t, "8.0 GiB", row[7])    // ObjStore
}

func TestNodeToRow_HeadNode(t *testing.T) {
	nodes := testNodes()
	head := nodes[0] // head node
	row := NodeToRow(head)

	assert.Equal(t, "head123456..", row[0]) // Truncated
	assert.Equal(t, "ALIVE", row[1])
	assert.Equal(t, "head", row[2]) // Role
	assert.Equal(t, "0", row[4])    // CPU shows 0
	assert.Equal(t, "-", row[5])    // GPU shows -
}

func TestNodeToRow_LongNodeID(t *testing.T) {
	n := ray.Node{
		NodeID:         "abcdefghijklmnopqrstuvwxyz",
		State:          "ALIVE",
		ResourcesTotal: map[string]float64{},
	}
	row := NodeToRow(n)
	assert.Equal(t, "abcdefghij..", row[0])
}

func TestColumns_WidthDistribution(t *testing.T) {
	cols := Columns(120)
	assert.Len(t, cols, 8)

	totalWidth := 0
	for _, c := range cols {
		totalWidth += c.Width
	}
	// Total column widths should be less than the available width
	// (padding takes some space).
	assert.LessOrEqual(t, totalWidth, 120)
}

func TestFilterActive(t *testing.T) {
	m := New()
	assert.False(t, m.FilterActive())

	// Activate filter via key press.
	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	assert.True(t, m.FilterActive())
}
