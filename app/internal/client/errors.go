package client

import "errors"

// Sentinel errors for downstream service communication
var (
	// Service errors
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrRequestFailed      = errors.New("request failed")
	ErrInvalidResponse    = errors.New("invalid response from service")

	// Hotel errors
	ErrHotelNotFound    = errors.New("hotel not found")
	ErrHotelExists      = errors.New("hotel already exists")
	ErrInvalidHotelData = errors.New("invalid hotel data")

	// Room errors
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomExists        = errors.New("room already exists")
	ErrRoomNotAvailable  = errors.New("room not available for selected dates")
	ErrInvalidRoomData   = errors.New("invalid room data")

	// Reservation errors
	ErrReservationNotFound     = errors.New("reservation not found")
	ErrReservationExists       = errors.New("reservation already exists")
	ErrInvalidReservationData  = errors.New("invalid reservation data")
	ErrInvalidDates            = errors.New("invalid check-in/check-out dates")
	ErrPastCheckIn             = errors.New("check-in date is in the past")
	ErrCheckOutBeforeCheckIn   = errors.New("check-out date must be after check-in date")
)
