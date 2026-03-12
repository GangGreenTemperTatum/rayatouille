package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfigDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	configDirOverride = dir
	t.Cleanup(func() { configDirOverride = "" })
}

func TestLoadProfileConfig_NoFile(t *testing.T) {
	setupTestConfigDir(t)

	cfg, err := LoadProfileConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.ActiveProfile)
	assert.NotNil(t, cfg.Profiles)
	assert.Len(t, cfg.Profiles, 0)
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	setupTestConfigDir(t)

	original := &ProfileConfig{
		ActiveProfile: "prod",
		Profiles: map[string]Profile{
			"prod": {
				Address: "https://ray-prod.example.com",
				Timeout: "30s",
			},
			"staging": {
				Address:         "https://ray-staging.example.com",
				Timeout:         "5s",
				RefreshInterval: "10s",
			},
		},
	}

	err := SaveProfileConfig(original)
	require.NoError(t, err)

	loaded, err := LoadProfileConfig()
	require.NoError(t, err)

	assert.Equal(t, "prod", loaded.ActiveProfile)
	assert.Len(t, loaded.Profiles, 2)
	assert.Equal(t, "https://ray-prod.example.com", loaded.Profiles["prod"].Address)
	assert.Equal(t, "30s", loaded.Profiles["prod"].Timeout)
	assert.Equal(t, "https://ray-staging.example.com", loaded.Profiles["staging"].Address)
	assert.Equal(t, "10s", loaded.Profiles["staging"].RefreshInterval)
}

func TestLoadActiveProfile_NotSet(t *testing.T) {
	setupTestConfigDir(t)

	_, err := LoadActiveProfile()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active profile set")
}

func TestLoadActiveProfile_Found(t *testing.T) {
	setupTestConfigDir(t)

	cfg := &ProfileConfig{
		ActiveProfile: "dev",
		Profiles: map[string]Profile{
			"dev": {
				Address:         "http://localhost:8265",
				Timeout:         "15s",
				RefreshInterval: "3s",
			},
		},
	}
	require.NoError(t, SaveProfileConfig(cfg))

	profile, err := LoadActiveProfile()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8265", profile.Address)
	assert.Equal(t, "15s", profile.Timeout)
	assert.Equal(t, "3s", profile.RefreshInterval)
}

func TestSetActiveProfile_Valid(t *testing.T) {
	setupTestConfigDir(t)

	cfg := &ProfileConfig{
		Profiles: map[string]Profile{
			"alpha": {Address: "https://alpha.example.com"},
			"beta":  {Address: "https://beta.example.com"},
		},
	}
	require.NoError(t, SaveProfileConfig(cfg))

	err := SetActiveProfile("beta")
	require.NoError(t, err)

	loaded, err := LoadProfileConfig()
	require.NoError(t, err)
	assert.Equal(t, "beta", loaded.ActiveProfile)
}

func TestSetActiveProfile_NotFound(t *testing.T) {
	setupTestConfigDir(t)

	cfg := &ProfileConfig{
		Profiles: map[string]Profile{
			"alpha": {Address: "https://alpha.example.com"},
		},
	}
	require.NoError(t, SaveProfileConfig(cfg))

	err := SetActiveProfile("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListProfileNames(t *testing.T) {
	setupTestConfigDir(t)

	cfg := &ProfileConfig{
		Profiles: map[string]Profile{
			"charlie": {Address: "https://c.example.com"},
			"alpha":   {Address: "https://a.example.com"},
			"bravo":   {Address: "https://b.example.com"},
		},
	}
	require.NoError(t, SaveProfileConfig(cfg))

	names, err := ListProfileNames()
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "bravo", "charlie"}, names)
}

func TestProfileDurations(t *testing.T) {
	t.Run("parses valid durations", func(t *testing.T) {
		p := Profile{Timeout: "30s", RefreshInterval: "10s"}
		assert.Equal(t, 30*time.Second, p.TimeoutDuration())
		assert.Equal(t, 10*time.Second, p.RefreshIntervalDuration())
	})

	t.Run("defaults on empty", func(t *testing.T) {
		p := Profile{}
		assert.Equal(t, 10*time.Second, p.TimeoutDuration())
		assert.Equal(t, 5*time.Second, p.RefreshIntervalDuration())
	})

	t.Run("defaults on invalid", func(t *testing.T) {
		p := Profile{Timeout: "not-a-duration", RefreshInterval: "xyz"}
		assert.Equal(t, 10*time.Second, p.TimeoutDuration())
		assert.Equal(t, 5*time.Second, p.RefreshIntervalDuration())
	})
}
