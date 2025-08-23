package url

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/11SF/tinyurl/app/datamodel"
	"github.com/google/uuid"
)

type Repository interface {
	CreateURL(url *datamodel.URL) error
	GetURLByShortCode(shortCode string) (*datamodel.URL, error)
	GetURLByCustomAlias(alias string) (*datamodel.URL, error)
	GetURLsByUserID(userID uuid.UUID, limit, offset int) ([]*datamodel.URL, error)
	UpdateURL(url *datamodel.URL) error
	DeleteURL(id uuid.UUID) error
	IncrementClickCount(urlID uuid.UUID) error
	CreateURLClick(click *datamodel.URLClick) error
	GetClickAnalytics(urlID uuid.UUID, limit, offset int) ([]*datamodel.URLClick, error)
	IsShortCodeExists(shortCode string) (bool, error)
	IsCustomAliasExists(alias string) (bool, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateURL(url *datamodel.URL) error {
	query := `
		INSERT INTO urls (id, short_code, original_url, custom_alias, user_id, expires_at, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.Exec(query, url.ID, url.ShortCode, url.OriginalURL, url.CustomAlias, url.UserID, 
		url.ExpiresAt, url.CreatedAt, url.UpdatedAt, url.IsActive)
	return err
}

func (r *repository) GetURLByShortCode(shortCode string) (*datamodel.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, user_id, click_count, expires_at, created_at, updated_at, is_active
		FROM urls WHERE short_code = $1 AND is_active = true`
	
	url := &datamodel.URL{}
	err := r.db.QueryRow(query, shortCode).Scan(
		&url.ID, &url.ShortCode, &url.OriginalURL, &url.CustomAlias, &url.UserID,
		&url.ClickCount, &url.ExpiresAt, &url.CreatedAt, &url.UpdatedAt, &url.IsActive,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("URL not found")
	}
	return url, err
}

func (r *repository) GetURLByCustomAlias(alias string) (*datamodel.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, user_id, click_count, expires_at, created_at, updated_at, is_active
		FROM urls WHERE custom_alias = $1 AND is_active = true`
	
	url := &datamodel.URL{}
	err := r.db.QueryRow(query, alias).Scan(
		&url.ID, &url.ShortCode, &url.OriginalURL, &url.CustomAlias, &url.UserID,
		&url.ClickCount, &url.ExpiresAt, &url.CreatedAt, &url.UpdatedAt, &url.IsActive,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("URL not found")
	}
	return url, err
}

func (r *repository) GetURLsByUserID(userID uuid.UUID, limit, offset int) ([]*datamodel.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, user_id, click_count, expires_at, created_at, updated_at, is_active
		FROM urls WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var urls []*datamodel.URL
	for rows.Next() {
		url := &datamodel.URL{}
		err := rows.Scan(
			&url.ID, &url.ShortCode, &url.OriginalURL, &url.CustomAlias, &url.UserID,
			&url.ClickCount, &url.ExpiresAt, &url.CreatedAt, &url.UpdatedAt, &url.IsActive,
		)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func (r *repository) UpdateURL(url *datamodel.URL) error {
	query := `
		UPDATE urls SET original_url = $2, custom_alias = $3, expires_at = $4, updated_at = $5, is_active = $6
		WHERE id = $1`
	
	_, err := r.db.Exec(query, url.ID, url.OriginalURL, url.CustomAlias, url.ExpiresAt, time.Now(), url.IsActive)
	return err
}

func (r *repository) DeleteURL(id uuid.UUID) error {
	query := `UPDATE urls SET is_active = false, updated_at = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, time.Now())
	return err
}

func (r *repository) IncrementClickCount(urlID uuid.UUID) error {
	query := `UPDATE urls SET click_count = click_count + 1, updated_at = $2 WHERE id = $1`
	_, err := r.db.Exec(query, urlID, time.Now())
	return err
}

func (r *repository) CreateURLClick(click *datamodel.URLClick) error {
	query := `
		INSERT INTO url_clicks (id, url_id, ip_address, user_agent, referer, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := r.db.Exec(query, click.ID, click.URLID, click.IPAddress, click.UserAgent, click.Referer, click.CreatedAt)
	return err
}

func (r *repository) GetClickAnalytics(urlID uuid.UUID, limit, offset int) ([]*datamodel.URLClick, error) {
	query := `
		SELECT id, url_id, ip_address, user_agent, referer, created_at
		FROM url_clicks WHERE url_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(query, urlID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var clicks []*datamodel.URLClick
	for rows.Next() {
		click := &datamodel.URLClick{}
		err := rows.Scan(&click.ID, &click.URLID, &click.IPAddress, &click.UserAgent, &click.Referer, &click.CreatedAt)
		if err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}
	return clicks, nil
}

func (r *repository) IsShortCodeExists(shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`
	var exists bool
	err := r.db.QueryRow(query, shortCode).Scan(&exists)
	return exists, err
}

func (r *repository) IsCustomAliasExists(alias string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE custom_alias = $1)`
	var exists bool
	err := r.db.QueryRow(query, alias).Scan(&exists)
	return exists, err
}