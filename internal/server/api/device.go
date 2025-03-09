package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/edgetainer/edgetainer/internal/shared/models"
)

// handleDevices handles the devices endpoint
func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List devices
		var devices []models.Device

		// Fetch devices from the database
		result := s.database.GetDB().Find(&devices)
		if result.Error != nil {
			s.logger.Error("Failed to fetch devices", result.Error)
			http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, devices, http.StatusOK)

	case http.MethodPost:
		// Create device
		var device models.Device

		if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Validate the device
		if device.Name == "" {
			http.Error(w, "Device name is required", http.StatusBadRequest)
			return
		}

		// Ensure hardware_info is a valid JSON object
		if device.HardwareInfo == "" {
			device.HardwareInfo = "{}" // Initialize with empty JSON object
		}

		// Save to the database
		if err := s.database.GetDB().Create(&device).Error; err != nil {
			s.logger.Error("Failed to create device", err)
			http.Error(w, "Failed to create device", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, device, http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDeviceByID handles the device by ID endpoint
func (s *Server) handleDeviceByID(w http.ResponseWriter, r *http.Request) {
	// Extract device ID from URL
	deviceID := filepath.Base(r.URL.Path)

	s.logger.Info(fmt.Sprintf("Device operation on ID: %s", deviceID))

	switch r.Method {
	case http.MethodGet:
		// Get device by ID
		var device models.Device

		// Fetch the device from the database
		result := s.database.GetDB().Where("device_id = ?", deviceID).First(&device)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to fetch device %s", deviceID), result.Error)
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		jsonResponse(w, device, http.StatusOK)

	case http.MethodPut:
		// Update device
		var device models.Device

		if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Validate the device
		if device.Name == "" {
			http.Error(w, "Device name is required", http.StatusBadRequest)
			return
		}

		// Ensure hardware_info is a valid JSON object
		if device.HardwareInfo == "" {
			device.HardwareInfo = "{}" // Initialize with empty JSON object
		}

		// Ensure deviceID from URL matches the one in the request
		device.DeviceID = deviceID

		// Update in the database
		result := s.database.GetDB().Where("device_id = ?", deviceID).Updates(&device)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to update device %s", deviceID), result.Error)
			http.Error(w, "Failed to update device", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		// Fetch the updated device to return
		s.database.GetDB().Where("device_id = ?", deviceID).First(&device)
		jsonResponse(w, device, http.StatusOK)

	case http.MethodDelete:
		// Delete device
		result := s.database.GetDB().Where("device_id = ?", deviceID).Delete(&models.Device{})
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to delete device %s", deviceID), result.Error)
			http.Error(w, "Failed to delete device", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
