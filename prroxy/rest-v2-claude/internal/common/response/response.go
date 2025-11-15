// Package response provides HTTP response helper functions
// for consistent JSON responses across the API.
package response

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
)

// JSON writes a JSON response with the given status code and data.
// Sets Content-Type header to application/json.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding error but don't panic - headers already sent
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

// Error writes an error response with appropriate HTTP status code.
// If the error is an AppError, uses its status code and message.
// Otherwise, returns a generic 500 Internal Server Error.
func Error(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		JSON(w, appErr.Status, map[string]string{"error": appErr.Message})
		return
	}
	JSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
}

// Success writes a 200 OK response with the given data.
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created writes a 201 Created response with the given data.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
