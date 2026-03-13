package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

func main() {
	address := os.Getenv("RAY_DASHBOARD_URL")
	if address == "" {
		address = "http://localhost:8265"
	}

	timeout := 10 * time.Second
	client := ray.NewClient(address, timeout)

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "rayatouille",
			Version: "0.1.0",
		},
		nil,
	)

	registerTools(server, client)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

var emptySchema = &jsonschema.Schema{Type: "object"}

func registerTools(server *mcp.Server, client ray.Client) {
	addSimpleTool(server, "ray_cluster_health",
		"Get Ray cluster health: version info, node counts, resource totals, job status summary",
		func(ctx context.Context) (*mcp.CallToolResult, error) {
			version, err := client.Ping(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Error connecting to cluster: %v", err)), nil
			}

			nodes, err := client.ListNodes(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Connected (Ray %s) but failed to list nodes: %v", version.RayVersion, err)), nil
			}

			jobs, err := client.ListJobs(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Connected (Ray %s) but failed to list jobs: %v", version.RayVersion, err)), nil
			}

			alive, dead := 0, 0
			var totalCPU, availCPU, totalMem, availMem float64
			for _, n := range nodes {
				if n.State == "ALIVE" {
					alive++
				} else {
					dead++
				}
				totalCPU += n.ResourcesTotal["CPU"]
				availCPU += n.ResourcesAvailable["CPU"]
				totalMem += n.ResourcesTotal["memory"]
				availMem += n.ResourcesAvailable["memory"]
			}

			running, failed, succeeded, pending := 0, 0, 0, 0
			for _, j := range jobs {
				switch j.Status {
				case "RUNNING":
					running++
				case "FAILED":
					failed++
				case "SUCCEEDED":
					succeeded++
				case "PENDING":
					pending++
				}
			}

			result := fmt.Sprintf(`Ray Cluster Health
Version: Ray %s (API %s)
Session: %s

Nodes: %d alive, %d dead (%d total)
CPU: %.0f/%.0f used
Memory: %.1f/%.1f GB used

Jobs: %d running, %d pending, %d failed, %d succeeded (%d total)`,
				version.RayVersion, version.Version, version.SessionName,
				alive, dead, len(nodes),
				totalCPU-availCPU, totalCPU,
				(totalMem-availMem)/1e9, totalMem/1e9,
				running, pending, failed, succeeded, len(jobs),
			)

			return textResult(result), nil
		})

	addSimpleTool(server, "ray_list_jobs",
		"List all Ray jobs with status, submission ID, entrypoint, and timing",
		func(ctx context.Context) (*mcp.CallToolResult, error) {
			jobs, err := client.ListJobDetails(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			return mustJSON(jobs), nil
		})

	addSimpleTool(server, "ray_list_nodes",
		"List all Ray cluster nodes with state, IP, resources, and labels",
		func(ctx context.Context) (*mcp.CallToolResult, error) {
			nodes, err := client.ListNodes(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			return mustJSON(nodes), nil
		})

	addSimpleTool(server, "ray_list_actors",
		"List all Ray actors with state, class name, job ID, PID, and node",
		func(ctx context.Context) (*mcp.CallToolResult, error) {
			actors, err := client.ListActors(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			return mustJSON(actors), nil
		})

	addSimpleTool(server, "ray_serve_status",
		"Get Ray Serve application status, deployments, and replica health",
		func(ctx context.Context) (*mcp.CallToolResult, error) {
			details, err := client.GetServeApplications(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			if details == nil {
				return textResult("Ray Serve is not running or not deployed on this cluster."), nil
			}
			return mustJSON(details), nil
		})

	addSimpleTool(server, "ray_cluster_events",
		"List cluster events with severity, timestamp, source, and message",
		func(ctx context.Context) (*mcp.CallToolResult, error) {
			events, err := client.ListClusterEvents(ctx)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			if len(events) == 0 {
				return textResult("No cluster events found."), nil
			}
			return mustJSON(events), nil
		})

	// Tools with parameters use the low-level API with custom schemas.

	server.AddTool(
		&mcp.Tool{
			Name:        "ray_job_logs",
			Description: "Get logs for a specific Ray job by submission ID",
			InputSchema: stringParamSchema("submission_id", "The job submission ID (e.g. raysubmit_XXX)"),
		},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := stringArg(req, "submission_id")
			logs, err := client.GetJobLogs(ctx, id)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			if logs == "" {
				return textResult("(no logs available)"), nil
			}
			return textResult(logs), nil
		})

	server.AddTool(
		&mcp.Tool{
			Name:        "ray_task_summary",
			Description: "Get task summary for a specific Ray job by job ID",
			InputSchema: stringParamSchema("job_id", "The Ray job ID (e.g. 01000000)"),
		},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := stringArg(req, "job_id")
			summary, err := client.GetTaskSummary(ctx, id)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			return mustJSON(summary), nil
		})

	server.AddTool(
		&mcp.Tool{
			Name:        "ray_node_logs",
			Description: "List available log files for a specific node",
			InputSchema: stringParamSchema("node_id", "The Ray node ID"),
		},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := stringArg(req, "node_id")
			listing, err := client.ListNodeLogs(ctx, id)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			return mustJSON(listing.Categories), nil
		})

	server.AddTool(
		&mcp.Tool{
			Name:        "ray_node_log_file",
			Description: "Get the contents of a specific log file from a node",
			InputSchema: twoStringParamSchema("node_id", "The Ray node ID", "filename", "The log filename to retrieve"),
		},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			nodeID := stringArg(req, "node_id")
			filename := stringArg(req, "filename")
			content, err := client.GetNodeLogFile(ctx, nodeID, filename)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			if content == "" {
				return textResult("(empty log file)"), nil
			}
			return textResult(content), nil
		})

	server.AddTool(
		&mcp.Tool{
			Name:        "ray_actor_logs",
			Description: "Get stdout logs for a specific actor by actor ID",
			InputSchema: stringParamSchema("actor_id", "The Ray actor ID"),
		},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := stringArg(req, "actor_id")
			logs, err := client.GetActorLogs(ctx, id)
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil
			}
			if logs == "" {
				return textResult("(no logs available)"), nil
			}
			return textResult(logs), nil
		})
}

// addSimpleTool registers a tool with no input parameters.
func addSimpleTool(server *mcp.Server, name, description string, fn func(ctx context.Context) (*mcp.CallToolResult, error)) {
	server.AddTool(
		&mcp.Tool{
			Name:        name,
			Description: description,
			InputSchema: emptySchema,
		},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return fn(ctx)
		},
	)
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

func mustJSON(v any) *mcp.CallToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return textResult(fmt.Sprintf("Error marshaling result: %v", err))
	}
	return textResult(string(data))
}

func stringArg(req *mcp.CallToolRequest, key string) string {
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return ""
	}
	s, _ := args[key].(string)
	return s
}

func stringParamSchema(name, description string) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{name},
		Properties: map[string]*jsonschema.Schema{
			name: {Type: "string", Description: description},
		},
	}
}

func twoStringParamSchema(name1, desc1, name2, desc2 string) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{name1, name2},
		Properties: map[string]*jsonschema.Schema{
			name1: {Type: "string", Description: desc1},
			name2: {Type: "string", Description: desc2},
		},
	}
}
