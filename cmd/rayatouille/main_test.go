package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
)

func setupTestConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	config.SetConfigDirOverride(dir)
	t.Cleanup(func() { config.SetConfigDirOverride("") })
}

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestProfileAdd(t *testing.T) {
	setupTestConfig(t)

	out, err := executeCommand(t, "profile", "add", "test-cluster", "--address", "https://example.com")
	require.NoError(t, err)
	assert.Contains(t, out, "added")
}

func TestProfileList(t *testing.T) {
	setupTestConfig(t)

	// Add a profile first
	_, err := executeCommand(t, "profile", "add", "test-cluster", "--address", "https://example.com")
	require.NoError(t, err)

	out, err := executeCommand(t, "profile", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "test-cluster")
	assert.Contains(t, out, "https://example.com")
}

func TestProfileUse(t *testing.T) {
	setupTestConfig(t)

	// Add a profile first
	_, err := executeCommand(t, "profile", "add", "test-cluster", "--address", "https://example.com")
	require.NoError(t, err)

	out, err := executeCommand(t, "profile", "use", "test-cluster")
	require.NoError(t, err)
	assert.Contains(t, out, "Active profile set")
}

func TestProfileRemove(t *testing.T) {
	setupTestConfig(t)

	// Add a profile first
	_, err := executeCommand(t, "profile", "add", "test-cluster", "--address", "https://example.com")
	require.NoError(t, err)

	out, err := executeCommand(t, "profile", "remove", "test-cluster")
	require.NoError(t, err)
	assert.Contains(t, out, "removed")
}

func TestProfileAddDuplicate(t *testing.T) {
	setupTestConfig(t)

	_, err := executeCommand(t, "profile", "add", "test-cluster", "--address", "https://example.com")
	require.NoError(t, err)

	_, err = executeCommand(t, "profile", "add", "test-cluster", "--address", "https://other.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCompletion(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			out, err := executeCommand(t, "completion", shell)
			require.NoError(t, err)
			assert.NotEmpty(t, out, "completion output for %s should not be empty", shell)
		})
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	_, err := executeCommand(t, "completion", "invalid")
	require.Error(t, err)
}

func TestVersionFlag(t *testing.T) {
	out, err := executeCommand(t, "--version")
	require.NoError(t, err)
	assert.Contains(t, out, "dev")
}
