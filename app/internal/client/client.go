// Package client provides HTTP clients for communicating with downstream microservices.
// Each service has its own client with methods for the available endpoints.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Config holds configuration for all downstream service clients
type Config struct {
	HotelServiceURL      string
	RoomServiceURL       string
	ReservationServiceURL string
	Timeout              time.Duration
}

// ServiceClient defines the interface for downstream service communication
type ServiceClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// BaseClient provides common HTTP client functionality
type BaseClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     *slog.Logger
}

// NewBaseClient creates a new base HTTP client
func NewBaseClient(baseURL string, timeout time.Duration, logger *slog.Logger) *BaseClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &BaseClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Logger: logger,
	}
}

// Do executes an HTTP request
func (c *BaseClient) Do(method, path string, body interface{}) (*http.Response, error) {
	return c.DoWithContext(context.Background(), method, path, body)
}

// DoWithContext executes an HTTP request with context
func (c *BaseClient) DoWithContext(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	token := GetTokenFromContext(ctx)
	return c.DoWithAuth(ctx, method, path, body, token)
}

// DoWithAuth executes an HTTP request with context and authorization header
func (c *BaseClient) DoWithAuth(ctx context.Context, method, path string, body interface{}, authToken string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authorization header if token is provided
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	c.Logger.Debug("making downstream request",
		"method", method,
		"url", url,
	)

	return c.HTTPClient.Do(req)
}

// contextKey is the type for context keys
type contextKey string

// TokenContextKey is the key for storing auth token in context
const TokenContextKey contextKey = "auth_token"

// WithToken adds an auth token to the context
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, TokenContextKey, token)
}

// GetTokenFromContext extracts the auth token from context
func GetTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(TokenContextKey).(string)
	return token
}
