package service

import (
	"context"
	"time"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// GetReservation retrieves a single reservation by ID
// PASSTHROUGH: Directly forwards to Reservation Service (with enrichment)
func (s *BFFService) GetReservation(ctx context.Context, reservationID string) (*models.Reservation, error) {
	reservation, err := s.reservationClient.GetReservation(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	// Enrich with hotel and room names (best effort)
	hotelName := ""
	roomNumber := ""
	if hotel, err := s.hotelClient.GetHotel(ctx, reservation.HotelID); err == nil {
		hotelName = hotel.Name
	}
	if room, err := s.roomClient.GetRoom(ctx, reservation.RoomID); err == nil {
		roomNumber = room.RoomNumber
	}

	return mapReservationClientToModel(reservation, hotelName, roomNumber), nil
}

// GetReservationDetails retrieves a reservation with full hotel and room details
// AGGREGATION: Calls 3 services, merges response
// This is a BFF value-add: frontend gets everything in one call
func (s *BFFService) GetReservationDetails(ctx context.Context, reservationID string) (*models.ReservationDetails, error) {
	// Fetch reservation from Reservation Service
	reservation, err := s.reservationClient.GetReservation(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	// Fetch hotel details from Hotel Service
	hotel, err := s.hotelClient.GetHotel(ctx, reservation.HotelID)
	if err != nil {
		return nil, err
	}

	// Fetch room details from Room Service
	room, err := s.roomClient.GetRoom(ctx, reservation.RoomID)
	if err != nil {
		return nil, err
	}

	return &models.ReservationDetails{
		Reservation: *mapReservationClientToModel(reservation, hotel.Name, room.RoomNumber),
		Hotel:       *mapHotelClientToModel(hotel),
		Room:        *mapRoomClientToModel(room, hotel.Name),
	}, nil
}

// GetReservations retrieves all reservations for a user
// PASSTHROUGH: Directly forwards to Reservation Service (with enrichment)
func (s *BFFService) GetReservations(ctx context.Context, userID string) ([]models.Reservation, error) {
	reservations, err := s.reservationClient.GetReservationsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]models.Reservation, len(reservations))
	for i, r := range reservations {
		// Enrich with hotel and room names (best effort)
		hotelName := ""
		roomNumber := ""
		if hotel, err := s.hotelClient.GetHotel(ctx, r.HotelID); err == nil {
			hotelName = hotel.Name
		}
		if room, err := s.roomClient.GetRoom(ctx, r.RoomID); err == nil {
			roomNumber = room.RoomNumber
		}

		result[i] = *mapReservationClientToModel(&r, hotelName, roomNumber)
	}

	return result, nil
}

// CreateReservation creates a new reservation after validating hotel and room exist
// BRIDGE: Validates hotel + room exist, calculates total, then creates reservation
// This is a BFF value-add: orchestrates complex validation across services
func (s *BFFService) CreateReservation(ctx context.Context, userID string, req *models.CreateReservationRequest) (*models.Reservation, error) {
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

	// Parse and validate dates
	checkIn, err := parseTime(req.CheckIn.Format(time.RFC3339))
	if err != nil {
		return nil, client.ErrInvalidDates
	}

	checkOut, err := parseTime(req.CheckOut.Format(time.RFC3339))
	if err != nil {
		return nil, client.ErrInvalidDates
	}

	// Validate check-in is not in the past
	if checkIn.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, client.ErrPastCheckIn
	}

	// Validate check-out is after check-in
	if !checkOut.After(checkIn) {
		return nil, client.ErrCheckOutBeforeCheckIn
	}

	// Calculate total amount
	nights := calculateNights(checkIn, checkOut)
	totalAmount := calculateTotalAmount(room.Price, nights)

	// Forward to Reservation Service
	createReq := &client.CreateReservationRequest{
		HotelID:     req.HotelID,
		RoomID:      req.RoomID,
		UserID:      userID,
		GuestName:   req.GuestName,
		GuestEmail:  req.GuestEmail,
		GuestPhone:  req.GuestPhone,
		CheckIn:     checkIn,
		CheckOut:    checkOut,
		TotalAmount: totalAmount,
		Notes:       req.Notes,
	}

	reservation, err := s.reservationClient.CreateReservation(ctx, createReq)
	if err != nil {
		return nil, err
	}

	// Return reservation with enriched data
	return mapReservationClientToModel(reservation, hotel.Name, room.RoomNumber), nil
}
