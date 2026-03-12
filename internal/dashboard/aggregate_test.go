package dashboard

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func TestAggregateClusterHealth_AllAlive(t *testing.T) {
	nodes := []ray.Node{
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 4, "GPU": 1, "memory": 8e9, "object_store_memory": 2e9}},
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 8, "GPU": 2, "memory": 16e9, "object_store_memory": 4e9}},
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 4, "GPU": 0, "memory": 8e9, "object_store_memory": 2e9}},
	}

	h := AggregateClusterHealth(nodes)

	assert.Equal(t, 3, h.NodeCount)
	assert.Equal(t, 3, h.AliveNodes)
	assert.Equal(t, 16.0, h.TotalCPU)
	assert.Equal(t, 3.0, h.TotalGPU)
	assert.Equal(t, 32e9, h.TotalMemory)
	assert.Equal(t, 8e9, h.TotalObjectStoreMemory)
	assert.Equal(t, "healthy", h.Status)
}

func TestAggregateClusterHealth_SomeDead(t *testing.T) {
	nodes := []ray.Node{
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 4, "memory": 8e9}},
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 4, "memory": 8e9}},
		{State: "DEAD", ResourcesTotal: map[string]float64{"CPU": 4, "memory": 8e9}},
	}

	h := AggregateClusterHealth(nodes)

	assert.Equal(t, 3, h.NodeCount)
	assert.Equal(t, 2, h.AliveNodes)
	assert.Equal(t, "degraded", h.Status)
}

func TestAggregateClusterHealth_NoNodes(t *testing.T) {
	h := AggregateClusterHealth([]ray.Node{})

	assert.Equal(t, 0, h.NodeCount)
	assert.Equal(t, 0, h.AliveNodes)
	assert.Equal(t, 0.0, h.TotalCPU)
	assert.Equal(t, 0.0, h.TotalGPU)
	assert.Equal(t, 0.0, h.TotalMemory)
	assert.Equal(t, 0.0, h.TotalObjectStoreMemory)
	assert.Equal(t, "unhealthy", h.Status)
}

func TestAggregateClusterHealth_AllDead(t *testing.T) {
	nodes := []ray.Node{
		{State: "DEAD", ResourcesTotal: map[string]float64{"CPU": 4}},
		{State: "DEAD", ResourcesTotal: map[string]float64{"CPU": 8}},
	}

	h := AggregateClusterHealth(nodes)

	assert.Equal(t, 2, h.NodeCount)
	assert.Equal(t, 0, h.AliveNodes)
	assert.Equal(t, "unhealthy", h.Status)
}

func TestAggregateClusterHealth_ResourceSumming(t *testing.T) {
	// Some nodes have GPU, some don't; some have object_store_memory, some don't
	nodes := []ray.Node{
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 4, "GPU": 2, "memory": 8e9}},
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 8, "memory": 16e9, "object_store_memory": 4e9}},
		{State: "ALIVE", ResourcesTotal: map[string]float64{"CPU": 2, "GPU": 1, "memory": 4e9, "object_store_memory": 1e9}},
	}

	h := AggregateClusterHealth(nodes)

	assert.Equal(t, 14.0, h.TotalCPU)
	assert.Equal(t, 3.0, h.TotalGPU)               // 2 + 0 + 1
	assert.Equal(t, 28e9, h.TotalMemory)           // 8 + 16 + 4
	assert.Equal(t, 5e9, h.TotalObjectStoreMemory) // 0 + 4 + 1
	assert.Equal(t, 0.0, h.UsedCPU)                // No resources_available provided
	assert.False(t, h.HasAvailableData)
}

func TestAggregateClusterHealth_WithResourcesAvailable(t *testing.T) {
	nodes := []ray.Node{
		{
			State:              "ALIVE",
			ResourcesTotal:     map[string]float64{"CPU": 8, "GPU": 2, "memory": 16e9, "object_store_memory": 4e9},
			ResourcesAvailable: map[string]float64{"CPU": 3, "GPU": 1, "memory": 4e9, "object_store_memory": 1e9},
		},
		{
			State:              "ALIVE",
			ResourcesTotal:     map[string]float64{"CPU": 4, "GPU": 0, "memory": 8e9, "object_store_memory": 2e9},
			ResourcesAvailable: map[string]float64{"CPU": 2, "GPU": 0, "memory": 6e9, "object_store_memory": 2e9},
		},
	}

	h := AggregateClusterHealth(nodes)

	assert.True(t, h.HasAvailableData)
	// Node 1: used CPU = 8-3 = 5, Node 2: used CPU = 4-2 = 2 → total used = 7
	assert.Equal(t, 7.0, h.UsedCPU)
	// Node 1: used GPU = 2-1 = 1, Node 2: used GPU = 0-0 = 0 → total used = 1
	assert.Equal(t, 1.0, h.UsedGPU)
	// Node 1: used mem = 16e9-4e9 = 12e9, Node 2: used mem = 8e9-6e9 = 2e9 → 14e9
	assert.Equal(t, 14e9, h.UsedMemory)
	// Node 1: used obj = 4e9-1e9 = 3e9, Node 2: used obj = 2e9-2e9 = 0 → 3e9
	assert.Equal(t, 3e9, h.UsedObjectStoreMemory)
	// Totals still correct
	assert.Equal(t, 12.0, h.TotalCPU)
	assert.Equal(t, 2.0, h.TotalGPU)
	assert.Equal(t, 24e9, h.TotalMemory)
	assert.Equal(t, 6e9, h.TotalObjectStoreMemory)
}

func TestAggregateClusterHealth_MixedAvailability(t *testing.T) {
	// One node has resources_available, one doesn't
	nodes := []ray.Node{
		{
			State:              "ALIVE",
			ResourcesTotal:     map[string]float64{"CPU": 8, "memory": 16e9},
			ResourcesAvailable: map[string]float64{"CPU": 3, "memory": 4e9},
		},
		{
			State:          "ALIVE",
			ResourcesTotal: map[string]float64{"CPU": 4, "memory": 8e9},
			// No ResourcesAvailable
		},
	}

	h := AggregateClusterHealth(nodes)

	assert.True(t, h.HasAvailableData, "should be true if any node has available data")
	// Only node 1 contributes to used: 8-3 = 5
	assert.Equal(t, 5.0, h.UsedCPU)
	assert.Equal(t, 12e9, h.UsedMemory) // 16e9 - 4e9
}

func TestAggregateJobSummary_MixedStatuses(t *testing.T) {
	jobs := []ray.JobDetail{
		{Status: "RUNNING"},
		{Status: "RUNNING"},
		{Status: "PENDING"},
		{Status: "FAILED"},
		{Status: "SUCCEEDED"},
		{Status: "SUCCEEDED"},
		{Status: "STOPPED"},
	}

	s := AggregateJobSummary(jobs)

	assert.Equal(t, 2, s.Running)
	assert.Equal(t, 1, s.Pending)
	assert.Equal(t, 1, s.Failed)
	assert.Equal(t, 2, s.Succeeded)
	assert.Equal(t, 1, s.Stopped)
	assert.Equal(t, 7, s.Total)
}

func TestAggregateJobSummary_Empty(t *testing.T) {
	s := AggregateJobSummary([]ray.JobDetail{})

	assert.Equal(t, 0, s.Running)
	assert.Equal(t, 0, s.Pending)
	assert.Equal(t, 0, s.Failed)
	assert.Equal(t, 0, s.Succeeded)
	assert.Equal(t, 0, s.Stopped)
	assert.Equal(t, 0, s.Total)
}

func TestAggregateJobSummary_UnknownStatus(t *testing.T) {
	jobs := []ray.JobDetail{
		{Status: "RUNNING"},
		{Status: "UNKNOWN_STATUS"},
		{Status: "SOME_FUTURE_STATUS"},
	}

	s := AggregateJobSummary(jobs)

	assert.Equal(t, 1, s.Running)
	assert.Equal(t, 3, s.Total) // Unknown statuses still counted in Total
}
