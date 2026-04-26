package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// PaymentClient handles communication with the Payment Service
type PaymentClient struct {
	BaseClient
}

type ProcessPaymentRequest struct {
	BookingID       string  `json:"booking_id"`
	Amount          float64 `json:"amount"`
	PaymentMethodID string  `json:"payment_method_id"`
}

type PaymentResponse struct {
	PaymentID string  `json:"payment_id"`
	BookingID string  `json:"booking_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

// NewPaymentClient creates a new PaymentClient
func NewPaymentClient(baseURL string, timeout time.Duration, logger *slog.Logger) *PaymentClient {
	return &PaymentClient{
		BaseClient: BaseClient{
			BaseURL: baseURL,
			HTTPClient: &http.Client{
				Timeout: timeout,
			},
			Logger: logger,
		},
	}
}

// ProcessPayment sends a request to the Payment Service to process a payment
func (c *PaymentClient) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*PaymentResponse, error) {
	url := fmt.Sprintf("%s/payments/process", c.BaseURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal process payment request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var result PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
