package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/edgetainer/edgetainer/internal/server/auth"
	"github.com/edgetainer/edgetainer/internal/server/provisioning"
	"github.com/edgetainer/edgetainer/internal/shared/models"
	"github.com/google/uuid"
)

// DeviceProvisionRequest represents a request for provisioning a new device
type DeviceProvisionRequest struct {
	Name        string            `json:"name"`
	FleetID     string            `json:"fleet_id,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Description string            `json:"description,omitempty"`
}

// DeviceProvisionResponse represents a response for a device provisioning request
type DeviceProvisionResponse struct {
	DeviceID  string `json:"device_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	ConfigURL string `json:"config_url"`
}

// handleDeviceProvisioning handles creating a new device provisioning configuration
func (s *Server) handleDeviceProvisioning(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var request DeviceProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate the request
	if request.Name == "" {
		http.Error(w, "Device name is required", http.StatusBadRequest)
		return
	}

	// Generate a unique device ID
	deviceID := generateDeviceID(request.Name)

	// Generate SSH key pair for the device
	keyPair, err := auth.GenerateKeyPair(deviceID, 4096)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to generate key pair: %v", err), err)
		http.Error(w, "Failed to generate key pair", http.StatusInternalServerError)
		return
	}

	// Extract the public and private keys as strings
	publicKeyString := string(keyPair.PublicKey)
	privateKeyString := string(keyPair.PrivateKey)

	// Parse the fleet ID if provided
	var fleetID *uuid.UUID
	if request.FleetID != "" {
		parsedID, err := uuid.Parse(request.FleetID)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Invalid fleet ID: %v", err), err)
			http.Error(w, "Invalid fleet ID", http.StatusBadRequest)
			return
		}
		fleetID = &parsedID
	}

	// No need to handle labels separately, as we're using the Device model directly

	// Create a pending device record in the database
	device := models.Device{
		DeviceID:     deviceID,
		Name:         request.Name,
		FleetID:      fleetID,
		Status:       models.DeviceStatusPending,
		LastSeen:     time.Now(),
		SSHPublicKey: publicKeyString,
		SSHPort:      2222, // Default SSH port
		HardwareInfo: "{}", // Initialize with empty JSON object
	}

	result := s.database.GetDB().Create(&device)
	if result.Error != nil {
		s.logger.Error(fmt.Sprintf("Failed to create pending device: %v", result.Error), result.Error)
		http.Error(w, "Failed to create pending device", http.StatusInternalServerError)
		return
	}

	s.logger.Info(fmt.Sprintf("Created pending device %s (%s)", request.Name, deviceID))

	// Set up template data with the private key
	templateData := &provisioning.TemplateData{
		DeviceID:      deviceID,
		SSHPrivateKey: privateKeyString,
		ServerHost:    s.host,
		ServerPort:    s.port,
		SSHPort:       2222,
	}

	// Get the template path
	templatePath := filepath.Join("config", "templates", "base.bu")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// If not found in development path, try the Docker container path
		templatePath = filepath.Join("/app", "templates", "base.bu")
	}

	// Render the Butane template
	butaneConfig, err := provisioning.RenderButaneTemplate(templatePath, templateData)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to render butane template: %v", err), err)
		http.Error(w, "Failed to render butane template", http.StatusInternalServerError)
		return
	}

	// Convert to Ignition JSON
	ignitionJSON, err := provisioning.ConvertButaneToIgnition(butaneConfig)
	if err != nil {
		s.logger.Info(fmt.Sprintf("Failed to convert butane to ignition (falling back to raw template): %v", err))

		// For now, respond with the Butane template as JSON
		response := map[string]interface{}{
			"device_id":       deviceID,
			"name":            request.Name,
			"status":          models.DeviceStatusPending,
			"butane_template": butaneConfig,
			"note":            "Butane conversion failed. For production, please install butane CLI or use the Go library.",
		}
		jsonResponse(w, response, http.StatusOK)
		return
	}

	// Return the Ignition configuration directly
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.ign\"", deviceID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ignitionJSON))
}

// generateDeviceID generates a unique device ID
func generateDeviceID(name string) string {
	// This is a simplified implementation
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("device-%s-%d", name, timestamp%100000) // Shorten the ID a bit
}
