package service

import (
	"context"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// GetRoom retrieves a single room by ID
// PASSTHROUGH: Directly forwards to Room Service (with hotel name enrichment)
func (s *BFFService) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	room, err := s.roomClient.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Enrich with hotel name (best effort)
	hotelName := ""
	if hotel, err := s.hotelClient.GetHotel(ctx, room.HotelID); err == nil {
		hotelName = hotel.Name
	}

	return mapRoomClientToModel(room, hotelName), nil
}

// CreateRoom creates a new room after validating the hotel exists.
// BRIDGE: Validates hotel exists (Hotel Service), then creates room (Room Service)
func (s *BFFService) CreateRoom(ctx context.Context, req *models.CreateRoomRequest) (*models.Room, error) {
	// BRIDGE: Validate hotel exists first
	hotel, err := s.hotelClient.GetHotel(ctx, req.HotelID)
	if err != nil {
		return nil, err
	}

	// Forward to Room Service
	createReq := &client.CreateRoomRequest{
		HotelID:     req.HotelID,
		Type:        req.Type,
		Description: req.Description,
		Price:       req.Price,
		Capacity:    req.Capacity,
	}

	room, err := s.roomClient.CreateRoom(ctx, createReq)
	if err != nil {
		return nil, err
	}

	// Return room with hotel name
	return mapRoomClientToModel(room, hotel.Name), nil
}
