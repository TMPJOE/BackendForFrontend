package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/helper"
)

// GetHotels handles GET /hotels - simple passthrough for navigation
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

// GetHotel handles GET /hotels/{hotelId} - simple passthrough for navigation
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

// GetHotelWithRooms handles GET /hotels/{hotelId}/details
// AGGREGATION: Returns hotel + all rooms (calls Hotel Service + Room Service)
// This is a BFF value-add: frontend gets everything in one call
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
