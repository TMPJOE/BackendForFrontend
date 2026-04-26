package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
)

// GetRoom handles GET /rooms/{roomId} - simple passthrough for detail view
func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	room, err := h.s.GetRoom(r.Context(), roomID)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "room not found")
			return
		}
		h.l.Error("failed to get room", "room_id", roomID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve room")
		return
	}

	respondJSON(w, http.StatusOK, room)
}

// CreateRoom handles POST /hotels/{hotelId}/rooms
// BRIDGE: Validates hotel exists, then creates room
// This prevents orphaned rooms and provides clear error messages
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	hotelID := chi.URLParam(r, "hotelId")
	if hotelID == "" {
		respondError(w, http.StatusBadRequest, "hotel ID is required")
		return
	}

	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Set hotel ID from URL path
	req.HotelID = hotelID

	// Validate request
	if err := helper.Validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	// BRIDGE: Service validates hotel exists, then creates room
	room, err := h.s.CreateRoom(r.Context(), &req)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "hotel not found")
			return
		}
		h.l.Error("failed to create room", "hotel_id", hotelID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create room")
		return
	}

	respondJSON(w, http.StatusCreated, room)
}
