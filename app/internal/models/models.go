// Package models defines the domain models used across the BFF.
// These models represent the aggregated data that the BFF returns to the frontend.
package models

import (
	"time"
)

// Hotel represents hotel data from the Hotel Service
type Hotel struct {
	ID          string    `json:"id"`
	AdminID     string    `json:"admin_id,omitempty"`
	Name        string    `json:"name"`
	City        string    `json:"city"`
	Description string    `json:"description"`
	Rating      float64   `json:"rating"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Room represents room data from the Room Service
type Room struct {
	ID                   string               `json:"id"`
	HotelID              string               `json:"hotel_id"`
	Name                 string               `json:"name"`
	Type                 string               `json:"type"`
	Price                float64              `json:"price"`
	Capacity             int                  `json:"capacity"`
	Description          string               `json:"description"`
	SpaceInfo            string               `json:"space_info"`
	BedDistribution      string               `json:"bed_distribution"`
	AmenityCount         int                  `json:"amenity_count"`
	RecommendationCoef   float64              `json:"recommendation_coef"`
	HighlightedAmenities []HighlightedAmenity `json:"highlighted_amenities"`
	AmenityCategories    []AmenityCategory    `json:"amenity_categories"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

type HighlightedAmenity struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
}

type AmenityCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Booking represents booking/reservation data from the Booking Service.
// This is the canonical response model for all reservation-related endpoints.
type Booking struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	HotelID     string    `json:"hotel_id"`
	RoomID      string    `json:"room_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	GuestCount  int       `json:"guest_count"`
	TotalPrice  float64   `json:"total_price"`
	Status      string    `json:"status"` // "pending", "confirmed", "cancelled", "completed"
	GuestName   string    `json:"guest_name,omitempty"`
	GuestEmail  string    `json:"guest_email,omitempty"`
	GuestPhone  string    `json:"guest_phone,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BookingDetails is a composite model that includes full hotel and room data
// alongside the booking — returned by the /reservations/{id}/details endpoint.
type BookingDetails struct {
	Booking Booking `json:"booking"`
	Hotel   Hotel   `json:"hotel"`
	Room    Room    `json:"room"`
}

// HotelWithRooms is a composite model that includes hotel data and its rooms
type HotelWithRooms struct {
	Hotel Hotel  `json:"hotel"`
	Rooms []Room `json:"rooms"`
}

// CreateHotelRequest represents the request to create a hotel
type CreateHotelRequest struct {
	Name        string  `json:"name" validate:"required"`
	City        string  `json:"city" validate:"required"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	HotelID              string               `json:"hotel_id" validate:"required,uuid"`
	Name                 string               `json:"name" validate:"required"`
	Type                 string               `json:"type" validate:"required"`
	Price                float64              `json:"price" validate:"required,gt=0"`
	Capacity             int                  `json:"capacity" validate:"required,gt=0"`
	Description          string               `json:"description" validate:"required"`
	SpaceInfo            string               `json:"space_info" validate:"required"`
	BedDistribution      string               `json:"bed_distribution" validate:"required"`
	Quantity             int                  `json:"quantity" validate:"required,gt=0"`
	HighlightedAmenities []HighlightedAmenity `json:"highlighted_amenities" validate:"omitempty,dive"`
	AmenityCategories    []AmenityCategory    `json:"amenity_categories" validate:"omitempty,dive"`
}

// CreateBookingRequest is what the frontend sends to the BFF when creating a reservation.
// The BFF enriches this with user_id (from JWT) and total_price (room.Price × nights)
// before forwarding to the Booking Service.
type CreateBookingRequest struct {
	HotelID    string    `json:"hotel_id" validate:"required"`
	RoomID     string    `json:"room_id" validate:"required"`
	StartDate  time.Time `json:"start_date" validate:"required"`
	EndDate    time.Time `json:"end_date" validate:"required"`
	GuestCount int       `json:"guest_count" validate:"required,min=1"`
	// Guest contact fields — optional, stored alongside the booking
	GuestName  string `json:"guest_name,omitempty"`
	GuestEmail string `json:"guest_email,omitempty"`
	GuestPhone string `json:"guest_phone,omitempty"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status          string    `json:"status"`
	HotelsService   string    `json:"hotels_service"`
	RoomsService    string    `json:"rooms_service"`
	BookingsService string    `json:"bookings_service"`
	Timestamp       time.Time `json:"timestamp"`
}
