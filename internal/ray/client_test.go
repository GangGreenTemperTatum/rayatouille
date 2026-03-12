package ray

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../test/testdata/" + name)
	require.NoError(t, err, "failed to load fixture: %s", name)
	return data
}

// --- Ping tests ---

func TestPing_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/version", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "version.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	info, err := client.Ping(context.Background())
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "4", info.Version)
	assert.Equal(t, "2.50.0", info.RayVersion)
	assert.Equal(t, "276c75c1c3e4a5510377c86f89f0b615ca802e31", info.RayCommit)
	assert.Equal(t, "session_2026-03-11_16-48-34_967147_1", info.SessionName)
}

func TestPing_Unreachable(t *testing.T) {
	client := NewClient("http://127.0.0.1:1", 1*time.Second)
	_, err := client.Ping(context.Background())
	assert.Error(t, err)
}

func TestPing_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// --- ListJobs tests ---

func TestListJobs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v0/jobs", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "jobs_state.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	jobs, err := client.ListJobs(context.Background())
	require.NoError(t, err)
	assert.Len(t, jobs, 3)

	// First job: SUCCEEDED with all fields
	assert.Equal(t, "SUCCEEDED", jobs[0].Status)
	assert.Equal(t, "SUBMISSION", jobs[0].Type)
	assert.Equal(t, "raysubmit_ABWGkRenHVev9s7A", jobs[0].SubmissionID)
	require.NotNil(t, jobs[0].JobID)
	assert.Equal(t, "01000000", *jobs[0].JobID)
	assert.Nil(t, jobs[0].ErrorType)
	require.NotNil(t, jobs[0].DriverInfo)
	assert.Equal(t, "01000000", jobs[0].DriverInfo.ID)

	// Second job: FAILED with error_type
	assert.Equal(t, "FAILED", jobs[1].Status)
	require.NotNil(t, jobs[1].ErrorType)
	assert.Equal(t, "JOB_ENTRYPOINT_COMMAND_ERROR", *jobs[1].ErrorType)
}

func TestListJobs_NullableFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "jobs_state.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	jobs, err := client.ListJobs(context.Background())
	require.NoError(t, err)

	// Third job has null job_id and null driver_info
	nullJob := jobs[2]
	assert.Nil(t, nullJob.JobID, "job_id should be nil for null value")
	assert.Nil(t, nullJob.DriverInfo, "driver_info should be nil for null value")
	assert.Nil(t, nullJob.ErrorType, "error_type should be nil for null value")

	// Accessing nil pointer fields should not panic
	if nullJob.JobID != nil {
		_ = *nullJob.JobID
	}
	if nullJob.DriverInfo != nil {
		_ = nullJob.DriverInfo.ID
	}
}

func TestListJobs_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.ListJobs(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestListJobs_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client := NewClient(server.URL, 100*time.Millisecond)
	_, err := client.ListJobs(context.Background())
	assert.Error(t, err)
}

// --- ListNodes tests ---

func TestListNodes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v0/nodes", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "nodes_state.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	nodes, err := client.ListNodes(context.Background())
	require.NoError(t, err)
	assert.Len(t, nodes, 2)

	assert.Equal(t, "ALIVE", nodes[0].State)
	assert.Equal(t, "100.64.32.128", nodes[0].NodeIP)
	assert.Equal(t, 4.0, nodes[0].ResourcesTotal["CPU"])
}

func TestListNodes_HeadNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "nodes_state.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	nodes, err := client.ListNodes(context.Background())
	require.NoError(t, err)

	// First node is head node
	assert.True(t, nodes[0].IsHeadNode, "first node should be head node")
	assert.False(t, nodes[1].IsHeadNode, "second node should NOT be head node")

	// Second node is DEAD with state_message
	assert.Equal(t, "DEAD", nodes[1].State)
	require.NotNil(t, nodes[1].StateMessage)
	assert.Contains(t, *nodes[1].StateMessage, "autoscaler")

	// First node has null state_message
	assert.Nil(t, nodes[0].StateMessage)
}

// --- ListActors tests ---

func TestListActors_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v0/actors", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "actors_state.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	actors, err := client.ListActors(context.Background())
	require.NoError(t, err)
	assert.Len(t, actors, 2)

	assert.Equal(t, "ALIVE", actors[0].State)
	assert.Equal(t, "RayTrainWorker", actors[0].ClassName)
	assert.Equal(t, 54321, actors[0].PID)
}

func TestListActors_DeathCause(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "actors_state.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	actors, err := client.ListActors(context.Background())
	require.NoError(t, err)

	// First actor (ALIVE) has no death_cause
	assert.Nil(t, actors[0].DeathCause)

	// Second actor (DEAD) has death_cause with nested error context
	deadActor := actors[1]
	assert.Equal(t, "DEAD", deadActor.State)
	require.NotNil(t, deadActor.DeathCause)
	require.NotNil(t, deadActor.DeathCause.ActorDiedErrorContext)
	assert.Equal(t, "WORKER_DIED", deadActor.DeathCause.ActorDiedErrorContext.Reason)
	assert.Contains(t, deadActor.DeathCause.ActorDiedErrorContext.ErrorMessage, "worker process has died")
	assert.Equal(t, 180530, deadActor.DeathCause.ActorDiedErrorContext.PID)
	assert.False(t, deadActor.DeathCause.ActorDiedErrorContext.NeverStarted)
}

// --- ListJobDetails tests ---

func TestListJobDetails_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/jobs/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "jobs_rest.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	jobs, err := client.ListJobDetails(context.Background())
	require.NoError(t, err)
	assert.Len(t, jobs, 3)

	// First job: SUCCEEDED with timestamps
	assert.Equal(t, "SUCCEEDED", jobs[0].Status)
	assert.Equal(t, "raysubmit_ABWGkRenHVev9s7A", jobs[0].SubmissionID)
	assert.Equal(t, int64(1773279930789), jobs[0].StartTime)
	assert.Equal(t, int64(1773279931659), jobs[0].EndTime)
	assert.Nil(t, jobs[0].JobID)
	require.NotNil(t, jobs[0].DriverExitCode)
	assert.Equal(t, 0, *jobs[0].DriverExitCode)

	// Second job: FAILED with metadata
	assert.Equal(t, "FAILED", jobs[1].Status)
	require.NotNil(t, jobs[1].JobID)
	assert.Equal(t, "02000000", *jobs[1].JobID)
	assert.Equal(t, "ads", jobs[1].Metadata["user"])
}

// --- Endpoint path tests ---

func TestEndpointPaths(t *testing.T) {
	receivedPaths := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPaths[r.URL.Path] = true
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/version":
			w.Write(loadFixture(t, "version.json"))
		case "/api/v0/jobs":
			w.Write(loadFixture(t, "jobs_state.json"))
		case "/api/v0/nodes":
			w.Write(loadFixture(t, "nodes_state.json"))
		case "/api/v0/actors":
			w.Write(loadFixture(t, "actors_state.json"))
		case "/api/jobs/":
			w.Write(loadFixture(t, "jobs_rest.json"))
		case "/api/jobs/raysubmit_test123/logs":
			w.Write(loadFixture(t, "job_logs.json"))
		case "/api/v0/tasks/summarize":
			w.Write(loadFixture(t, "task_summary.json"))
		case "/api/v0/logs":
			w.Write(loadFixture(t, "node_logs.json"))
		case "/api/v0/logs/file":
			w.Header().Set("Content-Type", "text/plain")
			w.Write(loadFixture(t, "node_log_file.txt"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	ctx := context.Background()

	_, err := client.Ping(ctx)
	require.NoError(t, err)
	_, err = client.ListJobs(ctx)
	require.NoError(t, err)
	_, err = client.ListNodes(ctx)
	require.NoError(t, err)
	_, err = client.ListActors(ctx)
	require.NoError(t, err)
	_, err = client.ListJobDetails(ctx)
	require.NoError(t, err)
	_, err = client.GetJobLogs(ctx, "raysubmit_test123")
	require.NoError(t, err)
	_, err = client.GetTaskSummary(ctx, "02000000")
	require.NoError(t, err)
	_, err = client.ListNodeLogs(ctx, "node123")
	require.NoError(t, err)
	_, err = client.GetNodeLogFile(ctx, "node123", "raylet.out")
	require.NoError(t, err)

	// Verify all expected paths were hit
	assert.True(t, receivedPaths["/api/version"], "should hit /api/version")
	assert.True(t, receivedPaths["/api/v0/jobs"], "should hit /api/v0/jobs (no trailing slash)")
	assert.True(t, receivedPaths["/api/v0/nodes"], "should hit /api/v0/nodes (no trailing slash)")
	assert.True(t, receivedPaths["/api/v0/actors"], "should hit /api/v0/actors (no trailing slash)")
	assert.True(t, receivedPaths["/api/jobs/"], "should hit /api/jobs/ (with trailing slash)")
	assert.True(t, receivedPaths["/api/jobs/raysubmit_test123/logs"], "should hit /api/jobs/{id}/logs")
	assert.True(t, receivedPaths["/api/v0/tasks/summarize"], "should hit /api/v0/tasks/summarize")
	assert.True(t, receivedPaths["/api/v0/logs"], "should hit /api/v0/logs")
	assert.True(t, receivedPaths["/api/v0/logs/file"], "should hit /api/v0/logs/file")

	// Verify endpoint helper functions produce correct strings
	assert.Equal(t, "/api/jobs/raysubmit_XXX/logs", JobLogsEndpoint("raysubmit_XXX"))
	assert.Contains(t, TasksSummarizeEndpoint("02000000"), "/api/v0/tasks/summarize")
	assert.Contains(t, TasksSummarizeEndpoint("02000000"), "filter_keys=job_id")
	assert.Contains(t, TasksSummarizeEndpoint("02000000"), "filter_values=02000000")
	assert.Equal(t, "/api/v0/logs?node_id=abc123", NodeLogsEndpoint("abc123"))
	assert.Contains(t, NodeLogFileEndpoint("abc123", "raylet.out"), "/api/v0/logs/file")
	assert.Contains(t, NodeLogFileEndpoint("abc123", "raylet.out"), "node_id=abc123")
	assert.Contains(t, NodeLogFileEndpoint("abc123", "raylet.out"), "filename=raylet.out")
}

// --- GetJobLogs tests ---

func TestGetJobLogs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/jobs/raysubmit_test123/logs", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "job_logs.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	logs, err := client.GetJobLogs(context.Background(), "raysubmit_test123")
	require.NoError(t, err)
	assert.Contains(t, logs, "Starting worker")
	assert.Contains(t, logs, "Task completed successfully")
	assert.Contains(t, logs, "WARNING")
}

func TestGetJobLogs_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.GetJobLogs(context.Background(), "raysubmit_test123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestGetJobLogs_EmptyLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"logs": ""}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	logs, err := client.GetJobLogs(context.Background(), "raysubmit_empty")
	require.NoError(t, err)
	assert.Equal(t, "", logs)
}

// --- GetTaskSummary tests ---

func TestGetTaskSummary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v0/tasks/summarize", r.URL.Path)
		assert.Equal(t, "job_id", r.URL.Query().Get("filter_keys"))
		assert.Equal(t, "=", r.URL.Query().Get("filter_predicates"))
		assert.Equal(t, "02000000", r.URL.Query().Get("filter_values"))
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "task_summary.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	summary, err := client.GetTaskSummary(context.Background(), "02000000")
	require.NoError(t, err)
	require.NotNil(t, summary)

	cluster, ok := summary.NodeIDToSummary["cluster"]
	require.True(t, ok, "should have 'cluster' key")
	assert.Equal(t, 3, cluster.TotalActorTasks)
	assert.Equal(t, "func_name", cluster.SummaryBy)

	scanTarget, ok := cluster.Summary["WebScanBatchWorker.scan_target"]
	require.True(t, ok)
	assert.Equal(t, "ACTOR_TASK", scanTarget.Type)
	assert.Equal(t, 3, scanTarget.StateCounts["FINISHED"])
}

func TestGetTaskSummary_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.GetTaskSummary(context.Background(), "02000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestGetTaskSummary_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": false, "msg": "internal error", "data": {"result": {"result": {}}}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.GetTaskSummary(context.Background(), "02000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal error")
}

func TestGetTaskSummary_EmptySummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": true, "msg": "", "data": {"result": {"total": 0, "result": {"node_id_to_summary": {"cluster": {"summary": {}, "total_tasks": 0, "total_actor_tasks": 0, "total_actor_scheduled": 0, "summary_by": "func_name"}}}}}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	summary, err := client.GetTaskSummary(context.Background(), "02000000")
	require.NoError(t, err)
	require.NotNil(t, summary)
	cluster := summary.NodeIDToSummary["cluster"]
	assert.Empty(t, cluster.Summary)
}

// --- ListNodeLogs tests ---

func TestListNodeLogs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v0/logs", r.URL.Path)
		assert.Equal(t, "node123", r.URL.Query().Get("node_id"))
		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "node_logs.json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	listing, err := client.ListNodeLogs(context.Background(), "node123")
	require.NoError(t, err)
	require.NotNil(t, listing)
	assert.Len(t, listing.Categories["gcs_server"], 2)
	assert.Len(t, listing.Categories["raylet"], 2)
	assert.Len(t, listing.Categories["dashboard"], 1)
	assert.Len(t, listing.Categories["worker_out"], 1)
	assert.Len(t, listing.Categories["worker_err"], 1)
}

func TestListNodeLogs_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.ListNodeLogs(context.Background(), "node123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestListNodeLogs_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": false, "msg": "node not found", "data": {"result": {}}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.ListNodeLogs(context.Background(), "badnode")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
}

// --- GetNodeLogFile tests ---

func TestGetNodeLogFile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v0/logs/file", r.URL.Path)
		assert.Equal(t, "node123", r.URL.Query().Get("node_id"))
		assert.Equal(t, "raylet.out", r.URL.Query().Get("filename"))
		w.Header().Set("Content-Type", "text/plain")
		w.Write(loadFixture(t, "node_log_file.txt"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	content, err := client.GetNodeLogFile(context.Background(), "node123", "raylet.out")
	require.NoError(t, err)
	assert.Contains(t, content, "Starting worker")
	assert.Contains(t, content, "Low memory")
	assert.Contains(t, content, "Task completed")
}

func TestGetNodeLogFile_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.GetNodeLogFile(context.Background(), "node123", "raylet.out")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestGetNodeLogFile_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// Write empty body.
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	content, err := client.GetNodeLogFile(context.Background(), "node123", "empty.log")
	require.NoError(t, err)
	assert.Equal(t, "", content)
}

func TestNewClient_StripsTrailingSlash(t *testing.T) {
	client := NewClient("http://localhost:8265/", 5*time.Second)
	assert.Equal(t, "http://localhost:8265", client.baseURL)

	client2 := NewClient("http://localhost:8265", 5*time.Second)
	assert.Equal(t, "http://localhost:8265", client2.baseURL)
}
