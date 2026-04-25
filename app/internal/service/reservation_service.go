package service

import (
	"context"
	"time"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

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

// GetReservation retrieves a single reservation by ID
func (s *BFFService) GetReservation(ctx context.Context, reservationID string) (*models.Reservation, error) {
	reservation, err := s.reservationClient.GetReservation(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	// Enrich with hotel and room names
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
func (s *BFFService) GetReservationDetails(ctx context.Context, reservationID string) (*models.ReservationDetails, error) {
	// Fetch reservation
	reservation, err := s.reservationClient.GetReservation(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	// Fetch hotel details
	hotel, err := s.hotelClient.GetHotel(ctx, reservation.HotelID)
	if err != nil {
		return nil, err
	}

	// Fetch room details
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

// GetReservationsByUser retrieves all reservations for a specific user
func (s *BFFService) GetReservationsByUser(ctx context.Context, userID string) ([]models.Reservation, error) {
	reservations, err := s.reservationClient.GetReservationsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]models.Reservation, len(reservations))
	for i, r := range reservations {
		// Enrich with hotel and room names
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
func (s *BFFService) CreateReservation(ctx context.Context, userID string, req *models.CreateReservationRequest) (*models.Reservation, error) {
	// Validate hotel exists
	hotel, err := s.hotelClient.GetHotel(ctx, req.HotelID)
	if err != nil {
		return nil, err
	}

	// Validate room exists and belongs to the hotel
	room, err := s.roomClient.GetRoom(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	if room.HotelID != req.HotelID {
		return nil, client.ErrInvalidReservationData
	}

	// Validate dates
	checkInTime, err := parseTime(req.CheckIn.Format(time.RFC3339))
	if err != nil {
		return nil, client.ErrInvalidDates
	}

	checkOutTime, err := parseTime(req.CheckOut.Format(time.RFC3339))
	if err != nil {
		return nil, client.ErrInvalidDates
	}

	// Check check-in is not in the past
	if checkInTime.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, client.ErrPastCheckIn
	}

	// Check check-out is after check-in
	if !checkOutTime.After(checkInTime) {
		return nil, client.ErrCheckOutBeforeCheckIn
	}

	// Calculate total amount
	nights := calculateNights(checkInTime, checkOutTime)
	totalAmount := calculateTotalAmount(room.Price, nights)

	// Create the reservation
	createReq := &client.CreateReservationRequest{
		HotelID:     req.HotelID,
		RoomID:      req.RoomID,
		UserID:      userID,
		GuestName:   req.GuestName,
		GuestEmail:  req.GuestEmail,
		GuestPhone:  req.GuestPhone,
		CheckIn:     checkInTime,
		CheckOut:    checkOutTime,
		TotalAmount: totalAmount,
		Notes:       req.Notes,
	}

	reservation, err := s.reservationClient.CreateReservation(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return mapReservationClientToModel(reservation, hotel.Name, room.RoomNumber), nil
}

// UpdateReservation updates an existing reservation
func (s *BFFService) UpdateReservation(ctx context.Context, reservationID string, req *models.UpdateReservationRequest) (*models.Reservation, error) {
	// Verify reservation exists
	existingReservation, err := s.reservationClient.GetReservation(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	// Fetch hotel and room names for enrichment
	hotelName := ""
	roomNumber := ""
	if hotel, err := s.hotelClient.GetHotel(ctx, existingReservation.HotelID); err == nil {
		hotelName = hotel.Name
	}
	if room, err := s.roomClient.GetRoom(ctx, existingReservation.RoomID); err == nil {
		roomNumber = room.RoomNumber
	}

	// Build update request
	updateReq := &client.UpdateReservationRequest{
		GuestName:  req.GuestName,
		GuestEmail: req.GuestEmail,
		GuestPhone: req.GuestPhone,
		Notes:      req.Notes,
	}

	// Only update dates if both are provided
	if !req.CheckIn.IsZero() && !req.CheckOut.IsZero() {
		updateReq.CheckIn = req.CheckIn
		updateReq.CheckOut = req.CheckOut
	}

	reservation, err := s.reservationClient.UpdateReservation(ctx, reservationID, updateReq)
	if err != nil {
		return nil, err
	}

	return mapReservationClientToModel(reservation, hotelName, roomNumber), nil
}

// CancelReservation cancels a reservation
func (s *BFFService) CancelReservation(ctx context.Context, reservationID string) error {
	// Verify reservation exists
	if _, err := s.reservationClient.GetReservation(ctx, reservationID); err != nil {
		return err
	}

	return s.reservationClient.CancelReservation(ctx, reservationID)
}

// DeleteReservation deletes a reservation by ID
func (s *BFFService) DeleteReservation(ctx context.Context, reservationID string) error {
	// Verify reservation exists
	if _, err := s.reservationClient.GetReservation(ctx, reservationID); err != nil {
		return err
	}

	return s.reservationClient.DeleteReservation(ctx, reservationID)
}
