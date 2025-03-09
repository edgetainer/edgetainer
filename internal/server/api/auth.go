package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/edgetainer/edgetainer/internal/shared/models"
)

// handleLogin handles the login endpoint
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would fetch the user from the database and validate the password
	var user models.User
	result := s.database.GetDB().Where("username = ?", loginRequest.Username).First(&user)
	if result.Error != nil {
		s.logger.Error("Failed to find user", result.Error)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password (this would use proper hashing in a real implementation)
	// For demo purposes, we're using a simple check
	if loginRequest.Password != "password" {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate a token
	token := generateAuthToken()

	// Store token in database
	apiToken := models.APIToken{
		UserID:      user.ID,
		Token:       token,
		Description: "Login token",
		ExpiresAt:   time.Now().AddDate(0, 0, 7), // 7 days expiration
	}

	if err := s.database.GetDB().Create(&apiToken).Error; err != nil {
		s.logger.Error("Failed to store token", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	}

	jsonResponse(w, response, http.StatusOK)
}

// generateAuthToken creates a new random token for authentication
func generateAuthToken() string {
	// In a real implementation, this would use a proper crypto library
	// For now, we'll just create a random string
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

// handleLogout handles the logout endpoint
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Remove 'Bearer ' prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Invalidate the token in the database
	if err := s.database.GetDB().Where("token = ?", token).Delete(&models.APIToken{}).Error; err != nil {
		s.logger.Error("Failed to invalidate token", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleGetCurrentUser handles the current user endpoint
func (s *Server) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Remove 'Bearer ' prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Find the token in the database
	var apiToken models.APIToken
	if err := s.database.GetDB().Where("token = ?", token).First(&apiToken).Error; err != nil {
		s.logger.Error("Invalid token", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if token is expired
	if apiToken.ExpiresAt.Before(time.Now()) {
		s.logger.Info("Token expired")
		http.Error(w, "Token expired", http.StatusUnauthorized)
		return
	}

	// Get the user from the database
	var user models.User
	if err := s.database.GetDB().First(&user, apiToken.UserID).Error; err != nil {
		s.logger.Error("Failed to find user for token", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return user without sensitive information
	userResponse := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	}

	jsonResponse(w, userResponse, http.StatusOK)
}
