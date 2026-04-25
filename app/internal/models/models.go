// Package models defines the domain models used across the BFF.
// These models represent the aggregated data that the BFF returns to the frontend.
package models

import (
	"time"
)

// Hotel represents hotel data from the Hotel Service
type Hotel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	Rating      float64   `json:"rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Room represents room data from the Room Service
type Room struct {
	ID          string    `json:"id"`
	HotelID     string    `json:"hotel_id"`
	HotelName   string    `json:"hotel_name,omitempty"` // populated by BFF
	RoomNumber  string    `json:"room_number"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Capacity    int       `json:"capacity"`
	Amenities   []string  `json:"amenities"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Reservation represents reservation data from the Reservation Service
type Reservation struct {
	ID          string    `json:"id"`
	HotelID     string    `json:"hotel_id"`
	HotelName   string    `json:"hotel_name,omitempty"` // populated by BFF
	RoomID      string    `json:"room_id"`
	RoomNumber  string    `json:"room_number,omitempty"` // populated by BFF
	UserID      string    `json:"user_id"`
	GuestName   string    `json:"guest_name"`
	GuestEmail  string    `json:"guest_email"`
	GuestPhone  string    `json:"guest_phone"`
	CheckIn     time.Time `json:"check_in"`
	CheckOut    time.Time `json:"check_out"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReservationDetails is a composite model that includes full hotel and room data
// This is what the BFF returns for GET /reservations/:id/details
type ReservationDetails struct {
	Reservation Reservation `json:"reservation"`
	Hotel       Hotel       `json:"hotel"`
	Room        Room        `json:"room"`
}

// HotelWithRooms is a composite model that includes hotel data and its rooms
// This is useful for hotel detail pages
type HotelWithRooms struct {
	Hotel Hotel  `json:"hotel"`
	Rooms []Room `json:"rooms"`
}

// CreateHotelRequest represents the request to create a hotel
type CreateHotelRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Address     string  `json:"address" validate:"required"`
	City        string  `json:"city" validate:"required"`
	Country     string  `json:"country" validate:"required"`
	Rating      float64 `json:"rating"`
}

// UpdateHotelRequest represents the request to update a hotel
type UpdateHotelRequest struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Address     string  `json:"address,omitempty"`
	City        string  `json:"city,omitempty"`
	Country     string  `json:"country,omitempty"`
	Rating      float64 `json:"rating,omitempty"`
}

// CreateRoomRequest represents the request to create a room
// BFF validates the hotel exists before creating the room
type CreateRoomRequest struct {
	HotelID     string   `json:"hotel_id" validate:"required"`
	RoomNumber  string   `json:"room_number" validate:"required"`
	Type        string   `json:"type" validate:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	Capacity    int      `json:"capacity" validate:"required,gt=0"`
	Amenities   []string `json:"amenities"`
}

// UpdateRoomRequest represents the request to update a room
type UpdateRoomRequest struct {
	RoomNumber  string   `json:"room_number,omitempty"`
	Type        string   `json:"type,omitempty"`
	Description string   `json:"description,omitempty"`
	Price       float64  `json:"price,omitempty"`
	Capacity    int      `json:"capacity,omitempty"`
	Amenities   []string `json:"amenities,omitempty"`
	IsAvailable *bool    `json:"is_available,omitempty"`
}

// CreateReservationRequest represents the request to create a reservation
// BFF validates the hotel and room exist before creating the reservation
type CreateReservationRequest struct {
	HotelID     string    `json:"hotel_id" validate:"required"`
	RoomID      string    `json:"room_id" validate:"required"`
	GuestName   string    `json:"guest_name" validate:"required"`
	GuestEmail  string    `json:"guest_email" validate:"required,email"`
	GuestPhone  string    `json:"guest_phone"`
	CheckIn     time.Time `json:"check_in" validate:"required"`
	CheckOut    time.Time `json:"check_out" validate:"required"`
	Notes       string    `json:"notes"`
}

// UpdateReservationRequest represents the request to update a reservation
type UpdateReservationRequest struct {
	GuestName   string    `json:"guest_name,omitempty"`
	GuestEmail  string    `json:"guest_email,omitempty"`
	GuestPhone  string    `json:"guest_phone,omitempty"`
	CheckIn     time.Time `json:"check_in,omitempty"`
	CheckOut    time.Time `json:"check_out,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// ReservationStatus represents valid reservation statuses
const (
	ReservationStatusPending   = "pending"
	ReservationStatusConfirmed = "confirmed"
	ReservationStatusCancelled = "cancelled"
	ReservationStatusCompleted = "completed"
)

// AvailabilityCheckRequest represents a request to check room availability
type AvailabilityCheckRequest struct {
	HotelID  string    `json:"hotel_id"`
	RoomID   string    `json:"room_id"`
	CheckIn  time.Time `json:"check_in" validate:"required"`
	CheckOut time.Time `json:"check_out" validate:"required"`
	Guests   int       `json:"guests" validate:"required,gt=0"`
}

// AvailabilityCheckResponse represents the response for room availability check
type AvailabilityCheckResponse struct {
	Available     bool    `json:"available"`
	HotelID       string  `json:"hotel_id,omitempty"`
	RoomID        string  `json:"room_id,omitempty"`
	EstimatedPrice float64 `json:"estimated_price,omitempty"`
	Nights        int     `json:"nights,omitempty"`
	Message       string  `json:"message,omitempty"`
}

// SearchHotelsRequest represents a request to search hotels
type SearchHotelsRequest struct {
	City     string    `json:"city"`
	Country  string    `json:"country"`
	CheckIn  time.Time `json:"check_in"`
	CheckOut time.Time `json:"check_out"`
	Guests   int       `json:"guests"`
	MinPrice float64   `json:"min_price"`
	MaxPrice float64   `json:"max_price"`
}

// SearchHotelsResponse represents a response for hotel search
type SearchHotelsResponse struct {
	Hotels []HotelWithRooms `json:"hotels"`
	Total  int              `json:"total"`
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
	Status               string `json:"status"`
	HotelsService        string `json:"hotels_service"`
	RoomsService         string `json:"rooms_service"`
	ReservationsService  string `json:"reservations_service"`
	Timestamp            time.Time `json:"timestamp"`
}
