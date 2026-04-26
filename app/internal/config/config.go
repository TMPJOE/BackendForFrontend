package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Server holds HTTP server configuration
type Server struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	IdleTimeout  string `yaml:"idle_timeout"`
}

// JWT holds JWT authentication configuration
type JWT struct {
	Issuer     string `yaml:"issuer"`
	Expiration string `yaml:"expiration"`
}

// DownstreamServices holds configuration for all downstream microservices
type DownstreamServices struct {
	HotelServiceURL        string `yaml:"hotel_service_url"`
	RoomServiceURL         string `yaml:"room_service_url"`
	BookingServiceURL      string `yaml:"booking_service_url"`
	MediaServiceURL        string `yaml:"media_service_url"`
	ReservationServiceURL  string `yaml:"reservation_service_url"`
	PaymentServiceURL      string `yaml:"payment_service_url"`
	NotificationServiceURL string `yaml:"notification_service_url"`
	Timeout                string `yaml:"timeout"`
}

// GetTimeout returns the parsed timeout duration
func (ds DownstreamServices) GetTimeout() time.Duration {
	d, err := time.ParseDuration(ds.Timeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// Health holds health check configuration
type Health struct {
	Path      string `yaml:"path"`
	ReadyPath string `yaml:"ready_path"`
}

// RateLimit holds rate limiting configuration
type RateLimit struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

// CircuitBreaker holds circuit breaker configuration
type CircuitBreaker struct {
	Enabled      bool   `yaml:"enabled"`
	MaxFailures  int    `yaml:"max_failures"`
	Timeout      string `yaml:"timeout"`
	ResetTimeout string `yaml:"reset_timeout"`
}

// Logging holds logging configuration
type Logging struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Config holds all application configuration
type Config struct {
	Server             Server             `yaml:"server"`
	JWT                JWT                `yaml:"jwt"`
	DownstreamServices DownstreamServices `yaml:"downstream_services"`
	RateLimit          RateLimit          `yaml:"rate_limit"`
	CircuitBreaker     CircuitBreaker     `yaml:"circuit_breaker"`
	Health             Health             `yaml:"health"`
	Logging            Logging            `yaml:"logging"`
}

// Load reads a YAML config file, expands environment variables, and returns the parsed config.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded := os.ExpandEnv(string(b))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Set defaults for optional fields
	if cfg.DownstreamServices.Timeout == "" {
		cfg.DownstreamServices.Timeout = "30s"
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "bff-service"
	}

	// Set defaults for downstream service URLs if not provided via env vars
	if cfg.DownstreamServices.HotelServiceURL == "" {
		cfg.DownstreamServices.HotelServiceURL = "http://localhost:8084"
	}
	if cfg.DownstreamServices.RoomServiceURL == "" {
		cfg.DownstreamServices.RoomServiceURL = "http://localhost:8085"
	}
	if cfg.DownstreamServices.BookingServiceURL == "" {
		cfg.DownstreamServices.BookingServiceURL = "http://localhost:8086"
	}
	if cfg.DownstreamServices.ReservationServiceURL == "" {
		cfg.DownstreamServices.ReservationServiceURL = "http://localhost:8086"
	}
	if cfg.DownstreamServices.MediaServiceURL == "" {
		cfg.DownstreamServices.MediaServiceURL = "http://localhost:8082"
	}
	if cfg.DownstreamServices.PaymentServiceURL == "" {
		cfg.DownstreamServices.PaymentServiceURL = "http://localhost:8088"
	}
	if cfg.DownstreamServices.NotificationServiceURL == "" {
		cfg.DownstreamServices.NotificationServiceURL = "http://localhost:8089"
	}

	return &cfg, nil
}
