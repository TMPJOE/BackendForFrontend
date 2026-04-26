package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// NotificationRequest represents the request to send a notification
type NotificationRequest struct {
	Type           string                 `json:"type"`
	RecipientEmail string                 `json:"recipient_email"`
	RecipientName  string                 `json:"recipient_name"`
	Data           map[string]interface{} `json:"data"`
}

// NotificationResponse represents the response from the notification service
type NotificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// NotificationClient communicates with the Notification microservice
type NotificationClient struct {
	*BaseClient
}

// NewNotificationClient creates a new notification service client
func NewNotificationClient(baseURL string, timeout time.Duration, logger *slog.Logger) *NotificationClient {
	return &NotificationClient{
		BaseClient: NewBaseClient(baseURL, timeout, logger),
	}
}

// SendNotification sends a notification request to the notification service.
// This is designed to be called as fire-and-forget — errors are logged but never
// propagated to the caller so that notification failures don't break business flows.
func (c *NotificationClient) SendNotification(ctx context.Context, req *NotificationRequest) {
	resp, err := c.DoWithContext(ctx, http.MethodPost, "/notifications/send", req)
	if err != nil {
		c.Logger.Warn("notification send failed (fire-and-forget)",
			"type", req.Type,
			"recipient", req.RecipientEmail,
			"error", err,
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Logger.Warn("notification service returned non-200",
			"type", req.Type,
			"recipient", req.RecipientEmail,
			"status", resp.StatusCode,
		)
		return
	}

	var notifResp NotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&notifResp); err != nil {
		c.Logger.Warn("failed to decode notification response",
			"type", req.Type,
			"error", err,
		)
		return
	}

	c.Logger.Info("notification sent successfully",
		"type", req.Type,
		"recipient", req.RecipientEmail,
	)
}

// SendNotificationAsync sends a notification in a background goroutine.
// Use this from service methods to avoid blocking the main request flow.
func (c *NotificationClient) SendNotificationAsync(ctx context.Context, req *NotificationRequest) {
	// Create a detached context so the notification isn't cancelled when the
	// parent HTTP request completes
	detachedCtx := context.WithoutCancel(ctx)

	go func() {
		// Add a timeout for the notification call itself
		sendCtx, cancel := context.WithTimeout(detachedCtx, 10*time.Second)
		defer cancel()
		c.SendNotification(sendCtx, req)
	}()
}

// Health checks the notification service health
func (c *NotificationClient) Health(ctx context.Context) error {
	resp, err := c.DoWithContext(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return fmt.Errorf("notification service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service unhealthy: status %d", resp.StatusCode)
	}
	return nil
}
