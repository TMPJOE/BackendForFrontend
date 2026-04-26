package service

import (
	"context"
	"time"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// GetReservation retrieves a single booking by ID from the Booking Service.
func (s *BFFService) GetReservation(ctx context.Context, reservationID string) (*models.Booking, error) {
	booking, err := s.bookingClient.GetBooking(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	return mapBookingClientToModel(booking), nil
}

// GetReservationDetails retrieves a booking with full hotel and room details.
// AGGREGATION: calls BookingService + HotelService + RoomService, merges response.
func (s *BFFService) GetReservationDetails(ctx context.Context, reservationID string) (*models.BookingDetails, error) {
	// Fetch booking from Booking Service
	booking, err := s.bookingClient.GetBooking(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	// Fetch hotel details from Hotel Service
	hotel, err := s.hotelClient.GetHotel(ctx, booking.HotelID)
	if err != nil {
		return nil, err
	}

	// Fetch room details from Room Service
	room, err := s.roomClient.GetRoom(ctx, booking.RoomID)
	if err != nil {
		return nil, err
	}

	return &models.BookingDetails{
		Booking: *mapBookingClientToModel(booking),
		Hotel:   *mapHotelClientToModel(hotel),
		Room:    *mapRoomClientToModel(room, hotel.Name),
	}, nil
}

// GetReservations retrieves all bookings for a user from the Booking Service.
func (s *BFFService) GetReservations(ctx context.Context, userID string) ([]models.Booking, error) {
	resp, err := s.bookingClient.GetBookingsByUser(ctx, userID, 1, 100)
	if err != nil {
		return nil, err
	}

	result := make([]models.Booking, len(resp.Bookings))
	for i, b := range resp.Bookings {
		b := b // capture loop variable
		result[i] = *mapBookingClientToModel(&b)
	}
	return result, nil
}

// CreateReservation creates a booking after orchestrating validation across services.
// BRIDGE:
//  1. Validates the hotel exists (HotelService)
//  2. Validates the room exists and belongs to that hotel (RoomService)
//  3. Calculates total_price = room.Price × nights
//  4. Forwards the assembled CreateBookingRequest to the Booking Service
func (s *BFFService) CreateReservation(ctx context.Context, userID string, req *models.CreateBookingRequest) (*models.Booking, error) {
	// BRIDGE: Validate hotel exists
	hotel, err := s.hotelClient.GetHotel(ctx, req.HotelID)
	if err != nil {
		return nil, err
	}

	// BRIDGE: Validate room exists and belongs to the hotel
	room, err := s.roomClient.GetRoom(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	if room.HotelID != req.HotelID {
		return nil, client.ErrInvalidReservationData
	}

	// Validate check-in is not in the past
	if req.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, client.ErrPastCheckIn
	}

	// Validate check-out is after check-in
	if !req.EndDate.After(req.StartDate) {
		return nil, client.ErrCheckOutBeforeCheckIn
	}

	// Calculate total price: room price × nights
	nights := calculateNights(req.StartDate, req.EndDate)
	totalPrice := calculateTotalAmount(room.Price, nights)

	// Forward to Booking Service with enriched data
	createReq := &client.CreateBookingRequest{
		UserID:     userID,
		HotelID:    req.HotelID,
		RoomID:     req.RoomID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		GuestCount: req.GuestCount,
		TotalPrice: totalPrice,
		GuestName:  req.GuestName,
		GuestEmail: req.GuestEmail,
		GuestPhone: req.GuestPhone,
	}

	booking, err := s.bookingClient.CreateBooking(ctx, createReq)
	if err != nil {
		return nil, err
	}

	s.logger.Info("booking created via BFF",
		"booking_id", booking.ID,
		"user_id", userID,
		"hotel", hotel.Name,
		"nights", nights,
		"total_price", totalPrice,
	)

	return mapBookingClientToModel(booking), nil
}
