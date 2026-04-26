// Package service provides the business logic layer for the BFF.
// It coordinates calls to downstream services and aggregates responses
// for the frontend consumption.
package service

import (
	"context"
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

	// BOOKING / RESERVATION OPERATIONS
	// GetReservations - simple passthrough (user's own bookings)
	GetReservations(ctx context.Context, userID string) ([]models.Booking, error)
	// GetReservation - simple passthrough
	GetReservation(ctx context.Context, reservationID string) (*models.Booking, error)
	// GetReservationDetails - AGGREGATION: booking + hotel + room merged
	GetReservationDetails(ctx context.Context, reservationID string) (*models.BookingDetails, error)
	// CreateReservation - BRIDGE: validates hotel + room exist, calculates total, creates booking
	CreateReservation(ctx context.Context, userID string, req *models.CreateBookingRequest) (*models.Booking, error)
}

// BFFService implements the Service interface
type BFFService struct {
	logger            *slog.Logger
	hotelClient       *client.HotelClient
	roomClient        *client.RoomClient
	reservationClient *client.ReservationClient
	bookingClient     *client.BookingClient
}

// New creates a new BFFService with the given dependencies
func New(
	logger *slog.Logger,
	hotelClient *client.HotelClient,
	roomClient *client.RoomClient,
	reservationClient *client.ReservationClient,
	bookingClient *client.BookingClient,
) Service {
	return &BFFService{
		logger:            logger,
		hotelClient:       hotelClient,
		roomClient:        roomClient,
		reservationClient: reservationClient,
		bookingClient:     bookingClient,
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
		City:        h.City,
		Rating:      h.Rating,
		Lat:         h.Lat,
		Lng:         h.Lng,
		CreatedAt:   h.CreatedAt,
		UpdatedAt:   h.UpdatedAt,
	}
}

// mapRoomClientToModel converts a client Room to a models Room
func mapRoomClientToModel(r *client.Room, _ string) *models.Room {
	if r == nil {
		return nil
	}
	return &models.Room{
		ID:          r.ID,
		HotelID:     r.HotelID,
		Type:        r.Type,
		Description: r.Description,
		Price:       r.Price,
		Capacity:    r.Capacity,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// mapBookingClientToModel converts a client Booking (BookingMicroService) to a models Booking.
func mapBookingClientToModel(b *client.Booking) *models.Booking {
	if b == nil {
		return nil
	}
	return &models.Booking{
		ID:         b.ID,
		UserID:     b.UserID,
		HotelID:    b.HotelID,
		RoomID:     b.RoomID,
		StartDate:  b.StartDate,
		EndDate:    b.EndDate,
		GuestCount: b.GuestCount,
		TotalPrice: b.TotalPrice,
		Status:     b.Status,
		GuestName:  b.GuestName,
		GuestEmail: b.GuestEmail,
		GuestPhone: b.GuestPhone,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
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
// Returns nil even if services are unhealthy - the BFF can still serve cached data
// and degraded functionality. Logs warnings for monitoring purposes.
func (s *BFFService) Check(ctx context.Context) error {
	var unhealthyServices []string

	if err := s.hotelClient.Health(ctx); err != nil {
		s.logger.Warn("hotel service health check failed", "error", err)
		unhealthyServices = append(unhealthyServices, "hotel")
	}

	if err := s.roomClient.Health(ctx); err != nil {
		s.logger.Warn("room service health check failed", "error", err)
		unhealthyServices = append(unhealthyServices, "room")
	}

	if err := s.bookingClient.Health(ctx); err != nil {
		s.logger.Warn("booking service health check failed", "error", err)
		unhealthyServices = append(unhealthyServices, "booking")
	}

	// BFF is considered ready even if downstream services are unhealthy
	// This allows the BFF to serve cached data and provide graceful degradation
	if len(unhealthyServices) > 0 {
		s.logger.Warn("bff is ready but some downstream services are unhealthy",
			"unhealthy_services", unhealthyServices)
	}

	return nil
}
