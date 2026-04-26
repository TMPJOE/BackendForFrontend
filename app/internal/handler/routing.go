package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
)

// NewServerMux creates and configures the HTTP router with minimal BFF routes
// Only exposes endpoints that aggregate data or bridge service calls
func (h *Handler) NewServerMux(rateLimiter *RateLimiter) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(httplog.RequestLogger(h.l, &httplog.Options{
		Level:         slog.LevelDebug,
		Schema:        httplog.SchemaOTEL,
		RecoverPanics: true,
	}))
	r.Use(SecureHeaders)
	r.Use(RequestID)
	r.Use(CORS)

	// Apply rate limiting if enabled
	if rateLimiter != nil {
		r.Use(RateLimitMiddleware(rateLimiter))
	}

	// Custom error handlers (JSON instead of default HTML)
	r.NotFound(h.notFoundHandler)
	r.MethodNotAllowed(h.methodNotAllowedHandler)

	// Public routes - no authentication required
	r.Group(func(r chi.Router) {
		r.Get("/health", h.healthCheck)
		r.Get("/ready", h.readinessCheck)
	})

	// Protected routes - require JWT authentication
	r.Group(func(r chi.Router) {
		r.Use(h.jwtAuth.Middleware())

		// HOTEL AGGREGATION ENDPOINTS
		// Get hotel with all its rooms (merged response)
		r.Get("/hotels/{hotelId}/details", h.GetHotelWithRooms)

		// Simple passthrough for hotel list/detail (frontend needs these for navigation)
		r.Get("/hotels", h.GetHotels)
		r.Get("/hotels/{hotelId}", h.GetHotel)

		// BRIDGE: Create room - BFF verifies hotel exists first, then forwards to Room Service
		r.Post("/hotels/{hotelId}/rooms", h.CreateRoom)

		// Simple passthrough for room detail (frontend needs this)
		r.Get("/rooms/{roomId}", h.GetRoom)

		// RESERVATION AGGREGATION ENDPOINTS
		// Get reservation with full hotel and room details (merged response)
		r.Get("/reservations/{reservationId}/details", h.GetReservationDetails)

		// Simple passthrough for user's reservations
		r.Get("/reservations", h.GetReservations)
		r.Get("/reservations/{reservationId}", h.GetReservation)

		// BRIDGE: Create reservation - BFF verifies hotel + room exist, calculates total, then forwards
		r.Post("/reservations", h.CreateReservation)
	})

	return r
}

// GetUserIDFromRequest extracts user ID from the authenticated request
func GetUserIDFromRequest(r *http.Request) string {
	return GetUserIDFromContext(r.Context())
}

// GetUserEmailFromRequest extracts user email from the authenticated request
func GetUserEmailFromRequest(r *http.Request) string {
	return GetUserEmailFromContext(r.Context())
}

// GetClaimsFromRequest extracts JWT claims from the authenticated request
func GetClaimsFromRequest(r *http.Request) *JWTClaims {
	return GetClaimsFromContext(r.Context())
}
