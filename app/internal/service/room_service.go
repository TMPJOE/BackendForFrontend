package service

import (
	"context"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// GetRoom retrieves a single room by ID
func (s *BFFService) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	room, err := s.roomClient.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Fetch hotel name for enrichment
	hotelName := ""
	if hotel, err := s.hotelClient.GetHotel(ctx, room.HotelID); err == nil {
		hotelName = hotel.Name
	}

	return mapRoomClientToModel(room, hotelName), nil
}

// GetRoomsByHotel retrieves all rooms for a specific hotel
func (s *BFFService) GetRoomsByHotel(ctx context.Context, hotelID string) ([]models.Room, error) {
	// Verify hotel exists
	if _, err := s.hotelClient.GetHotel(ctx, hotelID); err != nil {
		return nil, err
	}

	// Fetch hotel name for enrichment
	hotelName := ""
	if hotel, err := s.hotelClient.GetHotel(ctx, hotelID); err == nil {
		hotelName = hotel.Name
	}

	rooms, err := s.roomClient.GetRoomsByHotel(ctx, hotelID)
	if err != nil {
		return nil, err
	}

	result := make([]models.Room, len(rooms))
	for i, r := range rooms {
		result[i] = *mapRoomClientToModel(&r, hotelName)
	}
	return result, nil
}

// CreateRoom creates a new room after validating the hotel exists
func (s *BFFService) CreateRoom(ctx context.Context, req *models.CreateRoomRequest) (*models.Room, error) {
	// Validate hotel exists first
	hotel, err := s.hotelClient.GetHotel(ctx, req.HotelID)
	if err != nil {
		return nil, err
	}

	createReq := &client.CreateRoomRequest{
		HotelID:     req.HotelID,
		RoomNumber:  req.RoomNumber,
		Type:        req.Type,
		Description: req.Description,
		Price:       req.Price,
		Capacity:    req.Capacity,
		Amenities:   req.Amenities,
	}

	room, err := s.roomClient.CreateRoom(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return mapRoomClientToModel(room, hotel.Name), nil
}

// UpdateRoom updates an existing room
func (s *BFFService) UpdateRoom(ctx context.Context, roomID string, req *models.UpdateRoomRequest) (*models.Room, error) {
	// Verify room exists and get hotel info
	existingRoom, err := s.roomClient.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Fetch hotel name for enrichment
	hotelName := ""
	if hotel, err := s.hotelClient.GetHotel(ctx, existingRoom.HotelID); err == nil {
		hotelName = hotel.Name
	}

	updateReq := &client.UpdateRoomRequest{
		RoomNumber:  req.RoomNumber,
		Type:        req.Type,
		Description: req.Description,
		Price:       req.Price,
		Capacity:    req.Capacity,
		Amenities:   req.Amenities,
		IsAvailable: req.IsAvailable,
	}

	room, err := s.roomClient.UpdateRoom(ctx, roomID, updateReq)
	if err != nil {
		return nil, err
	}

	return mapRoomClientToModel(room, hotelName), nil
}

// DeleteRoom deletes a room by ID
func (s *BFFService) DeleteRoom(ctx context.Context, roomID string) error {
	// Verify room exists
	if _, err := s.roomClient.GetRoom(ctx, roomID); err != nil {
		return err
	}

	return s.roomClient.DeleteRoom(ctx, roomID)
}

// CheckAvailability checks if a room is available for a date range
func (s *BFFService) CheckAvailability(ctx context.Context, roomID string, checkIn, checkOut string) (bool, error) {
	// Parse dates and validate
	checkInTime, err := parseTime(checkIn)
	if err != nil {
		return false, err
	}

	checkOutTime, err := parseTime(checkOut)
	if err != nil {
		return false, err
	}

	// Validate check-out is after check-in
	if !checkOutTime.After(checkInTime) {
		return false, client.ErrCheckOutBeforeCheckIn
	}

	// Verify room exists
	if _, err := s.roomClient.GetRoom(ctx, roomID); err != nil {
		return false, err
	}

	return s.roomClient.CheckAvailability(ctx, roomID, checkInTime, checkOutTime)
}
