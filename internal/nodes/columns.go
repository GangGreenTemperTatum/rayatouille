package nodes

import (
	"fmt"

	"charm.land/bubbles/v2/table"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// Columns returns the table column definitions for the given width.
func Columns(width int) []table.Column {
	const (
		idWidth       = 12
		statusWidth   = 8
		roleWidth     = 6
		cpuWidth      = 6
		gpuWidth      = 6
		memoryWidth   = 10
		objStoreWidth = 10
		padding       = 8 // cell padding
	)

	ipWidth := width - idWidth - statusWidth - roleWidth - cpuWidth - gpuWidth - memoryWidth - objStoreWidth - padding
	if ipWidth < 12 {
		ipWidth = 12
	}

	return []table.Column{
		{Title: "Node ID", Width: idWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Role", Width: roleWidth},
		{Title: "IP", Width: ipWidth},
		{Title: "CPU", Width: cpuWidth},
		{Title: "GPU", Width: gpuWidth},
		{Title: "Memory", Width: memoryWidth},
		{Title: "ObjStore", Width: objStoreWidth},
	}
}

// NodeToRow converts a Node into a table row.
func NodeToRow(n ray.Node) table.Row {
	nodeID := n.NodeID
	if len(nodeID) > 10 {
		nodeID = nodeID[:10] + ".."
	}

	status := n.State

	role := "worker"
	if n.IsHeadNode {
		role = "head"
	}

	ip := n.NodeIP

	cpu := fmt.Sprintf("%.0f", n.ResourcesTotal["CPU"])
	gpu := fmt.Sprintf("%.0f", n.ResourcesTotal["GPU"])
	if gpu == "0" {
		gpu = "-"
	}

	memory := fmt.Sprintf("%.1f GiB", n.ResourcesTotal["memory"]/1073741824.0)
	objStore := fmt.Sprintf("%.1f GiB", n.ResourcesTotal["object_store_memory"]/1073741824.0)

	return table.Row{nodeID, status, role, ip, cpu, gpu, memory, objStore}
}
