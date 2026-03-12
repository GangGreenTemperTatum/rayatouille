package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// configDirOverride allows tests to redirect config file operations to a temp directory.
var configDirOverride string

// SetConfigDirOverride sets the config directory override for test isolation.
// Pass an empty string to clear the override.
func SetConfigDirOverride(dir string) {
	configDirOverride = dir
}

// ProfileConfig holds all saved cluster profiles and the active profile name.
type ProfileConfig struct {
	ActiveProfile string             `yaml:"active_profile,omitempty"`
	Profiles      map[string]Profile `yaml:"profiles,omitempty"`
}

// Profile holds connection settings for a single Ray cluster.
type Profile struct {
	Address         string `yaml:"address"`
	Timeout         string `yaml:"timeout,omitempty"`
	RefreshInterval string `yaml:"refresh_interval,omitempty"`
}

// TimeoutDuration parses the Timeout string as a duration.
// Returns 10s default on empty or parse error.
func (p Profile) TimeoutDuration() time.Duration {
	if p.Timeout == "" {
		return 10 * time.Second
	}
	d, err := time.ParseDuration(p.Timeout)
	if err != nil {
		return 10 * time.Second
	}
	return d
}

// RefreshIntervalDuration parses the RefreshInterval string as a duration.
// Returns 5s default on empty or parse error.
func (p Profile) RefreshIntervalDuration() time.Duration {
	if p.RefreshInterval == "" {
		return 5 * time.Second
	}
	d, err := time.ParseDuration(p.RefreshInterval)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

// ConfigDir returns the rayatouille config directory path.
func ConfigDir() (string, error) {
	if configDirOverride != "" {
		return configDirOverride, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	return filepath.Join(base, "rayatouille"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LoadProfileConfig reads the profile config from disk.
// Returns an empty config (not an error) if the file does not exist.
func LoadProfileConfig() (*ProfileConfig, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProfileConfig{Profiles: make(map[string]Profile)}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg ProfileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	return &cfg, nil
}

// SaveProfileConfig writes the profile config to disk, creating directories as needed.
func SaveProfileConfig(cfg *ProfileConfig) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// LoadActiveProfile returns the currently active profile.
// Returns an error if no active profile is set or the named profile does not exist.
func LoadActiveProfile() (*Profile, error) {
	cfg, err := LoadProfileConfig()
	if err != nil {
		return nil, err
	}

	if cfg.ActiveProfile == "" {
		return nil, fmt.Errorf("no active profile set")
	}

	p, ok := cfg.Profiles[cfg.ActiveProfile]
	if !ok {
		return nil, fmt.Errorf("active profile %q not found in config", cfg.ActiveProfile)
	}
	return &p, nil
}

// ListProfileNames returns a sorted list of all profile names.
func ListProfileNames() ([]string, error) {
	cfg, err := LoadProfileConfig()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// SetActiveProfile sets the active profile to the given name.
// Returns an error if the profile does not exist.
func SetActiveProfile(name string) error {
	cfg, err := LoadProfileConfig()
	if err != nil {
		return err
	}

	if _, ok := cfg.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}

	cfg.ActiveProfile = name
	return SaveProfileConfig(cfg)
}
