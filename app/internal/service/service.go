// Package service provides the business logic layer for the BFF.
// It coordinates calls to downstream services and aggregates responses
// for the frontend consumption.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// Service defines the minimal interface for BFF business logic
// Only includes operations that aggregate data or bridge service calls
// The BFF is NOT a passthrough - it only adds value when orchestrating

type Service interface {
	// Health checks downstream services
	Check(ctx context.Context) error

	// HOTEL OPERATIONS
	// GetHotels - simple passthrough (frontend needs for navigation)
	GetHotels(ctx context.Context, city, country string) ([]models.Hotel, error)
	// GetHotel - simple passthrough
	GetHotel(ctx context.Context, hotelID string) (*models.Hotel, error)
	// GetHotelWithRooms - AGGREGATION: hotel + rooms from different services
	GetHotelWithRooms(ctx context.Context, hotelID string) (*models.HotelWithRooms, error)

	// ROOM OPERATIONS
	// GetRoom - simple passthrough (frontend needs for detail view)
	GetRoom(ctx context.Context, roomID string) (*models.Room, error)
	// CreateRoom - BRIDGE: validates hotel exists, then creates room
	CreateRoom(ctx context.Context, req *models.CreateRoomRequest) (*models.Room, error)

	// RESERVATION OPERATIONS
	// GetReservations - simple passthrough (user's own reservations)
	GetReservations(ctx context.Context, userID string) ([]models.Reservation, error)
	// GetReservation - simple passthrough
	GetReservation(ctx context.Context, reservationID string) (*models.Reservation, error)
	// GetReservationDetails - AGGREGATION: reservation + hotel + room merged
	GetReservationDetails(ctx context.Context, reservationID string) (*models.ReservationDetails, error)
	// CreateReservation - BRIDGE: validates hotel + room exist, calculates total, creates reservation
	CreateReservation(ctx context.Context, userID string, req *models.CreateReservationRequest) (*models.Reservation, error)
}

// BFFService implements the Service interface
type BFFService struct {
	logger            *slog.Logger
	hotelClient       *client.HotelClient
	roomClient        *client.RoomClient
	reservationClient *client.ReservationClient
}

// New creates a new BFFService with the given dependencies
func New(
	logger *slog.Logger,
	hotelClient *client.HotelClient,
	roomClient *client.RoomClient,
	reservationClient *client.ReservationClient,
) Service {
	return &BFFService{
		logger:            logger,
		hotelClient:       hotelClient,
		roomClient:        roomClient,
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

// calculateNights calculates the number of nights between check-in and check-out
func calculateNights(checkIn, checkOut time.Time) int {
	duration := checkOut.Sub(checkIn)
	nights := int(duration.Hours() / 24)
	if nights < 1 {
		return 1
	}
	return nights
}

// calculateTotalAmount calculates the total price for a stay
func calculateTotalAmount(pricePerNight float64, nights int) float64 {
	return pricePerNight * float64(nights)
}

// parseTime parses a time string in multiple formats
func parseTime(t string) (time.Time, error) {
	// Try RFC3339 first
	if parsed, err := time.Parse(time.RFC3339, t); err == nil {
		return parsed, nil
	}
	// Try date-only format
	if parsed, err := time.Parse("2006-01-02", t); err == nil {
		return parsed, nil
	}
	// Try datetime format
	if parsed, err := time.Parse("2006-01-02T15:04:05", t); err == nil {
		return parsed, nil
	}
	return time.Time{}, client.ErrInvalidDates
}

// Check performs a health check on all downstream services
func (s *BFFService) Check(ctx context.Context) error {
	var errs []error

	if err := s.hotelClient.Health(ctx); err != nil {
		errs = append(errs, fmt.Errorf("hotel service: %w", err))
	}

	if err := s.roomClient.Health(ctx); err != nil {
		errs = append(errs, fmt.Errorf("room service: %w", err))
	}

	if err := s.reservationClient.Health(ctx); err != nil {
		errs = append(errs, fmt.Errorf("reservation service: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("one or more services are unhealthy: %v", errs)
	}

	return nil
}
