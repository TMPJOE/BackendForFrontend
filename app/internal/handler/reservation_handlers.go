package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/client"
	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
)

// GetReservations handles GET /reservations
// Returns all bookings for the authenticated user.
func (h *Handler) GetReservations(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user authentication required")
		return
	}

	bookings, err := h.s.GetReservations(r.Context(), userID)
	if err != nil {
		h.l.Error("failed to get reservations", "user_id", userID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve reservations")
		return
	}

	respondJSON(w, http.StatusOK, bookings)
}

// GetReservation handles GET /reservations/{reservationId}
// Returns a single booking by ID.
func (h *Handler) GetReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "reservationId")
	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "reservation ID is required")
		return
	}

	booking, err := h.s.GetReservation(r.Context(), reservationID)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "reservation not found")
			return
		}
		h.l.Error("failed to get reservation", "reservation_id", reservationID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve reservation")
		return
	}

	respondJSON(w, http.StatusOK, booking)
}

// GetReservationDetails handles GET /reservations/{reservationId}/details
// AGGREGATION: Returns booking + hotel + room merged (calls 3 services).
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
// BRIDGE: Validates hotel + room exist, calculates total price, then creates booking.
//
// Expected request body:
//
//	{
//	  "hotel_id":    "<uuid>",
//	  "room_id":     "<uuid>",
//	  "start_date":  "<RFC3339>",
//	  "end_date":    "<RFC3339>",
//	  "guest_count": <int>,
//	  "guest_name":  "<string>",   // optional
//	  "guest_email": "<string>",   // optional
//	  "guest_phone": "<string>"    // optional
//	}
//
// user_id is taken from the JWT token — the frontend never sends it.
// total_price is calculated by the BFF (room.Price × nights).
func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user authentication required")
		return
	}

	var req models.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if err := helper.Validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	// BRIDGE: service validates hotel + room, calculates price, creates booking
	booking, err := h.s.CreateReservation(r.Context(), userID, &req)
	if err != nil {
		switch err {
		case helper.ErrNotFound:
			respondError(w, http.StatusNotFound, "hotel or room not found")
		case client.ErrRoomNotAvailable:
			respondError(w, http.StatusConflict, "room not available for selected dates")
		case client.ErrPastCheckIn:
			respondError(w, http.StatusBadRequest, "check-in date cannot be in the past")
		case client.ErrCheckOutBeforeCheckIn:
			respondError(w, http.StatusBadRequest, "check-out must be after check-in")
		case client.ErrInvalidReservationData:
			respondError(w, http.StatusBadRequest, "room does not belong to specified hotel")
		default:
			h.l.Error("failed to create reservation", "error", err)
			respondError(w, http.StatusInternalServerError, "failed to create reservation")
		}
		return
	}

	respondJSON(w, http.StatusCreated, booking)
}
