package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
)

// NewServerMux creates and configures the HTTP router with all BFF routes
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

		// Hotel routes
		r.Route("/hotels", func(r chi.Router) {
			r.Get("/", h.GetHotels)
			r.Post("/", h.CreateHotel)

			r.Route("/{hotelId}", func(r chi.Router) {
				r.Get("/", h.GetHotel)
				r.Put("/", h.UpdateHotel)
				r.Delete("/", h.DeleteHotel)
				r.Get("/details", h.GetHotelWithRooms)

				// Room routes nested under hotel
				r.Get("/rooms", h.GetRoomsByHotel)
				r.Post("/rooms", h.CreateRoom)
			})
		})

		// Room routes (direct access)
		r.Route("/rooms", func(r chi.Router) {
			r.Get("/{roomId}", h.GetRoom)
			r.Put("/{roomId}", h.UpdateRoom)
			r.Delete("/{roomId}", h.DeleteRoom)
			r.Get("/{roomId}/availability", h.CheckAvailability)
		})

		// Reservation routes
		r.Route("/reservations", func(r chi.Router) {
			r.Get("/", h.GetReservations)
			r.Post("/", h.CreateReservation)

			r.Route("/{reservationId}", func(r chi.Router) {
				r.Get("/", h.GetReservation)
				r.Put("/", h.UpdateReservation)
				r.Delete("/", h.DeleteReservation)
				r.Get("/details", h.GetReservationDetails)
				r.Patch("/cancel", h.CancelReservation)
			})
		})
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
