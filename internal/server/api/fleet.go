package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/edgetainer/edgetainer/internal/shared/models"
)

// handleFleets handles the fleets endpoint
func (s *Server) handleFleets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List fleets
		var fleets []models.Fleet

		// Fetch fleets from the database
		result := s.database.GetDB().Find(&fleets)
		if result.Error != nil {
			s.logger.Error("Failed to fetch fleets", result.Error)
			http.Error(w, "Failed to fetch fleets", http.StatusInternalServerError)
			return
		}

		// Optionally load related devices for each fleet
		for i := range fleets {
			s.database.GetDB().Model(&fleets[i]).Association("Devices").Find(&fleets[i].Devices)
		}

		jsonResponse(w, fleets, http.StatusOK)

	case http.MethodPost:
		// Create fleet
		var fleet models.Fleet

		if err := json.NewDecoder(r.Body).Decode(&fleet); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Validate the fleet
		if fleet.Name == "" {
			http.Error(w, "Fleet name is required", http.StatusBadRequest)
			return
		}

		// Save to the database
		if err := s.database.GetDB().Create(&fleet).Error; err != nil {
			s.logger.Error("Failed to create fleet", err)
			http.Error(w, "Failed to create fleet", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, fleet, http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFleetByID handles the fleet by ID endpoint
func (s *Server) handleFleetByID(w http.ResponseWriter, r *http.Request) {
	// Extract fleet ID from URL
	fleetID := filepath.Base(r.URL.Path)

	s.logger.Info(fmt.Sprintf("Fleet operation on ID: %s", fleetID))

	switch r.Method {
	case http.MethodGet:
		// Get fleet by ID
		var fleet models.Fleet

		// Fetch the fleet from the database
		result := s.database.GetDB().First(&fleet, fleetID)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to fetch fleet %s", fleetID), result.Error)
			http.Error(w, "Fleet not found", http.StatusNotFound)
			return
		}

		// Load related devices
		s.database.GetDB().Model(&fleet).Association("Devices").Find(&fleet.Devices)

		jsonResponse(w, fleet, http.StatusOK)

	case http.MethodPut:
		// Update fleet
		var fleet models.Fleet

		if err := json.NewDecoder(r.Body).Decode(&fleet); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Validate the fleet
		if fleet.Name == "" {
			http.Error(w, "Fleet name is required", http.StatusBadRequest)
			return
		}

		// Update in the database
		result := s.database.GetDB().Model(&models.Fleet{}).Where("id = ?", fleetID).Updates(fleet)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to update fleet %s", fleetID), result.Error)
			http.Error(w, "Failed to update fleet", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "Fleet not found", http.StatusNotFound)
			return
		}

		// Fetch the updated fleet to return
		s.database.GetDB().First(&fleet, fleetID)
		s.database.GetDB().Model(&fleet).Association("Devices").Find(&fleet.Devices)

		jsonResponse(w, fleet, http.StatusOK)

	case http.MethodDelete:
		// Delete fleet
		result := s.database.GetDB().Delete(&models.Fleet{}, fleetID)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to delete fleet %s", fleetID), result.Error)
			http.Error(w, "Failed to delete fleet", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "Fleet not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
