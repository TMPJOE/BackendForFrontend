package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ReservationClient provides methods to communicate with the Reservation Service
type ReservationClient struct {
	*BaseClient
}

// Reservation represents a reservation from the Reservation Service
type Reservation struct {
	ID          string    `json:"id"`
	HotelID     string    `json:"hotel_id"`
	RoomID      string    `json:"room_id"`
	UserID      string    `json:"user_id"`
	GuestName   string    `json:"guest_name"`
	GuestEmail  string    `json:"guest_email"`
	GuestPhone  string    `json:"guest_phone"`
	CheckIn     time.Time `json:"check_in"`
	CheckOut    time.Time `json:"check_out"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"` // pending, confirmed, cancelled, completed
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateReservationRequest represents the request to create a reservation
type CreateReservationRequest struct {
	HotelID     string    `json:"hotel_id" validate:"required"`
	RoomID      string    `json:"room_id" validate:"required"`
	UserID      string    `json:"user_id" validate:"required"`
	GuestName   string    `json:"guest_name" validate:"required"`
	GuestEmail  string    `json:"guest_email" validate:"required,email"`
	GuestPhone  string    `json:"guest_phone"`
	CheckIn     time.Time `json:"check_in" validate:"required"`
	CheckOut    time.Time `json:"check_out" validate:"required"`
	TotalAmount float64   `json:"total_amount" validate:"required,gt=0"`
	Notes       string    `json:"notes"`
}

// UpdateReservationRequest represents the request to update a reservation
type UpdateReservationRequest struct {
	GuestName   string    `json:"guest_name,omitempty"`
	GuestEmail  string    `json:"guest_email,omitempty"`
	GuestPhone  string    `json:"guest_phone,omitempty"`
	CheckIn     time.Time `json:"check_in,omitempty"`
	CheckOut    time.Time `json:"check_out,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// UpdateReservationStatusRequest represents the request to update reservation status
type UpdateReservationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed cancelled completed"`
}

// ReservationFilter represents filters for listing reservations
type ReservationFilter struct {
	HotelID   string
	RoomID    string
	UserID    string
	Status    string
	CheckInFrom time.Time
	CheckInTo   time.Time
}

// NewReservationClient creates a new client for the Reservation Service
func NewReservationClient(baseURL string, timeout time.Duration, logger *slog.Logger) *ReservationClient {
	return &ReservationClient{
		BaseClient: NewBaseClient(baseURL, timeout, logger),
	}
}

// GetReservation retrieves a reservation by ID
func (c *ReservationClient) GetReservation(ctx context.Context, reservationID string) (*Reservation, error) {
	path := fmt.Sprintf("/reservations/%s", reservationID)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get reservation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get reservation returned status %d", resp.StatusCode)
	}

	var reservation Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, fmt.Errorf("decode reservation response: %w", err)
	}

	return &reservation, nil
}

// GetReservations retrieves reservations with optional filters
func (c *ReservationClient) GetReservations(ctx context.Context, filter ReservationFilter) ([]Reservation, error) {
	path := "/reservations"
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get reservations request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get reservations returned status %d", resp.StatusCode)
	}

	var reservations []Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
		return nil, fmt.Errorf("decode reservations response: %w", err)
	}

	return reservations, nil
}

// GetReservationsByUser retrieves all reservations for a specific user
func (c *ReservationClient) GetReservationsByUser(ctx context.Context, userID string) ([]Reservation, error) {
	path := fmt.Sprintf("/users/%s/reservations", userID)
	resp, err := c.DoWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get reservations by user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get reservations by user returned status %d", resp.StatusCode)
	}

	var reservations []Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
		return nil, fmt.Errorf("decode reservations response: %w", err)
	}

	return reservations, nil
}

// CreateReservation creates a new reservation
func (c *ReservationClient) CreateReservation(ctx context.Context, req *CreateReservationRequest) (*Reservation, error) {
	resp, err := c.DoWithContext(ctx, http.MethodPost, "/reservations", req)
	if err != nil {
		return nil, fmt.Errorf("create reservation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, ErrRoomNotAvailable
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create reservation returned status %d", resp.StatusCode)
	}

	var reservation Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, fmt.Errorf("decode reservation response: %w", err)
	}

	return &reservation, nil
}

// UpdateReservation updates an existing reservation
func (c *ReservationClient) UpdateReservation(ctx context.Context, reservationID string, req *UpdateReservationRequest) (*Reservation, error) {
	path := fmt.Sprintf("/reservations/%s", reservationID)
	resp, err := c.DoWithContext(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("update reservation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update reservation returned status %d", resp.StatusCode)
	}

	var reservation Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, fmt.Errorf("decode reservation response: %w", err)
	}

	return &reservation, nil
}

// UpdateReservationStatus updates the status of a reservation
func (c *ReservationClient) UpdateReservationStatus(ctx context.Context, reservationID string, status string) (*Reservation, error) {
	path := fmt.Sprintf("/reservations/%s/status", reservationID)
	req := UpdateReservationStatusRequest{Status: status}
	resp, err := c.DoWithContext(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("update reservation status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update reservation status returned status %d", resp.StatusCode)
	}

	var reservation Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, fmt.Errorf("decode reservation response: %w", err)
	}

	return &reservation, nil
}

// CancelReservation cancels a reservation
func (c *ReservationClient) CancelReservation(ctx context.Context, reservationID string) error {
	_, err := c.UpdateReservationStatus(ctx, reservationID, "cancelled")
	return err
}

// DeleteReservation deletes a reservation by ID
func (c *ReservationClient) DeleteReservation(ctx context.Context, reservationID string) error {
	path := fmt.Sprintf("/reservations/%s", reservationID)
	resp, err := c.DoWithContext(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("delete reservation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrReservationNotFound
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete reservation returned status %d", resp.StatusCode)
	}

	return nil
}

// Health checks if the Reservation Service is healthy
func (c *ReservationClient) Health(ctx context.Context) error {
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
