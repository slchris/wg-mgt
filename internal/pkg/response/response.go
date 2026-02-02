package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSON writes a JSON response.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Success writes a successful response.
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{Success: true, Data: data})
}

// Created writes a created response.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Response{Success: true, Data: data})
}

// NoContent writes a no content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes an error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, Response{Success: false, Error: message})
}

// BadRequest writes a bad request response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// Unauthorized writes an unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden writes a forbidden response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message)
}

// NotFound writes a not found response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

// InternalError writes an internal server error response.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message)
}
