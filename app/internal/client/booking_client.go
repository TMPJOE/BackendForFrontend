package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// BookingClient provides methods to communicate with the Booking Service
type BookingClient struct {
	*BaseClient
}

// Booking represents a booking from the Booking Service
type Booking struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	HotelID    string    `json:"hotel_id"`
	RoomID     string    `json:"room_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	GuestCount int       `json:"guest_count"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	GuestName  string    `json:"guest_name,omitempty"`
	GuestEmail string    `json:"guest_email,omitempty"`
	GuestPhone string    `json:"guest_phone,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BookingListResponse represents a paginated list of bookings
type BookingListResponse struct {
	Bookings   []Booking `json:"bookings"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

// CreateBookingRequest represents the request to create a booking
type CreateBookingRequest struct {
	UserID     string    `json:"user_id"`
	HotelID    string    `json:"hotel_id"`
	RoomID     string    `json:"room_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	GuestCount int       `json:"guest_count"`
	TotalPrice float64   `json:"total_price"`
	GuestName  string    `json:"guest_name,omitempty"`
	GuestEmail string    `json:"guest_email,omitempty"`
	GuestPhone string    `json:"guest_phone,omitempty"`
}

// UpdateBookingRequest represents the request to update a booking
type UpdateBookingRequest struct {
	GuestCount *int       `json:"guest_count,omitempty"`
	TotalPrice *float64   `json:"total_price,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Status     *string    `json:"status,omitempty"`
}

// UpdateBookingStatusRequest represents the request to update booking status
type UpdateBookingStatusRequest struct {
	Status string `json:"status"`
}

// UpdateBookingDatesRequest represents the request to update booking dates
type UpdateBookingDatesRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// AvailabilityResponse represents the response for room availability check
type AvailabilityResponse struct {
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Available bool   `json:"available"`
}

// NewBookingClient creates a new client for the Booking Service
func NewBookingClient(baseURL string, timeout time.Duration, logger *slog.Logger) *BookingClient {
	return &BookingClient{
		BaseClient: NewBaseClient(baseURL, timeout, logger),
	}
}

// GetBooking retrieves a booking by ID
func (c *BookingClient) GetBooking(ctx context.Context, bookingID string) (*Booking, error) {
	path := fmt.Sprintf("/bookings/%s", bookingID)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get booking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get booking returned status %d", resp.StatusCode)
	}

	var booking Booking
	if err := json.NewDecoder(resp.Body).Decode(&booking); err != nil {
		return nil, fmt.Errorf("decode booking response: %w", err)
	}

	return &booking, nil
}

// GetBookings retrieves bookings with optional filters
func (c *BookingClient) GetBookings(ctx context.Context, userID, hotelID, roomID, status string, page, pageSize int) (*BookingListResponse, error) {
	path := "/bookings"
	queryParams := ""
	if userID != "" {
		queryParams += fmt.Sprintf("?user_id=%s", userID)
	}
	if hotelID != "" {
		sep := "?"
		if queryParams != "" {
			sep = "&"
		}
		queryParams += fmt.Sprintf("%shotel_id=%s", sep, hotelID)
	}
	if roomID != "" {
		sep := "?"
		if queryParams != "" {
			sep = "&"
		}
		queryParams += fmt.Sprintf("%sroom_id=%s", sep, roomID)
	}
	if status != "" {
		sep := "?"
		if queryParams != "" {
			sep = "&"
		}
		queryParams += fmt.Sprintf("%sstatus=%s", sep, status)
	}
	
	sep := "?"
	if queryParams != "" {
		sep = "&"
	}
	queryParams += fmt.Sprintf("%spage=%d&page_size=%d", sep, page, pageSize)
	path += queryParams

	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get bookings request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get bookings returned status %d", resp.StatusCode)
	}

	var response BookingListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode bookings response: %w", err)
	}

	return &response, nil
}

// GetBookingsByUser retrieves all bookings for a specific user
func (c *BookingClient) GetBookingsByUser(ctx context.Context, userID string, page, pageSize int) (*BookingListResponse, error) {
	path := fmt.Sprintf("/users/%s/bookings?page=%d&page_size=%d", userID, page, pageSize)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get bookings by user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get bookings by user returned status %d", resp.StatusCode)
	}

	var response BookingListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode bookings response: %w", err)
	}

	return &response, nil
}

// GetBookingsByHotel retrieves all bookings for a specific hotel
func (c *BookingClient) GetBookingsByHotel(ctx context.Context, hotelID string, page, pageSize int) (*BookingListResponse, error) {
	path := fmt.Sprintf("/hotels/%s/bookings?page=%d&page_size=%d", hotelID, page, pageSize)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get bookings by hotel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get bookings by hotel returned status %d", resp.StatusCode)
	}

	var response BookingListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode bookings response: %w", err)
	}

	return &response, nil
}

// CreateBooking creates a new booking
func (c *BookingClient) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*Booking, error) {
	resp, err := c.DoWithContext(ctx, http.MethodPost, "/bookings", req)
	if err != nil {
		return nil, fmt.Errorf("create booking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, ErrRoomNotAvailable
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create booking returned status %d", resp.StatusCode)
	}

	var booking Booking
	if err := json.NewDecoder(resp.Body).Decode(&booking); err != nil {
		return nil, fmt.Errorf("decode booking response: %w", err)
	}

	return &booking, nil
}

// UpdateBooking updates an existing booking
func (c *BookingClient) UpdateBooking(ctx context.Context, bookingID string, req *UpdateBookingRequest) (*Booking, error) {
	path := fmt.Sprintf("/bookings/%s", bookingID)
	resp, err := c.DoWithContext(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("update booking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update booking returned status %d", resp.StatusCode)
	}

	var booking Booking
	if err := json.NewDecoder(resp.Body).Decode(&booking); err != nil {
		return nil, fmt.Errorf("decode booking response: %w", err)
	}

	return &booking, nil
}

// UpdateBookingStatus updates the status of a booking
func (c *BookingClient) UpdateBookingStatus(ctx context.Context, bookingID string, status string) (*Booking, error) {
	path := fmt.Sprintf("/bookings/%s/status", bookingID)
	req := UpdateBookingStatusRequest{Status: status}
	resp, err := c.DoWithContext(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("update booking status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update booking status returned status %d", resp.StatusCode)
	}

	var booking Booking
	if err := json.NewDecoder(resp.Body).Decode(&booking); err != nil {
		return nil, fmt.Errorf("decode booking response: %w", err)
	}

	return &booking, nil
}

// CancelBooking cancels a booking
func (c *BookingClient) CancelBooking(ctx context.Context, bookingID string) error {
	path := fmt.Sprintf("/bookings/%s/cancel", bookingID)
	resp, err := c.DoWithContext(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("cancel booking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel booking returned status %d", resp.StatusCode)
	}

	return nil
}

// ConfirmBooking confirms a booking
func (c *BookingClient) ConfirmBooking(ctx context.Context, bookingID string) error {
	path := fmt.Sprintf("/bookings/%s/confirm", bookingID)
	resp, err := c.DoWithContext(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("confirm booking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("confirm booking returned status %d", resp.StatusCode)
	}

	return nil
}

// CheckRoomAvailability checks if a room is available for the given dates
func (c *BookingClient) CheckRoomAvailability(ctx context.Context, roomID string, startDate, endDate time.Time) (bool, error) {
	path := fmt.Sprintf("/rooms/%s/availability?start_date=%s&end_date=%s",
		roomID,
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339))
	
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, fmt.Errorf("check availability request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("check availability returned status %d", resp.StatusCode)
	}

	var availability AvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&availability); err != nil {
		return false, fmt.Errorf("decode availability response: %w", err)
	}

	return availability.Available, nil
}

// DeleteBooking deletes a booking by ID
func (c *BookingClient) DeleteBooking(ctx context.Context, bookingID string) error {
	path := fmt.Sprintf("/bookings/%s", bookingID)
	resp, err := c.DoWithContext(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("delete booking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete booking returned status %d", resp.StatusCode)
	}

	return nil
}

// Health checks if the Booking Service is healthy
func (c *BookingClient) Health(ctx context.Context) error {
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
