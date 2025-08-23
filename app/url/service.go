package url

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/11SF/tinyurl/app/datamodel"
	"github.com/google/uuid"
)

type Service interface {
	ShortenURL(originalURL string, customAlias *string, userID *uuid.UUID, expiresAt *time.Time) (*datamodel.URL, error)
	GetOriginalURL(shortCodeOrAlias string) (*datamodel.URL, error)
	TrackClick(urlID uuid.UUID, ipAddress, userAgent, referer string) error
	GetUserURLs(userID uuid.UUID, limit, offset int) ([]*datamodel.URL, error)
	UpdateURL(urlID uuid.UUID, originalURL string, customAlias *string, expiresAt *time.Time) (*datamodel.URL, error)
	DeleteURL(urlID uuid.UUID) error
	GetAnalytics(urlID uuid.UUID, limit, offset int) ([]*datamodel.URLClick, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ShortenURL(originalURL string, customAlias *string, userID *uuid.UUID, expiresAt *time.Time) (*datamodel.URL, error) {
	if !isValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	var shortCode string
	var err error

	if customAlias != nil && *customAlias != "" {
		if !isValidAlias(*customAlias) {
			return nil, fmt.Errorf("invalid custom alias format")
		}
		
		exists, err := s.repo.IsCustomAliasExists(*customAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom alias: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("custom alias already exists")
		}
		shortCode = *customAlias
	} else {
		shortCode, err = s.generateUniqueShortCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	urlModel := &datamodel.URL{
		ID:          uuid.New(),
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CustomAlias: customAlias,
		UserID:      userID,
		ClickCount:  0,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	err = s.repo.CreateURL(urlModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	return urlModel, nil
}

func (s *service) GetOriginalURL(shortCodeOrAlias string) (*datamodel.URL, error) {
	url, err := s.repo.GetURLByShortCode(shortCodeOrAlias)
	if err != nil {
		url, err = s.repo.GetURLByCustomAlias(shortCodeOrAlias)
		if err != nil {
			return nil, fmt.Errorf("URL not found")
		}
	}

	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return nil, fmt.Errorf("URL has expired")
	}

	return url, nil
}

func (s *service) TrackClick(urlID uuid.UUID, ipAddress, userAgent, referer string) error {
	err := s.repo.IncrementClickCount(urlID)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	click := &datamodel.URLClick{
		ID:        uuid.New(),
		URLID:     urlID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   &referer,
		CreatedAt: time.Now(),
	}

	if referer == "" {
		click.Referer = nil
	}

	err = s.repo.CreateURLClick(click)
	if err != nil {
		return fmt.Errorf("failed to create click record: %w", err)
	}

	return nil
}

func (s *service) GetUserURLs(userID uuid.UUID, limit, offset int) ([]*datamodel.URL, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	urls, err := s.repo.GetURLsByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	return urls, nil
}

func (s *service) UpdateURL(urlID uuid.UUID, originalURL string, customAlias *string, expiresAt *time.Time) (*datamodel.URL, error) {
	if !isValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	if customAlias != nil && *customAlias != "" && !isValidAlias(*customAlias) {
		return nil, fmt.Errorf("invalid custom alias format")
	}

	urlModel := &datamodel.URL{
		ID:          urlID,
		OriginalURL: originalURL,
		CustomAlias: customAlias,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	err := s.repo.UpdateURL(urlModel)
	if err != nil {
		return nil, fmt.Errorf("failed to update URL: %w", err)
	}

	return urlModel, nil
}

func (s *service) DeleteURL(urlID uuid.UUID) error {
	err := s.repo.DeleteURL(urlID)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}
	return nil
}

func (s *service) GetAnalytics(urlID uuid.UUID, limit, offset int) ([]*datamodel.URLClick, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	clicks, err := s.repo.GetClickAnalytics(urlID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return clicks, nil
}

func (s *service) generateUniqueShortCode() (string, error) {
	const maxAttempts = 10
	
	for i := 0; i < maxAttempts; i++ {
		shortCode := generateShortCode()
		
		exists, err := s.repo.IsShortCodeExists(shortCode)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return shortCode, nil
		}
	}
	
	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxAttempts)
}

func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 7

	bytes := make([]byte, length*3/4+1)
	rand.Read(bytes)
	
	encoded := base64.URLEncoding.EncodeToString(bytes)
	encoded = strings.ReplaceAll(encoded, "-", "")
	encoded = strings.ReplaceAll(encoded, "_", "")
	encoded = strings.ReplaceAll(encoded, "=", "")
	
	if len(encoded) >= length {
		return encoded[:length]
	}
	
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = charset[int(bytes[i%len(bytes)])%len(charset)]
	}
	return string(b)
}

func isValidURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func isValidAlias(alias string) bool {
	if len(alias) < 3 || len(alias) > 50 {
		return false
	}
	
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, alias)
	return match
}