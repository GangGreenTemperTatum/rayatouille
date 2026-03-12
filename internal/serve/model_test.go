package serve

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func strPtr(s string) *string { return &s }

func makeTestApps() *ray.ServeInstanceDetails {
	return &ray.ServeInstanceDetails{
		Applications: map[string]ray.ApplicationDetails{
			"fraud-detector": {
				Name:        "fraud-detector",
				RoutePrefix: strPtr("/fraud"),
				Status:      "RUNNING",
				Message:     "",
				Deployments: map[string]ray.DeploymentDetails{
					"FraudModel":   {Name: "FraudModel", Status: "HEALTHY"},
					"Preprocessor": {Name: "Preprocessor", Status: "HEALTHY"},
				},
			},
			"recommender": {
				Name:        "recommender",
				RoutePrefix: strPtr("/recommend"),
				Status:      "DEPLOYING",
				Message:     "Deploying 2 replicas",
				Deployments: map[string]ray.DeploymentDetails{
					"RecModel": {Name: "RecModel", Status: "UPDATING"},
				},
			},
			"classifier": {
				Name:        "classifier",
				RoutePrefix: strPtr("/classify"),
				Status:      "DEPLOY_FAILED",
				Message:     "Replica failed to start",
				Deployments: map[string]ray.DeploymentDetails{
					"ClassModel": {Name: "ClassModel", Status: "UNHEALTHY"},
					"Tokenizer":  {Name: "Tokenizer", Status: "HEALTHY"},
				},
			},
		},
	}
}

func TestNew_InitialState(t *testing.T) {
	m := New()
	assert.Equal(t, SortByName, m.sortField)
	assert.Equal(t, SortAsc, m.sortOrder)
	assert.Equal(t, StatusAll, m.statusFilter)
	assert.False(t, m.ready)
	assert.Empty(t, m.allApps)
	assert.Empty(t, m.filteredApps)
}

func TestSetApps_PopulatesTable(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	assert.Len(t, m.allApps, 3)
	assert.Len(t, m.filteredApps, 3)
}

func TestSetApps_Nil_ServeNotRunning(t *testing.T) {
	m := New()
	m.SetApps(nil)

	assert.Len(t, m.allApps, 0)
	assert.Len(t, m.filteredApps, 0)
}

func TestSetApps_EmptyApplications(t *testing.T) {
	m := New()
	m.SetApps(&ray.ServeInstanceDetails{Applications: nil})

	assert.Len(t, m.allApps, 0)
	assert.Len(t, m.filteredApps, 0)
}

func TestStatusFilter_Running(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	m.statusFilter = StatusRunning
	m.applyFilters()

	assert.Len(t, m.filteredApps, 1)
	assert.Equal(t, "fraud-detector", m.filteredApps[0].Name)
}

func TestStatusFilter_Deploying(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	m.statusFilter = StatusDeploying
	m.applyFilters()

	assert.Len(t, m.filteredApps, 1)
	assert.Equal(t, "recommender", m.filteredApps[0].Name)
}

func TestStatusFilter_Failed(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	m.statusFilter = StatusFailed
	m.applyFilters()

	assert.Len(t, m.filteredApps, 1)
	assert.Equal(t, "classifier", m.filteredApps[0].Name)
}

func TestTextFilter_ByName(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	m.filter.SetValueForTest("fraud")
	m.applyFilters()

	assert.Len(t, m.filteredApps, 1)
	assert.Equal(t, "fraud-detector", m.filteredApps[0].Name)
}

func TestTextFilter_ByRoute(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	m.filter.SetValueForTest("/recommend")
	m.applyFilters()

	assert.Len(t, m.filteredApps, 1)
	assert.Equal(t, "recommender", m.filteredApps[0].Name)
}

func TestSortToggle(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	assert.Equal(t, SortByName, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByStatus, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByRoute, m.sortField)

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "s"})
	assert.Equal(t, SortByName, m.sortField)
}

func TestFilterActive(t *testing.T) {
	m := New()
	assert.False(t, m.FilterActive())

	m, _ = m.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	assert.True(t, m.FilterActive())
}

func TestSelectAppMsg_OnEnter(t *testing.T) {
	m := New()
	m.SetApps(makeTestApps())

	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Text: "enter"})
	require.NotNil(t, cmd)

	msg := cmd()
	selectMsg, ok := msg.(SelectAppMsg)
	require.True(t, ok)
	// Default sort is by name ascending: classifier < fraud-detector < recommender
	assert.Equal(t, "classifier", selectMsg.Name)
	assert.Equal(t, "DEPLOY_FAILED", selectMsg.App.Status)
}

func TestAppToRow_WithRoute(t *testing.T) {
	apps := makeTestApps()
	row := AppToRow("fraud-detector", apps.Applications["fraud-detector"])

	assert.Equal(t, "fraud-detector", row[0])
	assert.Equal(t, "RUNNING", row[1])
	assert.Equal(t, "/fraud", row[2])
	assert.Equal(t, "2", row[3])
	assert.Equal(t, "-", row[4]) // empty message
}

func TestAppToRow_NilRoute(t *testing.T) {
	app := ray.ApplicationDetails{
		Status:      "RUNNING",
		RoutePrefix: nil,
		Deployments: map[string]ray.DeploymentDetails{},
	}
	row := AppToRow("test-app", app)

	assert.Equal(t, "-", row[2]) // nil route shows "-"
}

func TestAppToRow_TruncateLongName(t *testing.T) {
	app := ray.ApplicationDetails{
		Status:      "RUNNING",
		RoutePrefix: strPtr("/"),
		Deployments: map[string]ray.DeploymentDetails{},
	}
	row := AppToRow("very-long-application-name-here", app)

	assert.Equal(t, "very-long-applicat..", row[0]) // 18 + ".."
}

func TestColumns_WidthDistribution(t *testing.T) {
	cols := Columns(120)
	assert.Len(t, cols, 5)

	totalWidth := 0
	for _, c := range cols {
		totalWidth += c.Width
	}
	assert.LessOrEqual(t, totalWidth, 120)
}
