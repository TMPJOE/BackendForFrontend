package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hotel.com/app/internal/client"
	"hotel.com/app/internal/config"
	"hotel.com/app/internal/handler"
	"hotel.com/app/internal/logging"
	"hotel.com/app/internal/service"
)

const (
	publicKeyPath = "/app/keys/public.pem"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Println("failed to load config:", err)
		os.Exit(1)
	}

	// Create logger
	l := logging.New(cfg.Logging.Level, cfg.Logging.Format)
	l.Info("BFF Service initiated")

	// Parse timeout duration
	timeout, err := time.ParseDuration(cfg.DownstreamServices.Timeout)
	if err != nil {
		l.Warn("invalid timeout duration, using default", "error", err)
		timeout = 30 * time.Second
	}

	// Initialize downstream service clients
	hotelClient := client.NewHotelClient(
		cfg.DownstreamServices.HotelServiceURL,
		timeout,
		l,
	)
	l.Info("Hotel Service client initialized", "url", cfg.DownstreamServices.HotelServiceURL)

	roomClient := client.NewRoomClient(
		cfg.DownstreamServices.RoomServiceURL,
		timeout,
		l,
	)
	l.Info("Room Service client initialized", "url", cfg.DownstreamServices.RoomServiceURL)

	reservationClient := client.NewReservationClient(
		cfg.DownstreamServices.ReservationServiceURL,
		timeout,
		l,
	)
	l.Info("Reservation Service client initialized", "url", cfg.DownstreamServices.ReservationServiceURL)

	bookingClient := client.NewBookingClient(
		cfg.DownstreamServices.BookingServiceURL,
		timeout,
		l,
	)
	l.Info("Booking Service client initialized", "url", cfg.DownstreamServices.BookingServiceURL)

	paymentClient := client.NewPaymentClient(
		cfg.DownstreamServices.PaymentServiceURL,
		timeout,
		l,
	)
	l.Info("Payment Service client initialized", "url", cfg.DownstreamServices.PaymentServiceURL)

	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		l.Error("JWT public key file not found", "path", publicKeyPath)
		os.Exit(1)
	}
	l.Info("JWT keys loaded successfully")

	// Create service layer
	svc := service.New(l, hotelClient, roomClient, reservationClient, bookingClient, paymentClient)
	l.Info("Service layer initialized")

	// JWT configuration
	jwtExpiration, _ := time.ParseDuration(cfg.JWT.Expiration)
	if jwtExpiration == 0 {
		jwtExpiration = 24 * time.Hour
	}

	jwtConfig := handler.JWTConfig{
		Issuer:     cfg.JWT.Issuer,
		Expiration: jwtExpiration,
	}
	jwtAuth := handler.NewJWTAuthenticator(jwtConfig, publicKeyPath)

	// Create HTTP handler
	h := handler.New(svc, l, jwtAuth)

	// Create rate limiter if enabled
	var rateLimiter *handler.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = handler.NewRateLimiter(
			cfg.RateLimit.RequestsPerSecond,
			cfg.RateLimit.Burst,
			true,
		)
		l.Info("Rate limiting enabled",
			"rps", cfg.RateLimit.RequestsPerSecond,
			"burst", cfg.RateLimit.Burst,
		)
	}

	// Create HTTP server
	mux := h.NewServerMux(rateLimiter)
	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	l.Info("BFF Server listening", "addr", srv.Addr)

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	l.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Error("Server forced to shutdown", "err", err)
	}

	l.Info("Server stopped")
}
