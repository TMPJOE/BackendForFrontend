package service

import (
	"context"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
)

// GetHotel retrieves a single hotel by ID
func (s *BFFService) GetHotel(ctx context.Context, hotelID string) (*models.Hotel, error) {
	hotel, err := s.hotelClient.GetHotel(ctx, hotelID)
	if err != nil {
		return nil, err
	}
	return mapHotelClientToModel(hotel), nil
}

// GetHotels retrieves all hotels with optional filters
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

// CreateHotel creates a new hotel
func (s *BFFService) CreateHotel(ctx context.Context, req *models.CreateHotelRequest) (*models.Hotel, error) {
	createReq := &client.CreateHotelRequest{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		City:        req.City,
		Country:     req.Country,
		Rating:      req.Rating,
	}

	hotel, err := s.hotelClient.CreateHotel(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return mapHotelClientToModel(hotel), nil
}

// UpdateHotel updates an existing hotel
func (s *BFFService) UpdateHotel(ctx context.Context, hotelID string, req *models.UpdateHotelRequest) (*models.Hotel, error) {
	// Verify hotel exists
	if _, err := s.hotelClient.GetHotel(ctx, hotelID); err != nil {
		return nil, err
	}

	updateReq := &client.UpdateHotelRequest{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		City:        req.City,
		Country:     req.Country,
		Rating:      req.Rating,
	}

	hotel, err := s.hotelClient.UpdateHotel(ctx, hotelID, updateReq)
	if err != nil {
		return nil, err
	}

	return mapHotelClientToModel(hotel), nil
}

// DeleteHotel deletes a hotel by ID
func (s *BFFService) DeleteHotel(ctx context.Context, hotelID string) error {
	// Verify hotel exists
	if _, err := s.hotelClient.GetHotel(ctx, hotelID); err != nil {
		return err
	}

	return s.hotelClient.DeleteHotel(ctx, hotelID)
}

// GetHotelWithRooms retrieves a hotel with all its rooms
func (s *BFFService) GetHotelWithRooms(ctx context.Context, hotelID string) (*models.HotelWithRooms, error) {
	// Fetch hotel
	hotel, err := s.hotelClient.GetHotel(ctx, hotelID)
	if err != nil {
		return nil, err
	}

	// Fetch rooms for this hotel
	rooms, err := s.roomClient.GetRoomsByHotel(ctx, hotelID)
	if err != nil {
		s.logger.Warn("failed to get rooms for hotel", "hotel_id", hotelID, "error", err)
		// Continue without rooms - they might be empty or error is non-fatal
		rooms = []client.Room{}
	}

	// Map rooms to model
	mappedRooms := make([]models.Room, len(rooms))
	for i, r := range rooms {
		mappedRooms[i] = *mapRoomClientToModel(&r, hotel.Name)
	}

	return &models.HotelWithRooms{
		Hotel: *mapHotelClientToModel(hotel),
		Rooms: mappedRooms,
	}, nil
}
