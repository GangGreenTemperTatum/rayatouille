package actors

import (
	"fmt"

	"charm.land/bubbles/v2/table"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// Columns returns the table column definitions for the given width.
func Columns(width int) []table.Column {
	const (
		actorIDWidth = 14
		stateWidth   = 10
		classWidth   = 25
		pidWidth     = 8
		jobIDWidth   = 12
		padding      = 7 // cell padding for 6 columns
	)

	nameWidth := width - actorIDWidth - stateWidth - classWidth - pidWidth - jobIDWidth - padding
	if nameWidth < 10 {
		nameWidth = 10
	}

	return []table.Column{
		{Title: "Actor ID", Width: actorIDWidth},
		{Title: "State", Width: stateWidth},
		{Title: "Class", Width: classWidth},
		{Title: "Name", Width: nameWidth},
		{Title: "PID", Width: pidWidth},
		{Title: "Job ID", Width: jobIDWidth},
	}
}

// ActorToRow converts an Actor into a table row.
func ActorToRow(a ray.Actor) table.Row {
	actorID := a.ActorID
	if len(actorID) > 12 {
		actorID = actorID[:12] + ".."
	}

	className := a.ClassName
	if len(className) > 20 {
		className = className[:20] + "..."
	}

	name := a.Name
	if name == "" {
		name = "-"
	}

	pid := "-"
	if a.PID != 0 {
		pid = fmt.Sprintf("%d", a.PID)
	}

	return table.Row{actorID, a.State, className, name, pid, a.JobID}
}
