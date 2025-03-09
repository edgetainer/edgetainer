package api

import (
	"encoding/json"
	"net/http"
)

// jsonResponse sends a JSON response
func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, log the error and send a 500 response
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
