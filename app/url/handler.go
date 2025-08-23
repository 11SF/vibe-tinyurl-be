package url

import (
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/11SF/tinyurl/configs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	config  configs.Config
	service Service
}

type CreateURLRequest struct {
	URL         string  `json:"url" validate:"required,url"`
	CustomAlias *string `json:"custom_alias,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type CreateURLResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Data    *CreateURLData   `json:"data,omitempty"`
}

type CreateURLData struct {
	ID          string  `json:"id"`
	ShortCode   string  `json:"short_code"`
	ShortURL    string  `json:"short_url"`
	OriginalURL string  `json:"original_url"`
	CustomAlias *string `json:"custom_alias,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type GetURLsResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    []*URLData `json:"data"`
}

type URLData struct {
	ID          string  `json:"id"`
	ShortCode   string  `json:"short_code"`
	ShortURL    string  `json:"short_url"`
	OriginalURL string  `json:"original_url"`
	CustomAlias *string `json:"custom_alias,omitempty"`
	ClickCount  int64   `json:"click_count"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type UpdateURLRequest struct {
	ID          string  `json:"id" validate:"required,uuid"`
	URL         string  `json:"url" validate:"required,url"`
	CustomAlias *string `json:"custom_alias,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type AnalyticsResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    []*ClickData   `json:"data"`
}

type ClickData struct {
	ID        string  `json:"id"`
	IPAddress string  `json:"ip_address"`
	UserAgent string  `json:"user_agent"`
	Referer   *string `json:"referer,omitempty"`
	CreatedAt string  `json:"created_at"`
}

func NewHandler(config configs.Config, service Service) *Handler {
	return &Handler{
		config:  config,
		service: service,
	}
}

func (h *Handler) CreateURL(c *fiber.Ctx) error {
	var req CreateURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(CreateURLResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	var userID *uuid.UUID
	if userIDStr := c.Get("X-User-ID"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userID = &id
		}
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			expiresAt = &t
		} else {
			return c.Status(400).JSON(CreateURLResponse{
				Status:  "error",
				Message: "Invalid expires_at format, use RFC3339",
			})
		}
	}

	url, err := h.service.ShortenURL(req.URL, req.CustomAlias, userID, expiresAt)
	if err != nil {
		slog.Error("Failed to create URL", "error", err)
		return c.Status(400).JSON(CreateURLResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	shortURL := h.config.BaseURL + "/" + url.ShortCode
	
	var expiresAtStr *string
	if url.ExpiresAt != nil {
		str := url.ExpiresAt.Format(time.RFC3339)
		expiresAtStr = &str
	}

	return c.JSON(CreateURLResponse{
		Status:  "success",
		Message: "URL created successfully",
		Data: &CreateURLData{
			ID:          url.ID.String(),
			ShortCode:   url.ShortCode,
			ShortURL:    shortURL,
			OriginalURL: url.OriginalURL,
			CustomAlias: url.CustomAlias,
			ExpiresAt:   expiresAtStr,
			CreatedAt:   url.CreatedAt.Format(time.RFC3339),
		},
	})
}

func (h *Handler) RedirectURL(c *fiber.Ctx) error {
	shortCode := c.Params("shortCode")
	if shortCode == "" {
		return c.Status(404).SendString("Short code not found")
	}

	url, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		slog.Error("Failed to get original URL", "shortCode", shortCode, "error", err)
		return c.Status(404).SendString("URL not found or expired")
	}

	ipAddress := getClientIP(c)
	userAgent := c.Get("User-Agent")
	referer := c.Get("Referer")

	go func() {
		if err := h.service.TrackClick(url.ID, ipAddress, userAgent, referer); err != nil {
			slog.Error("Failed to track click", "urlID", url.ID, "error", err)
		}
	}()

	return c.Redirect(url.OriginalURL, 301)
}

func (h *Handler) GetUserURLs(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(401).JSON(GetURLsResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(400).JSON(GetURLsResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	urls, err := h.service.GetUserURLs(userID, limit, offset)
	if err != nil {
		slog.Error("Failed to get user URLs", "userID", userID, "error", err)
		return c.Status(500).JSON(GetURLsResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	var urlsData []*URLData
	for _, url := range urls {
		shortURL := h.config.BaseURL + "/" + url.ShortCode
		
		var expiresAtStr *string
		if url.ExpiresAt != nil {
			str := url.ExpiresAt.Format(time.RFC3339)
			expiresAtStr = &str
		}

		urlsData = append(urlsData, &URLData{
			ID:          url.ID.String(),
			ShortCode:   url.ShortCode,
			ShortURL:    shortURL,
			OriginalURL: url.OriginalURL,
			CustomAlias: url.CustomAlias,
			ClickCount:  url.ClickCount,
			ExpiresAt:   expiresAtStr,
			CreatedAt:   url.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   url.UpdatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(GetURLsResponse{
		Status:  "success",
		Message: "URLs retrieved successfully",
		Data:    urlsData,
	})
}

func (h *Handler) UpdateURL(c *fiber.Ctx) error {
	var req UpdateURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(CreateURLResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	urlID, err := uuid.Parse(req.ID)
	if err != nil {
		return c.Status(400).JSON(CreateURLResponse{
			Status:  "error",
			Message: "Invalid URL ID",
		})
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			expiresAt = &t
		} else {
			return c.Status(400).JSON(CreateURLResponse{
				Status:  "error",
				Message: "Invalid expires_at format, use RFC3339",
			})
		}
	}

	url, err := h.service.UpdateURL(urlID, req.URL, req.CustomAlias, expiresAt)
	if err != nil {
		slog.Error("Failed to update URL", "urlID", urlID, "error", err)
		return c.Status(400).JSON(CreateURLResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	shortURL := h.config.BaseURL + "/" + url.ShortCode
	
	var expiresAtStr *string
	if url.ExpiresAt != nil {
		str := url.ExpiresAt.Format(time.RFC3339)
		expiresAtStr = &str
	}

	return c.JSON(CreateURLResponse{
		Status:  "success",
		Message: "URL updated successfully",
		Data: &CreateURLData{
			ID:          url.ID.String(),
			ShortCode:   url.ShortCode,
			ShortURL:    shortURL,
			OriginalURL: url.OriginalURL,
			CustomAlias: url.CustomAlias,
			ExpiresAt:   expiresAtStr,
		},
	})
}

func (h *Handler) DeleteURL(c *fiber.Ctx) error {
	urlIDStr := c.Params("id")
	urlID, err := uuid.Parse(urlIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid URL ID",
		})
	}

	err = h.service.DeleteURL(urlID)
	if err != nil {
		slog.Error("Failed to delete URL", "urlID", urlID, "error", err)
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "URL deleted successfully",
	})
}

func (h *Handler) GetAnalytics(c *fiber.Ctx) error {
	urlIDStr := c.Params("id")
	urlID, err := uuid.Parse(urlIDStr)
	if err != nil {
		return c.Status(400).JSON(AnalyticsResponse{
			Status:  "error",
			Message: "Invalid URL ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	clicks, err := h.service.GetAnalytics(urlID, limit, offset)
	if err != nil {
		slog.Error("Failed to get analytics", "urlID", urlID, "error", err)
		return c.Status(500).JSON(AnalyticsResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	var clicksData []*ClickData
	for _, click := range clicks {
		clicksData = append(clicksData, &ClickData{
			ID:        click.ID.String(),
			IPAddress: click.IPAddress,
			UserAgent: click.UserAgent,
			Referer:   click.Referer,
			CreatedAt: click.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(AnalyticsResponse{
		Status:  "success",
		Message: "Analytics retrieved successfully",
		Data:    clicksData,
	})
}

func getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header first
	ip := c.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	ip = c.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fall back to remote address
	ip = c.IP()
	if ip != "" {
		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			return ip
		}
		return host
	}

	return ""
}