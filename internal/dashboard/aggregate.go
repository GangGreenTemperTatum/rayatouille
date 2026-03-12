package dashboard

import (
	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// ClusterHealth summarizes the health and resource capacity of the cluster.
type ClusterHealth struct {
	NodeCount              int
	AliveNodes             int
	TotalCPU               float64
	UsedCPU                float64
	TotalGPU               float64
	UsedGPU                float64
	TotalMemory            float64 // bytes
	UsedMemory             float64 // bytes
	TotalObjectStoreMemory float64 // bytes
	UsedObjectStoreMemory  float64 // bytes
	HasAvailableData       bool    // true if any node reported resources_available
	Status                 string  // "healthy", "degraded", "unhealthy"
}

// JobSummary summarizes job counts by status.
type JobSummary struct {
	Running   int
	Pending   int
	Failed    int
	Succeeded int
	Stopped   int
	Total     int
}

// AggregateClusterHealth computes cluster health from a slice of nodes.
func AggregateClusterHealth(nodes []ray.Node) ClusterHealth {
	h := ClusterHealth{
		NodeCount: len(nodes),
	}
	for _, n := range nodes {
		if n.State == "ALIVE" {
			h.AliveNodes++
		}
		h.TotalCPU += n.ResourcesTotal["CPU"]
		h.TotalGPU += n.ResourcesTotal["GPU"]
		h.TotalMemory += n.ResourcesTotal["memory"]
		h.TotalObjectStoreMemory += n.ResourcesTotal["object_store_memory"]

		if n.ResourcesAvailable != nil {
			h.HasAvailableData = true
			h.UsedCPU += n.ResourcesTotal["CPU"] - n.ResourcesAvailable["CPU"]
			h.UsedGPU += n.ResourcesTotal["GPU"] - n.ResourcesAvailable["GPU"]
			h.UsedMemory += n.ResourcesTotal["memory"] - n.ResourcesAvailable["memory"]
			h.UsedObjectStoreMemory += n.ResourcesTotal["object_store_memory"] - n.ResourcesAvailable["object_store_memory"]
		}
	}
	switch {
	case h.AliveNodes == h.NodeCount && h.NodeCount > 0:
		h.Status = "healthy"
	case h.AliveNodes > 0:
		h.Status = "degraded"
	default:
		h.Status = "unhealthy"
	}
	return h
}

// AggregateJobSummary computes job counts by status from a slice of job details.
func AggregateJobSummary(jobs []ray.JobDetail) JobSummary {
	s := JobSummary{Total: len(jobs)}
	for _, j := range jobs {
		switch j.Status {
		case "RUNNING":
			s.Running++
		case "PENDING":
			s.Pending++
		case "FAILED":
			s.Failed++
		case "SUCCEEDED":
			s.Succeeded++
		case "STOPPED":
			s.Stopped++
		}
	}
	return s
}
