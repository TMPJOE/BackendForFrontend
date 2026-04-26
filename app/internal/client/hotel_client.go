package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// HotelClient provides methods to communicate with the Hotel Service
type HotelClient struct {
	*BaseClient
}

// Hotel represents a hotel from the Hotel Service
type Hotel struct {
	ID          string    `json:"id" db:"id"`
	AdminID     string    `json:"admin_id" db:"admin_id"`
	Name        string    `json:"name" db:"name"`
	City        string    `json:"city" db:"city"`
	Description string    `json:"description" db:"description"`
	Rating      float64   `json:"rating" db:"rating"`
	Lat         float64   `json:"lat" db:"lat"`
	Lng         float64   `json:"lng" db:"lng"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateHotelRequest represents the request to create a hotel
type CreateHotelRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Address     string  `json:"address" validate:"required"`
	City        string  `json:"city" validate:"required"`
	Country     string  `json:"country" validate:"required"`
	Rating      float64 `json:"rating"`
}

// UpdateHotelRequest represents the request to update a hotel
type UpdateHotelRequest struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Address     string  `json:"address,omitempty"`
	City        string  `json:"city,omitempty"`
	Country     string  `json:"country,omitempty"`
	Rating      float64 `json:"rating,omitempty"`
}

// NewHotelClient creates a new client for the Hotel Service
func NewHotelClient(baseURL string, timeout time.Duration, logger *slog.Logger) *HotelClient {
	return &HotelClient{
		BaseClient: NewBaseClient(baseURL, timeout, logger),
	}
}

// GetHotel retrieves a hotel by ID
func (c *HotelClient) GetHotel(ctx context.Context, hotelID string) (*Hotel, error) {
	path := fmt.Sprintf("/hotels/%s", hotelID)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get hotel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrHotelNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get hotel returned status %d", resp.StatusCode)
	}

	var hotel Hotel
	if err := json.NewDecoder(resp.Body).Decode(&hotel); err != nil {
		return nil, fmt.Errorf("decode hotel response: %w", err)
	}

	return &hotel, nil
}

// GetHotels retrieves all hotels with optional filters
func (c *HotelClient) GetHotels(ctx context.Context, city, country string) ([]Hotel, error) {
	path := "/hotels"
	if city != "" || country != "" {
		path = fmt.Sprintf("/hotels?city=%s&country=%s", city, country)
	}

	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get hotels request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get hotels returned status %d", resp.StatusCode)
	}

	var hotels []Hotel
	if err := json.NewDecoder(resp.Body).Decode(&hotels); err != nil {
		return nil, fmt.Errorf("decode hotels response: %w", err)
	}

	return hotels, nil
}

// CreateHotel creates a new hotel
func (c *HotelClient) CreateHotel(ctx context.Context, req *CreateHotelRequest) (*Hotel, error) {
	resp, err := c.DoWithContext(ctx, http.MethodPost, "/hotels", req)
	if err != nil {
		return nil, fmt.Errorf("create hotel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create hotel returned status %d", resp.StatusCode)
	}

	var hotel Hotel
	if err := json.NewDecoder(resp.Body).Decode(&hotel); err != nil {
		return nil, fmt.Errorf("decode hotel response: %w", err)
	}

	return &hotel, nil
}

// UpdateHotel updates an existing hotel
func (c *HotelClient) UpdateHotel(ctx context.Context, hotelID string, req *UpdateHotelRequest) (*Hotel, error) {
	path := fmt.Sprintf("/hotels/%s", hotelID)
	resp, err := c.DoWithContext(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("update hotel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrHotelNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update hotel returned status %d", resp.StatusCode)
	}

	var hotel Hotel
	if err := json.NewDecoder(resp.Body).Decode(&hotel); err != nil {
		return nil, fmt.Errorf("decode hotel response: %w", err)
	}

	return &hotel, nil
}

// DeleteHotel deletes a hotel by ID
func (c *HotelClient) DeleteHotel(ctx context.Context, hotelID string) error {
	path := fmt.Sprintf("/hotels/%s", hotelID)
	resp, err := c.DoWithContext(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("delete hotel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrHotelNotFound
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete hotel returned status %d", resp.StatusCode)
	}

	return nil
}

// Health checks if the Hotel Service is healthy
func (c *HotelClient) Health(ctx context.Context) error {
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
