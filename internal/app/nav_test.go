package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNavStack_StartsAtDashboard(t *testing.T) {
	nav := NewNavStack()
	assert.Equal(t, ViewDashboard, nav.Current())
	assert.Equal(t, 1, nav.Depth())
}

func TestNavStack_Push(t *testing.T) {
	nav := NewNavStack()
	nav.Push(ViewJobs)

	assert.Equal(t, ViewJobs, nav.Current())
	assert.Equal(t, 2, nav.Depth())
}

func TestNavStack_Pop(t *testing.T) {
	nav := NewNavStack()
	nav.Push(ViewJobs)

	popped := nav.Pop()
	assert.Equal(t, ViewJobs, popped)
	assert.Equal(t, ViewDashboard, nav.Current())
	assert.Equal(t, 1, nav.Depth())
}

func TestNavStack_Pop_NeverPopsRoot(t *testing.T) {
	nav := NewNavStack()

	popped := nav.Pop()
	assert.Equal(t, ViewDashboard, popped)
	assert.Equal(t, 1, nav.Depth())
	assert.Equal(t, ViewDashboard, nav.Current())

	// Pop again -- still safe.
	popped = nav.Pop()
	assert.Equal(t, ViewDashboard, popped)
	assert.Equal(t, 1, nav.Depth())
}

func TestNavStack_Breadcrumbs_SingleView(t *testing.T) {
	nav := NewNavStack()
	assert.Equal(t, "Cluster", nav.Breadcrumbs())
}

func TestNavStack_Breadcrumbs_MultipleViews(t *testing.T) {
	nav := NewNavStack()
	nav.Push(ViewJobs)
	assert.Equal(t, "Cluster > Jobs", nav.Breadcrumbs())
}

func TestViewName(t *testing.T) {
	assert.Equal(t, "Cluster", viewName(ViewDashboard))
	assert.Equal(t, "Jobs", viewName(ViewJobs))
	assert.Equal(t, "Unknown", viewName(View(99)))
}
