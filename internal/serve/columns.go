package serve

import (
	"fmt"

	"charm.land/bubbles/v2/table"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// Columns returns the table column definitions for the given width.
func Columns(width int) []table.Column {
	const (
		appWidth    = 20
		statusWidth = 14
		routeWidth  = 15
		deployWidth = 12
		padding     = 6 // cell padding for 5 columns
	)

	messageWidth := width - appWidth - statusWidth - routeWidth - deployWidth - padding
	if messageWidth < 10 {
		messageWidth = 10
	}

	return []table.Column{
		{Title: "Application", Width: appWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Route", Width: routeWidth},
		{Title: "Deployments", Width: deployWidth},
		{Title: "Message", Width: messageWidth},
	}
}

// AppToRow converts an application name and details into a table row.
func AppToRow(name string, app ray.ApplicationDetails) table.Row {
	displayName := name
	if len(displayName) > 18 {
		displayName = displayName[:18] + ".."
	}

	route := "-"
	if app.RoutePrefix != nil {
		route = *app.RoutePrefix
	}

	deployments := fmt.Sprintf("%d", len(app.Deployments))

	message := app.Message
	if message == "" {
		message = "-"
	}

	return table.Row{displayName, app.Status, route, deployments, message}
}
