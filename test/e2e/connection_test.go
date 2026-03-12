//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getClusterURL reads RAY_DASHBOARD_URL from the environment.
// If not set, the test is skipped.
func getClusterURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("RAY_DASHBOARD_URL")
	if url == "" {
		t.Skip("RAY_DASHBOARD_URL not set, skipping e2e tests")
	}
	return url
}

func TestConnection_Ping(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	version, err := client.Ping(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, version.RayVersion, "RayVersion should not be empty")
	assert.NotEmpty(t, version.Version, "Version should not be empty")
	t.Logf("Connected to Ray %s", version.RayVersion)
}

func TestConnection_InvalidURL(t *testing.T) {
	client := ray.NewClient("https://nonexistent.invalid:9999", 2*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Ping(ctx)
	assert.Error(t, err, "Ping to invalid URL should return an error")
}

func TestConnection_VersionFields(t *testing.T) {
	url := getClusterURL(t)
	client := ray.NewClient(url, 10*time.Second)

	version, err := client.Ping(context.Background())
	require.NoError(t, err)

	assert.NotEmpty(t, version.Version, "Version should not be empty")
	assert.NotEmpty(t, version.RayVersion, "RayVersion should not be empty")
	assert.NotEmpty(t, version.RayCommit, "RayCommit should not be empty")
	assert.NotEmpty(t, version.SessionName, "SessionName should not be empty")

	t.Logf("Version: %s, RayVersion: %s, Commit: %s, Session: %s",
		version.Version, version.RayVersion, version.RayCommit, version.SessionName)
}
