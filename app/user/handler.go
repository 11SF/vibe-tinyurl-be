package user

import "github.com/11SF/tinyurl/configs"

type Handler struct {
	config     configs.Config
	authClient *AuthClient
}

func NewHandler(config configs.Config) *Handler {
	return &Handler{
		config:     config,
		authClient: NewAuthClient(config.AuthConfig.ServiceURL),
	}
}