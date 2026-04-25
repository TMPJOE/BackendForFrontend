package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
)

// GetReservations handles GET /reservations - list user's reservations
func (h *Handler) GetReservations(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user authentication required")
		return
	}

	reservations, err := h.s.GetReservationsByUser(r.Context(), userID)
	if err != nil {
		h.l.Error("failed to get reservations", "user_id", userID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve reservations")
		return
	}

	respondJSON(w, http.StatusOK, reservations)
}

// GetReservation handles GET /reservations/{reservationId} - get a single reservation
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

// GetReservationDetails handles GET /reservations/{reservationId}/details - get full reservation details
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

// CreateReservation handles POST /reservations - create a new reservation
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

// UpdateReservation handles PUT /reservations/{reservationId} - update an existing reservation
func (h *Handler) UpdateReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "reservationId")
	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "reservation ID is required")
		return
	}

	var req models.UpdateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reservation, err := h.s.UpdateReservation(r.Context(), reservationID, &req)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "reservation not found")
			return
		}
		h.l.Error("failed to update reservation", "reservation_id", reservationID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to update reservation")
		return
	}

	respondJSON(w, http.StatusOK, reservation)
}

// CancelReservation handles PATCH /reservations/{reservationId}/cancel - cancel a reservation
func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "reservationId")
	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "reservation ID is required")
		return
	}

	if err := h.s.CancelReservation(r.Context(), reservationID); err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "reservation not found")
			return
		}
		h.l.Error("failed to cancel reservation", "reservation_id", reservationID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to cancel reservation")
		return
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Message: "reservation cancelled successfully",
	})
}

// DeleteReservation handles DELETE /reservations/{reservationId} - delete a reservation
func (h *Handler) DeleteReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "reservationId")
	if reservationID == "" {
		respondError(w, http.StatusBadRequest, "reservation ID is required")
		return
	}

	if err := h.s.DeleteReservation(r.Context(), reservationID); err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "reservation not found")
			return
		}
		h.l.Error("failed to delete reservation", "reservation_id", reservationID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to delete reservation")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
