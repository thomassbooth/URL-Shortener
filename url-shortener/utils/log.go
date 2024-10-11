package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ValidateRequest checks if the request method is valid.
func ValidateRequest(w http.ResponseWriter, r *http.Request, expectedMethod string) error {
	if r.Method != expectedMethod {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return fmt.Errorf("invalid request method: %s, expected: %s", r.Method, expectedMethod)
	}
	return nil
}

// DecodeJSON decodes a JSON request body into the provided structure.
func DecodeJSON(body io.ReadCloser, v interface{}) error {
	return json.NewDecoder(body).Decode(v)
}

// RespondWithJSON prepares a JSON response with the specified status code.
func RespondWithJSON(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Write the JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
