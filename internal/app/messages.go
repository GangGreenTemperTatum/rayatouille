package app

import (
	"time"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// ErrMsg is sent when an async operation encounters an error.
type ErrMsg struct {
	Err error
}

// TickMsg is sent on each periodic polling tick.
type TickMsg time.Time

// UITickMsg is sent every second to force UI re-renders (e.g., status bar timer).
type UITickMsg time.Time

// ClusterDataMsg carries fetched cluster data as a Bubble Tea message.
type ClusterDataMsg struct {
	Nodes    []ray.Node
	Jobs     []ray.JobDetail
	Actors   []ray.Actor
	Serve    *ray.ServeInstanceDetails
	Events   []ray.ClusterEvent
	FetchErr error
	Latency  time.Duration
}

// SwitchProfileMsg requests a runtime switch to a named profile.
type SwitchProfileMsg struct {
	Name string
}

// ProfileSwitchedMsg carries the result of a profile switch attempt.
type ProfileSwitchedMsg struct {
	Name    string
	Client  ray.Client
	Version *ray.VersionInfo
	Err     error
}
