package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_FlagTakesPrecedence(t *testing.T) {
	t.Setenv("RAY_DASHBOARD_URL", "https://env-cluster.example.com")

	cfg := Config{Address: "https://flag-cluster.example.com"}
	cfg.Resolve()

	assert.Equal(t, "https://flag-cluster.example.com", cfg.Address)
}

func TestResolve_EnvVarFallback(t *testing.T) {
	t.Setenv("RAY_DASHBOARD_URL", "https://env-cluster.example.com")

	cfg := Config{}
	cfg.Resolve()

	assert.Equal(t, "https://env-cluster.example.com", cfg.Address)
}

func TestResolve_Defaults(t *testing.T) {
	cfg := Config{Address: "https://example.com"}
	cfg.Resolve()

	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Equal(t, 5*time.Second, cfg.RefreshInterval)
}

func TestResolve_StripsTrailingSlash(t *testing.T) {
	cfg := Config{Address: "https://example.com/"}
	cfg.Resolve()

	assert.Equal(t, "https://example.com", cfg.Address)
}

func TestValidate_MissingAddress(t *testing.T) {
	cfg := Config{}
	cfg.Resolve()

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--address or RAY_DASHBOARD_URL is required")
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := Config{Address: "https://example.com"}
	cfg.Resolve()

	err := cfg.Validate()
	assert.NoError(t, err)
}
