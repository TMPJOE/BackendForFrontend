// Package service provides the business logic layer for the BFF.
// It coordinates calls to downstream services and aggregates responses
// for the frontend consumption.
package service

import (
	"context"
	"fmt"
	"log/slog"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// Service defines the interface for BFF business logic
type Service interface {
	// Health checks
	Check(ctx context.Context) error

	// Hotel operations
	GetHotel(ctx context.Context, hotelID string) (*models.Hotel, error)
	GetHotels(ctx context.Context, city, country string) ([]models.Hotel, error)
	CreateHotel(ctx context.Context, req *models.CreateHotelRequest) (*models.Hotel, error)
	UpdateHotel(ctx context.Context, hotelID string, req *models.UpdateHotelRequest) (*models.Hotel, error)
	DeleteHotel(ctx context.Context, hotelID string) error

	// Room operations
	GetRoom(ctx context.Context, roomID string) (*models.Room, error)
	GetRoomsByHotel(ctx context.Context, hotelID string) ([]models.Room, error)
	CreateRoom(ctx context.Context, req *models.CreateRoomRequest) (*models.Room, error)
	UpdateRoom(ctx context.Context, roomID string, req *models.UpdateRoomRequest) (*models.Room, error)
	DeleteRoom(ctx context.Context, roomID string) error
	CheckAvailability(ctx context.Context, roomID string, checkIn, checkOut string) (bool, error)

	// Reservation operations
	GetReservation(ctx context.Context, reservationID string) (*models.Reservation, error)
	GetReservationDetails(ctx context.Context, reservationID string) (*models.ReservationDetails, error)
	GetReservationsByUser(ctx context.Context, userID string) ([]models.Reservation, error)
	CreateReservation(ctx context.Context, userID string, req *models.CreateReservationRequest) (*models.Reservation, error)
	UpdateReservation(ctx context.Context, reservationID string, req *models.UpdateReservationRequest) (*models.Reservation, error)
	CancelReservation(ctx context.Context, reservationID string) error
	DeleteReservation(ctx context.Context, reservationID string) error

	// Composite operations
	GetHotelWithRooms(ctx context.Context, hotelID string) (*models.HotelWithRooms, error)
}

// BFFService implements the Service interface
type BFFService struct {
	logger      *slog.Logger
	hotelClient *client.HotelClient
	roomClient  *client.RoomClient
	reservationClient *client.ReservationClient
}

// Config holds configuration for the service layer
type Config struct {
	HotelServiceURL       string
	RoomServiceURL        string
	ReservationServiceURL string
}

// New creates a new BFFService with the given dependencies
func New(
	logger *slog.Logger,
	hotelClient *client.HotelClient,
	roomClient *client.RoomClient,
	reservationClient *client.ReservationClient,
) Service {
	return &BFFService{
		logger:      logger,
		hotelClient: hotelClient,
		roomClient:  roomClient,
		reservationClient: reservationClient,
	}
}

// mapHotelClientToModel converts a client Hotel to a models Hotel
func mapHotelClientToModel(h *client.Hotel) *models.Hotel {
	if h == nil {
		return nil
	}
	return &models.Hotel{
		ID:          h.ID,
		Name:        h.Name,
		Description: h.Description,
		Address:     h.Address,
		City:        h.City,
		Country:     h.Country,
		Rating:      h.Rating,
		CreatedAt:   h.CreatedAt,
		UpdatedAt:   h.UpdatedAt,
	}
}

// mapRoomClientToModel converts a client Room to a models Room
func mapRoomClientToModel(r *client.Room, hotelName string) *models.Room {
	if r == nil {
		return nil
	}
	return &models.Room{
		ID:          r.ID,
		HotelID:     r.HotelID,
		HotelName:   hotelName,
		RoomNumber:  r.RoomNumber,
		Type:        r.Type,
		Description: r.Description,
		Price:       r.Price,
		Capacity:    r.Capacity,
		Amenities:   r.Amenities,
		IsAvailable: r.IsAvailable,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// mapReservationClientToModel converts a client Reservation to a models Reservation
func mapReservationClientToModel(r *client.Reservation, hotelName, roomNumber string) *models.Reservation {
	if r == nil {
		return nil
	}
	return &models.Reservation{
		ID:          r.ID,
		HotelID:     r.HotelID,
		HotelName:   hotelName,
		RoomID:      r.RoomID,
		RoomNumber:  roomNumber,
		UserID:      r.UserID,
		GuestName:   r.GuestName,
		GuestEmail:  r.GuestEmail,
		GuestPhone:  r.GuestPhone,
		CheckIn:     r.CheckIn,
		CheckOut:    r.CheckOut,
		TotalAmount: r.TotalAmount,
		Status:      r.Status,
		Notes:       r.Notes,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// Check performs a health check on all downstream services
func (s *BFFService) Check(ctx context.Context) error {
	var errors []error

	if err := s.hotelClient.Health(ctx); err != nil {
		errors = append(errors, fmt.Errorf("hotel service: %w", err))
	}

	if err := s.roomClient.Health(ctx); err != nil {
		errors = append(errors, fmt.Errorf("room service: %w", err))
	}

	if err := s.reservationClient.Health(ctx); err != nil {
		errors = append(errors, fmt.Errorf("reservation service: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("one or more services are unhealthy: %v", errors)
	}

	return nil
}
