package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// RoomClient provides methods to communicate with the Room Service
type RoomClient struct {
	*BaseClient
}

// Room represents a room from the Room Service
type Room struct {
	ID                   string               `json:"id"`
	HotelID              string               `json:"hotel_id"`
	Name                 string               `json:"name"`
	Type                 string               `json:"type"`
	Price                float64              `json:"price"`
	Capacity             int                  `json:"capacity"`
	Description          string               `json:"description"`
	SpaceInfo            string               `json:"space_info"`
	BedDistribution      string               `json:"bed_distribution"`
	Quantity             int                  `json:"quantity"`
	AmenityCount         int                  `json:"amenity_count"`
	RecommendationCoef   float64              `json:"recommendation_coef"`
	HighlightedAmenities []HighlightedAmenity `json:"highlighted_amenities"`
	AmenityCategories    []AmenityCategory    `json:"amenity_categories"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

type HighlightedAmenity struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	Icon      string    `json:"icon"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AmenityCategory struct {
	ID           string    `json:"id"`
	RoomID       string    `json:"room_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Tier         string    `json:"tier"`
	AmenityCount int       `json:"amenity_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RoomListResponse is used to parse the paginated response
type RoomListResponse struct {
	Rooms      []Room `json:"rooms"`
	TotalCount int    `json:"total_count"`
}

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	HotelID              string               `json:"hotel_id" validate:"required"`
	Name                 string               `json:"name" validate:"required"`
	Type                 string               `json:"type" validate:"required"`
	Price                float64              `json:"price" validate:"required,gt=0"`
	Capacity             int                  `json:"capacity" validate:"required,gt=0"`
	Description          string               `json:"description" validate:"required"`
	SpaceInfo            string               `json:"space_info" validate:"required"`
	BedDistribution      string               `json:"bed_distribution" validate:"required"`
	Quantity             int                  `json:"quantity" validate:"required,gt=0"`
	HighlightedAmenities []HighlightedAmenity `json:"highlighted_amenities" validate:"omitempty,dive"`
	AmenityCategories    []AmenityCategory    `json:"amenity_categories" validate:"omitempty,dive"`
}

// UpdateRoomRequest represents the request to update a room
type UpdateRoomRequest struct {
	RoomNumber  string   `json:"room_number,omitempty"`
	Type        string   `json:"type,omitempty"`
	Description string   `json:"description,omitempty"`
	Price       float64  `json:"price,omitempty"`
	Capacity    int      `json:"capacity,omitempty"`
	Amenities   []string `json:"amenities,omitempty"`
	IsAvailable *bool    `json:"is_available,omitempty"`
}

// RoomFilter represents filters for listing rooms
type RoomFilter struct {
	HotelID     string
	Type        string
	IsAvailable *bool
	MinPrice    float64
	MaxPrice    float64
	Capacity    int
}

// NewRoomClient creates a new client for the Room Service
func NewRoomClient(baseURL string, timeout time.Duration, logger *slog.Logger) *RoomClient {
	return &RoomClient{
		BaseClient: NewBaseClient(baseURL, timeout, logger),
	}
}

// GetRoom retrieves a room by ID
func (c *RoomClient) GetRoom(ctx context.Context, roomID string) (*Room, error) {
	path := fmt.Sprintf("/rooms/%s", roomID)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get room request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrRoomNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get room returned status %d", resp.StatusCode)
	}

	var room Room
	if err := json.NewDecoder(resp.Body).Decode(&room); err != nil {
		return nil, fmt.Errorf("decode room response: %w", err)
	}

	return &room, nil
}

// GetRooms retrieves rooms with optional filters
func (c *RoomClient) GetRooms(ctx context.Context, filter RoomFilter) ([]Room, error) {
	path := "/rooms"
	if filter.HotelID != "" {
		path = fmt.Sprintf("/hotels/%s/rooms", filter.HotelID)
	}

	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get rooms request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get rooms returned status %d", resp.StatusCode)
	}

	var result RoomListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode rooms response: %w", err)
	}

	return result.Rooms, nil
}

// GetRoomsByHotel retrieves all rooms for a specific hotel
func (c *RoomClient) GetRoomsByHotel(ctx context.Context, hotelID string) ([]Room, error) {
	path := fmt.Sprintf("/hotels/%s/rooms", hotelID)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get rooms by hotel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrHotelNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get rooms by hotel returned status %d", resp.StatusCode)
	}

	var result RoomListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode rooms response: %w", err)
	}

	return result.Rooms, nil
}

// CreateRoom creates a new room
// Calls POST /hotels/{hotel_id}/rooms endpoint in Room Service
func (c *RoomClient) CreateRoom(ctx context.Context, req *CreateRoomRequest) (*Room, error) {
	path := fmt.Sprintf("/hotels/%s/rooms", req.HotelID)
	resp, err := c.DoWithContext(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("create room request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create room returned status %d", resp.StatusCode)
	}

	// Room service returns an array of created rooms (one per quantity)
	var rooms []Room
	if err := json.NewDecoder(resp.Body).Decode(&rooms); err != nil {
		return nil, fmt.Errorf("decode room response: %w", err)
	}

	// Return the first room
	if len(rooms) > 0 {
		return &rooms[0], nil
	}
	return nil, fmt.Errorf("no rooms created")
}

// UpdateRoom updates an existing room
func (c *RoomClient) UpdateRoom(ctx context.Context, roomID string, req *UpdateRoomRequest) (*Room, error) {
	path := fmt.Sprintf("/rooms/%s", roomID)
	resp, err := c.DoWithContext(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("update room request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrRoomNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update room returned status %d", resp.StatusCode)
	}

	var room Room
	if err := json.NewDecoder(resp.Body).Decode(&room); err != nil {
		return nil, fmt.Errorf("decode room response: %w", err)
	}

	return &room, nil
}

// DeleteRoom deletes a room by ID
func (c *RoomClient) DeleteRoom(ctx context.Context, roomID string) error {
	path := fmt.Sprintf("/rooms/%s", roomID)
	resp, err := c.DoWithContext(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("delete room request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrRoomNotFound
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete room returned status %d", resp.StatusCode)
	}

	return nil
}

// CheckAvailability checks if a room is available for a date range
func (c *RoomClient) CheckAvailability(ctx context.Context, roomID string, checkIn, checkOut time.Time) (bool, error) {
	path := fmt.Sprintf("/rooms/%s/availability?check_in=%s&check_out=%s",
		roomID, checkIn.Format(time.RFC3339), checkOut.Format(time.RFC3339))

	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, fmt.Errorf("check availability request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, ErrRoomNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("check availability returned status %d", resp.StatusCode)
	}

	var result struct {
		Available bool `json:"available"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode availability response: %w", err)
	}

	return result.Available, nil
}

// Health checks if the Room Service is healthy
func (c *RoomClient) Health(ctx context.Context) error {
	resp, err := c.DoWithContext(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}
