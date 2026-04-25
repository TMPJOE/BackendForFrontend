package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
)

// GetRoomsByHotel handles GET /hotels/{hotelId}/rooms - list rooms for a hotel
func (h *Handler) GetRoomsByHotel(w http.ResponseWriter, r *http.Request) {
	hotelID := chi.URLParam(r, "hotelId")
	if hotelID == "" {
		respondError(w, http.StatusBadRequest, "hotel ID is required")
		return
	}

	rooms, err := h.s.GetRoomsByHotel(r.Context(), hotelID)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "hotel not found")
			return
		}
		h.l.Error("failed to get rooms", "hotel_id", hotelID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve rooms")
		return
	}

	respondJSON(w, http.StatusOK, rooms)
}

// GetRoom handles GET /rooms/{roomId} - get a single room
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

// CreateRoom handles POST /hotels/{hotelId}/rooms - create a new room
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

// UpdateRoom handles PUT /rooms/{roomId} - update an existing room
func (h *Handler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	var req models.UpdateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, err := h.s.UpdateRoom(r.Context(), roomID, &req)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "room not found")
			return
		}
		h.l.Error("failed to update room", "room_id", roomID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to update room")
		return
	}

	respondJSON(w, http.StatusOK, room)
}

// DeleteRoom handles DELETE /rooms/{roomId} - delete a room
func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	if err := h.s.DeleteRoom(r.Context(), roomID); err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "room not found")
			return
		}
		h.l.Error("failed to delete room", "room_id", roomID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to delete room")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckAvailability handles GET /rooms/{roomId}/availability - check room availability
func (h *Handler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	checkIn := r.URL.Query().Get("check_in")
	checkOut := r.URL.Query().Get("check_out")

	if checkIn == "" || checkOut == "" {
		respondError(w, http.StatusBadRequest, "check_in and check_out query parameters are required")
		return
	}

	available, err := h.s.CheckAvailability(r.Context(), roomID, checkIn, checkOut)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "room not found")
			return
		}
		h.l.Error("failed to check availability", "room_id", roomID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to check availability")
		return
	}

	respondJSON(w, http.StatusOK, models.AvailabilityCheckResponse{
		Available: available,
		RoomID:    roomID,
	})
}
