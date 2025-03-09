package protocol

import (
	"time"

	"github.com/google/uuid"
)

// SSH constants for the tunnel connection
const (
	DefaultSSHPort   = 2222
	DefaultStartPort = 10000
	DefaultEndPort   = 20000
)

// Status constants for heartbeat messages
const (
	StatusOK       = "ok"
	StatusUpdating = "updating"
	StatusError    = "error"
)

// Command types for server to agent communication
const (
	CmdDeploy       = "deploy"
	CmdUndeploy     = "undeploy"
	CmdUpdateEnvVar = "update_env_var"
	CmdRestart      = "restart"
	CmdExecute      = "execute"
	CmdGetStatus    = "get_status"
	CmdGetLogs      = "get_logs"
)

// Response types for agent to server communication
const (
	RespSuccess = "success"
	RespError   = "error"
	RespStatus  = "status"
	RespLogs    = "logs"
	RespOutput  = "output"
)

// Command represents a message sent from server to agent
type Command struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// Response represents a message sent from agent to server
type Response struct {
	CommandID string                 `json:"command_id,omitempty"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Heartbeat represents a periodic check-in message from agent
type Heartbeat struct {
	DeviceID   string                 `json:"device_id"`
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	IP         string                 `json:"ip"`
	Version    string                 `json:"version"`
	Metrics    map[string]interface{} `json:"metrics,omitempty"`
	Containers []ContainerStatus      `json:"containers,omitempty"`
}

// ContainerStatus represents the status of a container on a device
type ContainerStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Image   string `json:"image"`
	Created string `json:"created"`
}

// DeployPayload represents the payload for a deployment command
type DeployPayload struct {
	SoftwareID    uuid.UUID         `json:"software_id"`
	Version       string            `json:"version"`
	ComposeConfig string            `json:"compose_config"`
	EnvVars       map[string]string `json:"env_vars"`
}

// ExecutePayload represents the payload for an execute command
type ExecutePayload struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout"` // in seconds, 0 means no timeout
}

// StatusPayload represents the payload for a status command
type StatusPayload struct {
	IncludeMetrics     bool `json:"include_metrics"`
	IncludeContainers  bool `json:"include_containers"`
	IncludeSystemStats bool `json:"include_system_stats"`
}

// LogsPayload represents the payload for a logs command
type LogsPayload struct {
	Container string `json:"container"`
	Lines     int    `json:"lines"`
	Follow    bool   `json:"follow"`
}

// LogResponse represents a log entry response
type LogResponse struct {
	Container string    `json:"container"`
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"` // stdout or stderr
	Message   string    `json:"message"`
}

// NewCommand creates a new command with a unique ID
func NewCommand(cmdType string, payload map[string]interface{}) *Command {
	return &Command{
		ID:        uuid.New().String(),
		Type:      cmdType,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// NewResponse creates a new response to a command
func NewResponse(cmdID string, respType string, success bool, message string) *Response {
	return &Response{
		CommandID: cmdID,
		Type:      respType,
		Timestamp: time.Now(),
		Success:   success,
		Message:   message,
		Data:      make(map[string]interface{}),
	}
}

// NewHeartbeat creates a new heartbeat message
func NewHeartbeat(deviceID string, status string) *Heartbeat {
	return &Heartbeat{
		DeviceID:   deviceID,
		Status:     status,
		Timestamp:  time.Now(),
		Containers: make([]ContainerStatus, 0),
		Metrics:    make(map[string]interface{}),
	}
}
