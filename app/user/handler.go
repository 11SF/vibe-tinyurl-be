package user

import (
	"fmt"

	"github.com/11SF/tinyurl/configs"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	config     configs.Config
	authClient AuthenticationClient
}

func NewHandler(config configs.Config, authClient AuthenticationClient) *Handler {
	return &Handler{
		config:     config,
		authClient: authClient,
	}
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req CoreAuthServiceLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Email and password are required",
		})
	}

	ctx := c.Context()
	resp, err := h.authClient.Login(ctx, req.Email, req.Password, "")
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	fmt.Print(resp)

	return c.JSON(*resp)
}

func (h *Handler) ValidateToken(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Authorization header required",
		})
	}

	// Extract token from "Bearer TOKEN"
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid authorization header format",
		})
	}

	token := authHeader[len(bearerPrefix):]
	userInfo, err := h.authClient.VerifyToken(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(userInfo)
}
