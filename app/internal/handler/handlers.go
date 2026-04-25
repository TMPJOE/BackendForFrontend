// Package handler provides HTTP request handlers for the BFF.
// It handles incoming HTTP requests, delegates to the service layer,
// and returns JSON responses with appropriate status codes.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"hotel.com/app/internal/models"
	"hotel.com/app/internal/service"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	s       service.Service
	l       *slog.Logger
	jwtAuth *JWTAuthenticator
}

// New creates a new Handler with dependencies
func New(s service.Service, l *slog.Logger, jwtAuth *JWTAuthenticator) *Handler {
	return &Handler{
		s:       s,
		l:       l,
		jwtAuth: jwtAuth,
	}
}

// respondJSON sends a JSON response with the given status code and data
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends a JSON error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.ErrorResponse{
		Code:    http.StatusText(status),
		Message:   message,
	})
}

// healthCheck handles GET /health
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readinessCheck handles GET /ready - checks downstream services
func (h *Handler) readinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Check(r.Context()); err != nil {
		h.l.Warn("readiness check failed", "error", err)
		respondJSON(w, http.StatusServiceUnavailable, models.HealthResponse{
			Status: "not ready",
		})
		return
	}

	respondJSON(w, http.StatusOK, models.HealthResponse{
		Status: "ready",
	})
}

// notFoundHandler handles 404 errors
func (h *Handler) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusNotFound, "endpoint not found")
}

// methodNotAllowedHandler handles 405 errors
func (h *Handler) methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusMethodNotAllowed, "method not allowed")
}
