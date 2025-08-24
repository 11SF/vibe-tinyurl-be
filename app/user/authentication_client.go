package user

import (
	"context"
	"encoding/json"

	httpclient "github.com/11SF/go-common/http_client"
	"github.com/11SF/tinyurl/configs"
)

type authenticationClient struct {
	config configs.Config
	client httpclient.HTTPClient
}

type AuthenticationClient interface {
	Login(_ context.Context, email, password, customAttr string) (*CoreAuthServiceLoginResponse, error)
	RefreshToken(_ context.Context, refreshToken string) (*CoreAuthServiceRefreshTokenResponse, error)
	VerifyToken(_ context.Context, accessToken string) (*VerifyTokenResponse, error)
}

func NewAuthenticationClient(config configs.Config, client httpclient.HTTPClient) AuthenticationClient {
	return &authenticationClient{
		config: config,
		client: client,
	}
}

type CoreAuthServiceLoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	CustomAttr string `json:"customAttr"`
}

type CoreAuthServiceLoginResponse struct {
	httpclient.CommonResponse
	Data CoreAuthServiceLoginResponseData `json:"data"`
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

func (c *authenticationClient) Login(_ context.Context, email, password string, customAttr string) (*CoreAuthServiceLoginResponse, error) {
	request := CoreAuthServiceLoginRequest{
		Email:      email,
		Password:   password,
		CustomAttr: customAttr,
	}

	resByte, err := c.client.Post(c.config.AuthenticationSrevice.PathLogin, request, nil)
	if err != nil {
		return nil, err
	}

	response := &CoreAuthServiceLoginResponse{}
	err = json.Unmarshal(resByte, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type CoreAuthServiceRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type CoreAuthServiceRefreshTokenResponse struct {
	httpclient.CommonResponse
	Data CoreAuthServiceLoginResponseData `json:"data"`
}

type CoreAuthServiceRefreshTokenResponseData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (c *authenticationClient) RefreshToken(_ context.Context, refreshToken string) (*CoreAuthServiceRefreshTokenResponse, error) {
	request := CoreAuthServiceRefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	resByte, err := c.client.Post(c.config.AuthenticationSrevice.PathRefresh, request, nil)
	if err != nil {
		return nil, err
	}

	response := &CoreAuthServiceRefreshTokenResponse{}
	err = json.Unmarshal(resByte, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type VerifyTokenRequest struct {
	AccessToken string `json:"accessToken"`
}

type VerifyTokenResponse struct {
	httpclient.CommonResponse
	Data VerifyTokenResponseData `json:"data"`
}

type VerifyTokenResponseData struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	Audience  []string `json:"aud"`
	ExpiresAt int64    `json:"exp"`
	NotBefore int64    `json:"nbf"`
	IssuedAt  int64    `json:"iat"`
	Jti       string   `json:"jti"`
	CustAttr  string   `json:"custAttr"`
}

func (c *authenticationClient) VerifyToken(_ context.Context, accessToken string) (*VerifyTokenResponse, error) {
	request := VerifyTokenRequest{
		AccessToken: accessToken,
	}

	resByte, err := c.client.Post(c.config.AuthenticationSrevice.PathVerifyToken, request, nil)
	if err != nil {
		return nil, err
	}

	response := &VerifyTokenResponse{}
	err = json.Unmarshal(resByte, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
