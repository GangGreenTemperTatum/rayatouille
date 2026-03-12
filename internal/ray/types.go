package ray

// StateAPIResponse is the generic wrapper for /api/v0/* endpoints.
type StateAPIResponse[T any] struct {
	Result bool   `json:"result"`
	Msg    string `json:"msg"`
	Data   struct {
		Result StateAPIResult[T] `json:"result"`
	} `json:"data"`
}

// StateAPIResult contains the inner result data from State API responses.
type StateAPIResult[T any] struct {
	Total              int    `json:"total"`
	NumAfterTruncation int    `json:"num_after_truncation"`
	NumFiltered        int    `json:"num_filtered"`
	Result             []T    `json:"result"`
	PartialFailureWarn string `json:"partial_failure_warning"`
}

// VersionInfo is returned by GET /api/version.
type VersionInfo struct {
	Version     string `json:"version"`
	RayVersion  string `json:"ray_version"`
	RayCommit   string `json:"ray_commit"`
	SessionName string `json:"session_name"`
}

// Job represents a Ray job from the State API (/api/v0/jobs).
type Job struct {
	Type         string      `json:"type"`
	SubmissionID string      `json:"submission_id"`
	JobID        *string     `json:"job_id"`     // nullable
	ErrorType    *string     `json:"error_type"` // nullable
	Entrypoint   string      `json:"entrypoint"`
	Status       string      `json:"status"`
	DriverInfo   *DriverInfo `json:"driver_info"` // nullable
	Message      string      `json:"message"`
	// detail=true fields
	StartTime              int64             `json:"start_time,omitempty"`
	EndTime                int64             `json:"end_time,omitempty"`
	DriverExitCode         *int              `json:"driver_exit_code,omitempty"`
	DriverAgentHTTPAddress string            `json:"driver_agent_http_address,omitempty"`
	DriverNodeID           string            `json:"driver_node_id,omitempty"`
	Metadata               map[string]string `json:"metadata,omitempty"`
	RuntimeEnv             map[string]any    `json:"runtime_env,omitempty"`
}

// DriverInfo contains driver process information for a job.
type DriverInfo struct {
	ID            string `json:"id"`
	NodeIPAddress string `json:"node_ip_address"`
	PID           string `json:"pid"`
}

// JobDetail represents a job from the Jobs REST API (/api/jobs/).
// This endpoint returns a bare JSON array (no StateAPIResponse wrapper)
// and includes richer fields like start_time, end_time, metadata, etc.
type JobDetail struct {
	Type                   string            `json:"type"`
	JobID                  *string           `json:"job_id"` // nullable
	SubmissionID           string            `json:"submission_id"`
	DriverInfo             *DriverInfo       `json:"driver_info"` // nullable
	Status                 string            `json:"status"`
	Entrypoint             string            `json:"entrypoint"`
	Message                string            `json:"message"`
	ErrorType              *string           `json:"error_type"` // nullable
	StartTime              int64             `json:"start_time"`
	EndTime                int64             `json:"end_time"`
	Metadata               map[string]string `json:"metadata"`
	RuntimeEnv             map[string]any    `json:"runtime_env"`
	DriverAgentHTTPAddress string            `json:"driver_agent_http_address"`
	DriverNodeID           string            `json:"driver_node_id"`
	DriverExitCode         *int              `json:"driver_exit_code"`
}

// Node represents a Ray cluster node from /api/v0/nodes.
type Node struct {
	State              string             `json:"state"`
	ResourcesTotal     map[string]float64 `json:"resources_total"`
	ResourcesAvailable map[string]float64 `json:"resources_available"`
	StateMessage       *string            `json:"state_message"` // nullable
	NodeIP             string             `json:"node_ip"`
	IsHeadNode         bool               `json:"is_head_node"`
	Labels             map[string]string  `json:"labels"`
	NodeID             string             `json:"node_id"`
	NodeName           string             `json:"node_name"`
	// detail=true fields
	StartTimeMs int64 `json:"start_time_ms,omitempty"`
	EndTimeMs   int64 `json:"end_time_ms,omitempty"`
}

// Actor represents a Ray actor from /api/v0/actors.
type Actor struct {
	State        string `json:"state"`
	ActorID      string `json:"actor_id"`
	ClassName    string `json:"class_name"`
	JobID        string `json:"job_id"`
	RayNamespace string `json:"ray_namespace"`
	PID          int    `json:"pid"`
	NodeID       string `json:"node_id"`
	Name         string `json:"name"`
	// detail=true fields
	IsDetached        bool           `json:"is_detached,omitempty"`
	PlacementGroupID  *string        `json:"placement_group_id,omitempty"` // nullable
	ReprName          string         `json:"repr_name,omitempty"`
	RequiredResources map[string]any `json:"required_resources,omitempty"`
	DeathCause        *DeathCause    `json:"death_cause,omitempty"` // nullable
	NumRestarts       string         `json:"num_restarts,omitempty"`
	CallSite          *string        `json:"call_site,omitempty"` // nullable
}

// DeathCause contains the reason an actor died.
type DeathCause struct {
	ActorDiedErrorContext *ActorDiedErrorContext `json:"actor_died_error_context,omitempty"`
}

// ActorDiedErrorContext contains detailed error information for a dead actor.
type ActorDiedErrorContext struct {
	ErrorMessage   string `json:"error_message"`
	OwnerID        string `json:"owner_id"`
	OwnerIPAddress string `json:"owner_ip_address"`
	NodeIPAddress  string `json:"node_ip_address"`
	PID            int    `json:"pid"`
	Name           string `json:"name"`
	RayNamespace   string `json:"ray_namespace"`
	ClassName      string `json:"class_name"`
	ActorID        string `json:"actor_id"`
	Reason         string `json:"reason"`
	NeverStarted   bool   `json:"never_started"`
}

// TaskSummaryResponse is the inner result from /api/v0/tasks/summarize.
// Access via: data.result.result.node_id_to_summary["cluster"].summary
type TaskSummaryResponse struct {
	NodeIDToSummary map[string]NodeTaskSummary `json:"node_id_to_summary"`
}

// NodeTaskSummary contains task summary data for a node (or "cluster" aggregate).
type NodeTaskSummary struct {
	Summary             map[string]TaskFuncSummary `json:"summary"`
	TotalTasks          int                        `json:"total_tasks"`
	TotalActorTasks     int                        `json:"total_actor_tasks"`
	TotalActorScheduled int                        `json:"total_actor_scheduled"`
	SummaryBy           string                     `json:"summary_by"`
}

// TaskFuncSummary summarizes tasks by function/class name with state counts.
type TaskFuncSummary struct {
	FuncOrClassName string         `json:"func_or_class_name"`
	Type            string         `json:"type"`
	StateCounts     map[string]int `json:"state_counts"`
}

// NodeLogListing contains categorized log file names for a node.
type NodeLogListing struct {
	Categories map[string][]string // category name -> list of filenames
}

// ServeActorDetails contains Serve controller/proxy actor info.
type ServeActorDetails struct {
	NodeID      *string `json:"node_id"`
	NodeIP      *string `json:"node_ip"`
	ActorID     *string `json:"actor_id"`
	ActorName   *string `json:"actor_name"`
	WorkerID    *string `json:"worker_id"`
	LogFilePath *string `json:"log_file_path"`
	Status      string  `json:"status"`
}

// ProxyDetails contains Serve proxy information.
type ProxyDetails struct {
	Status      string  `json:"status"`
	NodeID      *string `json:"node_id"`
	NodeIP      *string `json:"node_ip"`
	ActorID     *string `json:"actor_id"`
	ActorName   *string `json:"actor_name"`
	WorkerID    *string `json:"worker_id"`
	LogFilePath *string `json:"log_file_path"`
}

// ServeInstanceDetails is the top-level response from GET /api/serve/applications/.
// This is NOT wrapped in StateAPIResponse -- it's parsed directly.
type ServeInstanceDetails struct {
	ControllerInfo *ServeActorDetails            `json:"controller_info"`
	ProxyLocation  string                        `json:"proxy_location"`
	HTTPOptions    map[string]any                `json:"http_options"`
	GRPCOptions    map[string]any                `json:"grpc_options"`
	Proxies        map[string]ProxyDetails       `json:"proxies"`
	DeployMode     string                        `json:"deploy_mode"`
	Applications   map[string]ApplicationDetails `json:"applications"`
	TargetCapacity *float64                      `json:"target_capacity"`
}

// ApplicationDetails represents a single Serve application.
type ApplicationDetails struct {
	Name              string                       `json:"name"`
	RoutePrefix       *string                      `json:"route_prefix"`
	DocsPath          *string                      `json:"docs_path"`
	Status            string                       `json:"status"`
	Message           string                       `json:"message"`
	LastDeployedTimeS float64                      `json:"last_deployed_time_s"`
	DeployedAppConfig map[string]any               `json:"deployed_app_config"`
	Deployments       map[string]DeploymentDetails `json:"deployments"`
}

// DeploymentDetails represents a single deployment within a Serve application.
type DeploymentDetails struct {
	Name              string           `json:"name"`
	Status            string           `json:"status"`
	StatusTrigger     string           `json:"status_trigger"`
	Message           string           `json:"message"`
	TargetNumReplicas int              `json:"target_num_replicas"`
	Replicas          []ReplicaDetails `json:"replicas"`
}

// ReplicaDetails represents a single replica within a deployment.
type ReplicaDetails struct {
	ReplicaID   string  `json:"replica_id"`
	State       string  `json:"state"`
	PID         *int    `json:"pid"`
	ActorName   *string `json:"actor_name"`
	ActorID     *string `json:"actor_id"`
	NodeID      *string `json:"node_id"`
	NodeIP      *string `json:"node_ip"`
	StartTimeS  float64 `json:"start_time_s"`
	LogFilePath *string `json:"log_file_path"`
	WorkerID    *string `json:"worker_id"`
}

// ClusterEvent represents a cluster event from /api/v0/cluster_events.
type ClusterEvent struct {
	Severity     string         `json:"severity"`
	Time         string         `json:"time"`
	SourceType   string         `json:"source_type"`
	Message      string         `json:"message"`
	EventID      string         `json:"event_id"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}
