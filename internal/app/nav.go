package app

import "strings"

// View represents which view is currently active.
type View int

const (
	// ViewDashboard is the cluster overview dashboard.
	ViewDashboard View = iota
	// ViewJobs is the jobs list view.
	ViewJobs
	// ViewJobDetail is the job detail view.
	ViewJobDetail
	// ViewNodes is the nodes list view.
	ViewNodes
	// ViewNodeDetail is the node detail view.
	ViewNodeDetail
	// ViewActors is the actors list view.
	ViewActors
	// ViewActorDetail is the actor detail view.
	ViewActorDetail
	// ViewServe is the serve applications list view.
	ViewServe
	// ViewServeDetail is the serve application detail view.
	ViewServeDetail
	// ViewEvents is the cluster events timeline view.
	ViewEvents
)

// viewName maps a View to its display name.
func viewName(v View) string {
	switch v {
	case ViewDashboard:
		return "Cluster"
	case ViewJobs:
		return "Jobs"
	case ViewJobDetail:
		return "Job Detail"
	case ViewNodes:
		return "Nodes"
	case ViewNodeDetail:
		return "Node Detail"
	case ViewActors:
		return "Actors"
	case ViewActorDetail:
		return "Actor Detail"
	case ViewServe:
		return "Serve"
	case ViewServeDetail:
		return "Serve Detail"
	case ViewEvents:
		return "Events"
	default:
		return "Unknown"
	}
}

// NavStack manages a stack of views for navigation.
type NavStack struct {
	stack []View
}

// NewNavStack returns a NavStack initialized with ViewDashboard as root.
func NewNavStack() NavStack {
	return NavStack{
		stack: []View{ViewDashboard},
	}
}

// Push adds a view to the top of the stack.
func (n *NavStack) Push(v View) {
	n.stack = append(n.stack, v)
}

// Pop removes and returns the top view. Never pops the root view.
func (n *NavStack) Pop() View {
	if len(n.stack) <= 1 {
		return n.stack[0]
	}
	top := n.stack[len(n.stack)-1]
	n.stack = n.stack[:len(n.stack)-1]
	return top
}

// Current returns the top view on the stack.
func (n *NavStack) Current() View {
	return n.stack[len(n.stack)-1]
}

// Depth returns the number of views on the stack.
func (n *NavStack) Depth() int {
	return len(n.stack)
}

// Breadcrumbs returns view names joined with " > ".
func (n *NavStack) Breadcrumbs() string {
	names := make([]string, len(n.stack))
	for i, v := range n.stack {
		names[i] = viewName(v)
	}
	return strings.Join(names, " > ")
}
