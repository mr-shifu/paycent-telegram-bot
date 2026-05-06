package client

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type PaycentClient struct {
	http    *resty.Client
	baseURL string
}

func NewPaycentClient(baseURL, botAPIKey string) *PaycentClient {
	r := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Bot-API-Key", botAPIKey)

	return &PaycentClient{http: r, baseURL: baseURL}
}

// --- Auth ---

type VerifyPhoneRequest struct {
	Phone string `json:"phone"`
}

type VerifyPhoneResponse struct {
	Message string `json:"message"`
}

// RequestOTP asks Paycent to send an SMS verification code to the phone.
func (c *PaycentClient) RequestOTP(phone string) error {
	var errResp map[string]any
	resp, err := c.http.R().
		SetBody(VerifyPhoneRequest{Phone: phone}).
		SetError(&errResp).
		Post("/api/v1/auth/telegram/verify")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("request OTP failed: %v", errResp)
	}
	return nil
}

type LinkTelegramRequest struct {
	Phone          string `json:"phone"`
	OTP            string `json:"otp"`
	TelegramUserID int64  `json:"telegram_user_id"`
}

type LinkTelegramResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// VerifyOTP confirms the OTP and links the Telegram user to a Paycent account.
// Returns a JWT for subsequent API calls.
func (c *PaycentClient) VerifyOTP(phone, otp string, telegramUserID int64) (*LinkTelegramResponse, error) {
	var result LinkTelegramResponse
	var errResp map[string]any
	resp, err := c.http.R().
		SetBody(LinkTelegramRequest{Phone: phone, OTP: otp, TelegramUserID: telegramUserID}).
		SetResult(&result).
		SetError(&errResp).
		Post("/api/v1/auth/telegram/link")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("verify OTP failed: %v", errResp)
	}
	return &result, nil
}

// --- Groups ---

type Group struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type CreateGroupRequest struct {
	Name           string `json:"name"`
	TelegramChatID int64  `json:"telegram_chat_id"`
}

func (c *PaycentClient) WithJWT(jwt string) *resty.Client {
	return c.http.SetAuthToken(jwt)
}

func (c *PaycentClient) CreateGroup(jwt, name string, telegramChatID int64) (*Group, error) {
	var result Group
	var errResp map[string]any
	resp, err := c.http.R().
		SetAuthToken(jwt).
		SetBody(CreateGroupRequest{Name: name, TelegramChatID: telegramChatID}).
		SetResult(&result).
		SetError(&errResp).
		Post("/api/v1/group/groups")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("create group failed: %v", errResp)
	}
	return &result, nil
}

type LinkGroupRequest struct {
	TelegramChatID int64 `json:"telegram_chat_id"`
}

func (c *PaycentClient) LinkGroup(jwt, groupID string, telegramChatID int64) error {
	var errResp map[string]any
	resp, err := c.http.R().
		SetAuthToken(jwt).
		SetBody(LinkGroupRequest{TelegramChatID: telegramChatID}).
		SetError(&errResp).
		Post(fmt.Sprintf("/api/v1/group/groups/%s/telegram/link", groupID))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("link group failed: %v", errResp)
	}
	return nil
}

// --- Expenses ---

type AddExpenseRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	PaidBy      string  `json:"paid_by"`
}

type Expense struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

func (c *PaycentClient) AddExpense(jwt, groupID, description string, amount float64) (*Expense, error) {
	var result Expense
	var errResp map[string]any
	resp, err := c.http.R().
		SetAuthToken(jwt).
		SetBody(AddExpenseRequest{Description: description, Amount: amount}).
		SetResult(&result).
		SetError(&errResp).
		Post(fmt.Sprintf("/api/v1/group/groups/%s/expenses", groupID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("add expense failed: %v", errResp)
	}
	return &result, nil
}

// --- Balances ---

type Balance struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
}

func (c *PaycentClient) GetBalances(jwt, groupID string) ([]Balance, error) {
	var result []Balance
	var errResp map[string]any
	resp, err := c.http.R().
		SetAuthToken(jwt).
		SetResult(&result).
		SetError(&errResp).
		Get(fmt.Sprintf("/api/v1/group/groups/%s/balances", groupID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("get balances failed: %v", errResp)
	}
	return result, nil
}

// --- Payments ---

type CreatePaymentRequest struct {
	PaidTo string  `json:"paid_to"`
	Amount float64 `json:"amount"`
}

type Payment struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

func (c *PaycentClient) CreatePayment(jwt, groupID, paidToUserID string, amount float64) (*Payment, error) {
	var result Payment
	var errResp map[string]any
	resp, err := c.http.R().
		SetAuthToken(jwt).
		SetBody(CreatePaymentRequest{PaidTo: paidToUserID, Amount: amount}).
		SetResult(&result).
		SetError(&errResp).
		Post(fmt.Sprintf("/api/v1/group/groups/%s/payments", groupID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("create payment failed: %v", errResp)
	}
	return &result, nil
}

// --- Activity ---

type ActivityItem struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func (c *PaycentClient) GetActivity(jwt, groupID string) ([]ActivityItem, error) {
	var result []ActivityItem
	var errResp map[string]any
	resp, err := c.http.R().
		SetAuthToken(jwt).
		SetResult(&result).
		SetError(&errResp).
		Get(fmt.Sprintf("/api/v1/group/groups/%s/activity", groupID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("get activity failed: %v", errResp)
	}
	return result, nil
}
