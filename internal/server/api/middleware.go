package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/edgetainer/edgetainer/internal/shared/models"
)

// loggingMiddleware logs incoming requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s.logger.Info(fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr))

		next.ServeHTTP(w, r)

		s.logger.Debug(fmt.Sprintf("%s %s %s completed in %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start)))
	})
}

// authMiddleware handles authentication for API routes
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		// Create context with user
		ctx := context.WithValue(r.Context(), "user", user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
