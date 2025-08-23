package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AuthClient struct {
	baseURL string
	client  *http.Client
}

type CoreAuthServiceLoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	CustomAttr string `json:"customAttr"`
}

type CoreAuthServiceLoginResponse struct {
	Status  string                              `json:"status"`
	Message string                              `json:"message"`
	Data    CoreAuthServiceLoginResponseData    `json:"data"`
}

type CoreAuthServiceLoginResponseData struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	UserInfo     UserInfo `json:"user_info"`
}

type UserInfo struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	PictureProfile string `json:"picture_profile"`
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (ac *AuthClient) Login(req CoreAuthServiceLoginRequest) (*CoreAuthServiceLoginResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", ac.baseURL+"/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var loginResp CoreAuthServiceLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, loginResp.Message)
	}

	return &loginResp, nil
}

func (ac *AuthClient) ValidateToken(token string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", ac.baseURL+"/validate", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := ac.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed with status %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userInfo, nil
}