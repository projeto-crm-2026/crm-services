package mercadopago

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.mercadopago.com"

type Client struct {
	accessToken string
	httpClient  *http.Client
}

func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("mercadopago API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

type AutoRecurring struct {
	Frequency         int     `json:"frequency"`
	FrequencyType     string  `json:"frequency_type"`
	TransactionAmount float64 `json:"transaction_amount"`
	CurrencyID        string  `json:"currency_id"`
}

type CreatePlanRequest struct {
	Reason        string        `json:"reason"`
	AutoRecurring AutoRecurring `json:"auto_recurring"`
	BackURL       string        `json:"back_url,omitempty"`
}

type PlanResponse struct {
	ID            string        `json:"id"`
	Status        string        `json:"status"`
	Reason        string        `json:"reason"`
	AutoRecurring AutoRecurring `json:"auto_recurring"`
	InitPoint     string        `json:"init_point"`
	DateCreated   string        `json:"date_created"`
}

func (c *Client) CreatePlan(ctx context.Context, req *CreatePlanRequest) (*PlanResponse, error) {
	data, err := c.do(ctx, http.MethodPost, "/preapproval_plan", req)
	if err != nil {
		return nil, err
	}

	var resp PlanResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan response: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetPlan(ctx context.Context, planID string) (*PlanResponse, error) {
	data, err := c.do(ctx, http.MethodGet, "/preapproval_plan/"+planID, nil)
	if err != nil {
		return nil, err
	}

	var resp PlanResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan response: %w", err)
	}

	return &resp, nil
}

type CreateSubscriptionRequest struct {
	PreapprovalPlanID string         `json:"preapproval_plan_id"`
	Reason            string         `json:"reason"`
	PayerEmail        string         `json:"payer_email"`
	CardTokenID       string         `json:"card_token_id"`
	AutoRecurring     *AutoRecurring `json:"auto_recurring,omitempty"`
	BackURL           string         `json:"back_url,omitempty"`
	Status            string         `json:"status"`
}

type SubscriptionResponse struct {
	ID                string  `json:"id"`
	PreapprovalPlanID string  `json:"preapproval_plan_id"`
	PayerID           int64   `json:"payer_id"`
	PayerEmail        string  `json:"payer_email"`
	Status            string  `json:"status"`
	Reason            string  `json:"reason"`
	DateCreated       string  `json:"date_created"`
	NextPaymentDate   *string `json:"next_payment_date"`
}

func (c *Client) CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	data, err := c.do(ctx, http.MethodPost, "/preapproval", req)
	if err != nil {
		return nil, err
	}

	var resp SubscriptionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetSubscription(ctx context.Context, subscriptionID string) (*SubscriptionResponse, error) {
	data, err := c.do(ctx, http.MethodGet, "/preapproval/"+subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	var resp SubscriptionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return &resp, nil
}

type UpdateSubscriptionRequest struct {
	Status string `json:"status,omitempty"`
}

func (c *Client) UpdateSubscription(ctx context.Context, subscriptionID string, req *UpdateSubscriptionRequest) (*SubscriptionResponse, error) {
	data, err := c.do(ctx, http.MethodPut, "/preapproval/"+subscriptionID, req)
	if err != nil {
		return nil, err
	}

	var resp SubscriptionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return &resp, nil
}

type PaymentResponse struct {
	ID                int64   `json:"id"`
	Status            string  `json:"status"`
	StatusDetail      string  `json:"status_detail"`
	TransactionAmount float64 `json:"transaction_amount"`
	CurrencyID        string  `json:"currency_id"`
	DateCreated       string  `json:"date_created"`
	DateApproved      *string `json:"date_approved"`
	PaymentMethodID   string  `json:"payment_method_id"`
	Description       string  `json:"description"`
}

func (c *Client) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	data, err := c.do(ctx, http.MethodGet, "/v1/payments/"+paymentID, nil)
	if err != nil {
		return nil, err
	}

	var resp PaymentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment response: %w", err)
	}

	return &resp, nil
}

type WebhookNotification struct {
	ID          int64  `json:"id"`
	LiveMode    bool   `json:"live_mode"`
	Type        string `json:"type"`
	Action      string `json:"action"`
	DateCreated string `json:"date_created"`
	Data        struct {
		ID string `json:"id"`
	} `json:"data"`
	UserID int64 `json:"user_id"`
}
