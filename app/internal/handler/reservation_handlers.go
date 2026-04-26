package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
)

// GetReservations handles GET /reservations - simple passthrough for user's reservations
func (h *Handler) GetReservations(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user authentication required")
		return
	}

	reservations, err := h.s.GetReservations(r.Context(), userID)
	if err != nil {
		h.l.Error("failed to get reservations", "user_id", userID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve reservations")
		return
	}

	respondJSON(w, http.StatusOK, reservations)
}

// GetReservation handles GET /reservations/{reservationId} - simple passthrough
func (h *Handler) GetReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "reservationId")
	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "reservation ID is required")
		return
	}

	reservation, err := h.s.GetReservation(r.Context(), reservationID)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "reservation not found")
			return
		}
		h.l.Error("failed to get reservation", "reservation_id", reservationID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve reservation")
		return
	}

	respondJSON(w, http.StatusOK, reservation)
}

// GetReservationDetails handles GET /reservations/{reservationId}/details
// AGGREGATION: Returns reservation + hotel + room merged (calls 3 services)
// This is a BFF value-add: frontend gets everything in one call
func (h *Handler) GetReservationDetails(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "reservationId")
	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "reservation ID is required")
		return
	}

	details, err := h.s.GetReservationDetails(r.Context(), reservationID)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "reservation not found")
			return
		}
		h.l.Error("failed to get reservation details", "reservation_id", reservationID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve reservation details")
		return
	}

	respondJSON(w, http.StatusOK, details)
}

// CreateReservation handles POST /reservations
// BRIDGE: Validates hotel + room exist, calculates total, then creates reservation
// This is a BFF value-add: orchestrates complex validation across services
func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user authentication required")
		return
	}

	var req models.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := helper.Validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	// BRIDGE: Service validates hotel + room exist, calculates total, then creates reservation
	reservation, err := h.s.CreateReservation(r.Context(), userID, &req)
	if err != nil {
		switch err {
		case helper.ErrNotFound:
			respondError(w, http.StatusNotFound, "hotel or room not found")
		case helper.ErrBadRequest:
			respondError(w, http.StatusBadRequest, "invalid dates or room not available")
		default:
			h.l.Error("failed to create reservation", "error", err)
			respondError(w, http.StatusInternalServerError, "failed to create reservation")
		}
		return
	}

	respondJSON(w, http.StatusCreated, reservation)
}
