package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/edgetainer/edgetainer/internal/shared/protocol"
)

// handleAgentHeartbeat handles the agent heartbeat endpoint
func (s *Server) handleAgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var heartbeat protocol.Heartbeat

	if err := json.NewDecoder(r.Body).Decode(&heartbeat); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// This is a placeholder - in a real implementation, you would:
	// 1. Update device status in the database
	// 2. Process container status updates

	s.logger.Info(fmt.Sprintf("Received heartbeat from device %s with status %s", heartbeat.DeviceID, heartbeat.Status))

	// Send a response with the current time
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	jsonResponse(w, response, http.StatusOK)
}

// handleAgentStatus handles the agent status endpoint
func (s *Server) handleAgentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var statusReport struct {
		DeviceID   string                     `json:"device_id"`
		Status     string                     `json:"status"`
		Metrics    map[string]interface{}     `json:"metrics"`
		Containers []protocol.ContainerStatus `json:"containers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&statusReport); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// This is a placeholder - in a real implementation, you would:
	// 1. Update device status in the database
	// 2. Process container status updates

	s.logger.Info(fmt.Sprintf("Received status report from device %s with %d containers",
		statusReport.DeviceID, len(statusReport.Containers)))

	// Send a response
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	jsonResponse(w, response, http.StatusOK)
}
