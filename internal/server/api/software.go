package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/edgetainer/edgetainer/internal/shared/models"
)

// handleSoftware handles the software endpoint
func (s *Server) handleSoftware(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List software
		var software []models.Software

		// Fetch software from the database
		result := s.database.GetDB().Find(&software)
		if result.Error != nil {
			s.logger.Error("Failed to fetch software", result.Error)
			http.Error(w, "Failed to fetch software", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, software, http.StatusOK)

	case http.MethodPost:
		// Create software
		var software models.Software

		if err := json.NewDecoder(r.Body).Decode(&software); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Validate the software
		if software.Name == "" {
			http.Error(w, "Software name is required", http.StatusBadRequest)
			return
		}

		if software.Source == "" {
			http.Error(w, "Source is required", http.StatusBadRequest)
			return
		}

		// Save to the database
		if err := s.database.GetDB().Create(&software).Error; err != nil {
			s.logger.Error("Failed to create software", err)
			http.Error(w, "Failed to create software", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, software, http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSoftwareByID handles the software by ID endpoint
func (s *Server) handleSoftwareByID(w http.ResponseWriter, r *http.Request) {
	// Extract software ID from URL
	softwareID := filepath.Base(r.URL.Path)

	s.logger.Info(fmt.Sprintf("Software operation on ID: %s", softwareID))

	switch r.Method {
	case http.MethodGet:
		// Get software by ID
		var software models.Software

		// Fetch the software from the database
		result := s.database.GetDB().First(&software, softwareID)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to fetch software %s", softwareID), result.Error)
			http.Error(w, "Software not found", http.StatusNotFound)
			return
		}

		jsonResponse(w, software, http.StatusOK)

	case http.MethodPut:
		// Update software
		var software models.Software

		if err := json.NewDecoder(r.Body).Decode(&software); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Validate the software
		if software.Name == "" {
			http.Error(w, "Software name is required", http.StatusBadRequest)
			return
		}

		// Update in the database
		result := s.database.GetDB().Model(&models.Software{}).Where("id = ?", softwareID).Updates(software)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to update software %s", softwareID), result.Error)
			http.Error(w, "Failed to update software", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "Software not found", http.StatusNotFound)
			return
		}

		// Fetch the updated software to return
		s.database.GetDB().First(&software, softwareID)

		jsonResponse(w, software, http.StatusOK)

	case http.MethodDelete:
		// Delete software
		result := s.database.GetDB().Delete(&models.Software{}, softwareID)
		if result.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to delete software %s", softwareID), result.Error)
			http.Error(w, "Failed to delete software", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "Software not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
