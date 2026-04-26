package service

import (
	"context"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// GetHotels retrieves all hotels with optional filters
// PASSTHROUGH: Directly forwards to Hotel Service
func (s *BFFService) GetHotels(ctx context.Context, city, country string) ([]models.Hotel, error) {
	hotels, err := s.hotelClient.GetHotels(ctx, city, country)
	if err != nil {
		return nil, err
	}

	result := make([]models.Hotel, len(hotels))
	for i, h := range hotels {
		result[i] = *mapHotelClientToModel(&h)
	}
	return result, nil
}

// GetHotel retrieves a single hotel by ID
// PASSTHROUGH: Directly forwards to Hotel Service
func (s *BFFService) GetHotel(ctx context.Context, hotelID string) (*models.Hotel, error) {
	hotel, err := s.hotelClient.GetHotel(ctx, hotelID)
	if err != nil {
		return nil, err
	}
	return mapHotelClientToModel(hotel), nil
}

// GetHotelWithRooms retrieves a hotel with all its rooms
// AGGREGATION: Calls Hotel Service + Room Service, merges response
// This is a BFF value-add: frontend gets hotel + rooms in one call
func (s *BFFService) GetHotelWithRooms(ctx context.Context, hotelID string) (*models.HotelWithRooms, error) {
	// Fetch hotel from Hotel Service
	hotel, err := s.hotelClient.GetHotel(ctx, hotelID)
	if err != nil {
		return nil, err
	}

	// Fetch rooms from Room Service
	rooms, err := s.roomClient.GetRoomsByHotel(ctx, hotelID)
	if err != nil {
		// Don't fail if rooms can't be fetched, just return empty
		s.logger.Warn("failed to get rooms for hotel", "hotel_id", hotelID, "error", err)
		rooms = []client.Room{}
	}

	// Map rooms with hotel name
	mappedRooms := make([]models.Room, len(rooms))
	for i, r := range rooms {
		mappedRooms[i] = *mapRoomClientToModel(&r, hotel.Name)
	}

	return &models.HotelWithRooms{
		Hotel: *mapHotelClientToModel(hotel),
		Rooms: mappedRooms,
	}, nil
}
