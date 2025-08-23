package datamodel

import (
	"time"
	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	CustomAlias *string    `json:"custom_alias,omitempty" db:"custom_alias"`
	UserID      *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	ClickCount  int64      `json:"click_count" db:"click_count"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
}

type URLClick struct {
	ID        uuid.UUID `json:"id" db:"id"`
	URLID     uuid.UUID `json:"url_id" db:"url_id"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Referer   *string   `json:"referer,omitempty" db:"referer"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}