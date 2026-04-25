package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
)

// GetHotels handles GET /hotels - list all hotels with optional filters
func (h *Handler) GetHotels(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	country := r.URL.Query().Get("country")

	hotels, err := h.s.GetHotels(r.Context(), city, country)
	if err != nil {
		h.l.Error("failed to get hotels", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve hotels")
		return
	}

	respondJSON(w, http.StatusOK, hotels)
}

// GetHotel handles GET /hotels/{hotelId} - get a single hotel
func (h *Handler) GetHotel(w http.ResponseWriter, r *http.Request) {
	hotelID := chi.URLParam(r, "hotelId")
	if hotelID == "" {
		respondError(w, http.StatusBadRequest, "hotel ID is required")
		return
	}

	hotel, err := h.s.GetHotel(r.Context(), hotelID)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "hotel not found")
			return
		}
		h.l.Error("failed to get hotel", "hotel_id", hotelID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve hotel")
		return
	}

	respondJSON(w, http.StatusOK, hotel)
}

// GetHotelWithRooms handles GET /hotels/{hotelId}/details - get hotel with all rooms
func (h *Handler) GetHotelWithRooms(w http.ResponseWriter, r *http.Request) {
	hotelID := chi.URLParam(r, "hotelId")
	if hotelID == "" {
		respondError(w, http.StatusBadRequest, "hotel ID is required")
		return
	}

	hotelWithRooms, err := h.s.GetHotelWithRooms(r.Context(), hotelID)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "hotel not found")
			return
		}
		h.l.Error("failed to get hotel with rooms", "hotel_id", hotelID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to retrieve hotel details")
		return
	}

	respondJSON(w, http.StatusOK, hotelWithRooms)
}

// CreateHotel handles POST /hotels - create a new hotel
func (h *Handler) CreateHotel(w http.ResponseWriter, r *http.Request) {
	var req models.CreateHotelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := helper.Validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	hotel, err := h.s.CreateHotel(r.Context(), &req)
	if err != nil {
		h.l.Error("failed to create hotel", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create hotel")
		return
	}

	respondJSON(w, http.StatusCreated, hotel)
}

// UpdateHotel handles PUT /hotels/{hotelId} - update an existing hotel
func (h *Handler) UpdateHotel(w http.ResponseWriter, r *http.Request) {
	hotelID := chi.URLParam(r, "hotelId")
	if hotelID == "" {
		respondError(w, http.StatusBadRequest, "hotel ID is required")
		return
	}

	var req models.UpdateHotelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hotel, err := h.s.UpdateHotel(r.Context(), hotelID, &req)
	if err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "hotel not found")
			return
		}
		h.l.Error("failed to update hotel", "hotel_id", hotelID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to update hotel")
		return
	}

	respondJSON(w, http.StatusOK, hotel)
}

// DeleteHotel handles DELETE /hotels/{hotelId} - delete a hotel
func (h *Handler) DeleteHotel(w http.ResponseWriter, r *http.Request) {
	hotelID := chi.URLParam(r, "hotelId")
	if hotelID == "" {
		respondError(w, http.StatusBadRequest, "hotel ID is required")
		return
	}

	if err := h.s.DeleteHotel(r.Context(), hotelID); err != nil {
		if err == helper.ErrNotFound {
			respondError(w, http.StatusNotFound, "hotel not found")
			return
		}
		h.l.Error("failed to delete hotel", "hotel_id", hotelID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to delete hotel")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
