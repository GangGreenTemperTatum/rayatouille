package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds CLI configuration for rayatouille.
type Config struct {
	Address         string
	Timeout         time.Duration
	RefreshInterval time.Duration
}

// Resolve fills in missing values from environment variables, active profile, and defaults.
// Priority: CLI flag > environment variable > active profile > default value.
func (c *Config) Resolve() {
	if c.Address == "" {
		c.Address = os.Getenv("RAY_DASHBOARD_URL")
	}

	// Fill from active profile (only zero-valued fields).
	if profile, err := LoadActiveProfile(); err == nil {
		if c.Address == "" {
			c.Address = profile.Address
		}
		if c.Timeout == 0 {
			c.Timeout = profile.TimeoutDuration()
		}
		if c.RefreshInterval == 0 {
			c.RefreshInterval = profile.RefreshIntervalDuration()
		}
	}

	// Strip trailing slash to prevent double-slash in URL construction.
	c.Address = strings.TrimRight(c.Address, "/")

	if c.Timeout == 0 {
		c.Timeout = 10 * time.Second
	}
	if c.RefreshInterval == 0 {
		c.RefreshInterval = 5 * time.Second
	}
}

// Validate checks that required configuration is present.
// Call Resolve() before Validate().
func (c *Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("--address or RAY_DASHBOARD_URL is required")
	}
	return nil
}
