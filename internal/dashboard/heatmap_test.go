package dashboard

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// testNodesForHeatmap returns a set of nodes for heatmap testing:
// - head node: ALIVE, no CPU, has memory
// - worker1: ALIVE, CPU=4, available CPU=1 (75% used)
// - worker2: ALIVE, CPU=4, ResourcesAvailable nil (unknown)
// - worker3: DEAD, CPU=4
func testNodesForHeatmap() []ray.Node {
	head := ray.Node{
		State:      "ALIVE",
		NodeIP:     "10.0.0.1",
		IsHeadNode: true,
		ResourcesTotal: map[string]float64{
			"memory": 16 * 1073741824,
		},
		ResourcesAvailable: map[string]float64{
			"memory": 8 * 1073741824,
		},
	}

	worker1 := ray.Node{
		State:      "ALIVE",
		NodeIP:     "10.0.0.101",
		IsHeadNode: false,
		ResourcesTotal: map[string]float64{
			"CPU":    4,
			"memory": 16 * 1073741824,
		},
		ResourcesAvailable: map[string]float64{
			"CPU":    1,
			"memory": 4 * 1073741824,
		},
	}

	worker2 := ray.Node{
		State:      "ALIVE",
		NodeIP:     "10.0.0.102",
		IsHeadNode: false,
		ResourcesTotal: map[string]float64{
			"CPU":    4,
			"memory": 16 * 1073741824,
		},
		ResourcesAvailable: nil,
	}

	worker3 := ray.Node{
		State:      "DEAD",
		NodeIP:     "10.0.0.103",
		IsHeadNode: false,
		ResourcesTotal: map[string]float64{
			"CPU":    4,
			"memory": 16 * 1073741824,
		},
		ResourcesAvailable: map[string]float64{
			"CPU":    4,
			"memory": 16 * 1073741824,
		},
	}

	return []ray.Node{head, worker1, worker2, worker3}
}

func TestHeatColor_Gradient(t *testing.T) {
	tests := []struct {
		name     string
		ratio    float64
		expected color.Color
	}{
		{"negative (N/A)", -1, lipgloss.Color("#626262")},
		{"low (10%)", 0.1, lipgloss.Color("#04B575")},
		{"moderate (40%)", 0.4, lipgloss.Color("#6FBF40")},
		{"elevated (70%)", 0.7, lipgloss.Color("#FFCC00")},
		{"high (90%)", 0.9, lipgloss.Color("#FF4444")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := heatColor(tc.ratio)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestNodeResourceRatio_CPU_WithAvailable(t *testing.T) {
	nodes := testNodesForHeatmap()
	// worker1: CPU total=4, available=1 -> used=3, ratio=0.75
	ratio := nodeResourceRatio(nodes[1], HeatCPU)
	assert.InDelta(t, 0.75, ratio, 0.001)
}

func TestNodeResourceRatio_CPU_NilAvailable(t *testing.T) {
	nodes := testNodesForHeatmap()
	// worker2: ResourcesAvailable is nil
	ratio := nodeResourceRatio(nodes[2], HeatCPU)
	assert.Equal(t, float64(-1), ratio)
}

func TestNodeResourceRatio_CPU_ZeroTotal(t *testing.T) {
	nodes := testNodesForHeatmap()
	// head node: no CPU in ResourcesTotal
	ratio := nodeResourceRatio(nodes[0], HeatCPU)
	assert.Equal(t, float64(-1), ratio)
}

func TestNodeResourceRatio_Memory(t *testing.T) {
	nodes := testNodesForHeatmap()
	// worker1: memory total=16GiB, available=4GiB -> used=12GiB, ratio=0.75
	ratio := nodeResourceRatio(nodes[1], HeatMemory)
	assert.InDelta(t, 0.75, ratio, 0.001)
}

func TestNodeResourceRatio_GPU_Missing(t *testing.T) {
	nodes := testNodesForHeatmap()
	// No GPU in any test node's ResourcesTotal
	for i, n := range nodes {
		ratio := nodeResourceRatio(n, HeatGPU)
		assert.Equal(t, float64(-1), ratio, "node %d should have no GPU", i)
	}
}

func TestRenderHeatmapCell_AliveNode(t *testing.T) {
	nodes := testNodesForHeatmap()
	// worker1: ALIVE
	cell := renderHeatmapCell(nodes[1], HeatCPU)
	assert.NotEmpty(t, cell)
	assert.NotContains(t, cell, "X")
}

func TestRenderHeatmapCell_DeadNode(t *testing.T) {
	nodes := testNodesForHeatmap()
	// worker3: DEAD
	cell := renderHeatmapCell(nodes[3], HeatCPU)
	assert.Contains(t, cell, "X")
}

func TestRenderHeatmapCell_HeadNode(t *testing.T) {
	nodes := testNodesForHeatmap()
	// head node
	cell := renderHeatmapCell(nodes[0], HeatCPU)
	assert.Contains(t, cell, "*")
}

func TestRenderHeatmap_EmptyNodes(t *testing.T) {
	result := renderHeatmap(nil, HeatCPU, 80)
	assert.Equal(t, "No nodes", result)
}

func TestRenderHeatmap_MultipleNodes(t *testing.T) {
	nodes := testNodesForHeatmap()
	result := renderHeatmap(nodes, HeatCPU, 80)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "CPU Usage")
}

func TestHeatmapResourceLabel(t *testing.T) {
	assert.Equal(t, "CPU", heatmapResourceLabel(HeatCPU))
	assert.Equal(t, "Memory", heatmapResourceLabel(HeatMemory))
	assert.Equal(t, "GPU", heatmapResourceLabel(HeatGPU))
}
