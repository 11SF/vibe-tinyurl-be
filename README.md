# TinyURL Backend Service

A high-performance URL shortening service built with Go, Fiber, and PostgreSQL.

## Features

- **URL Shortening**: Convert long URLs to short, memorable links
- **Custom Aliases**: Create personalized short URLs
- **Click Analytics**: Track and analyze link performance
- **Expiration Dates**: Set automatic link expiration
- **Link Management**: Full CRUD operations for URL management
- **User Authentication**: Integration with Auth.NS service
- **High Performance**: Built with Fiber web framework
- **Database**: PostgreSQL with connection pooling
- **Monitoring**: Prometheus metrics and health checks

## API Endpoints

### URL Management

#### Create Short URL
```
POST /api/tinyurl/v1/shorten
```

**Request:**
```json
{
  "url": "https://example.com/very/long/url",
  "custom_alias": "my-link", // optional
  "expires_at": "2024-12-31T23:59:59Z" // optional
}
```

**Response:**
```json
{
  "status": "success",
  "message": "URL created successfully",
  "data": {
    "id": "uuid",
    "short_code": "abc123",
    "short_url": "https://your-domain.com/abc123",
    "original_url": "https://example.com/very/long/url",
    "custom_alias": "my-link",
    "expires_at": "2024-12-31T23:59:59Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Redirect to Original URL
```
GET /{shortCode}
```

Returns a 301 redirect to the original URL and tracks the click.

#### Get User URLs
```
GET /api/tinyurl/v1/urls?limit=20&offset=0
```

**Headers:**
```
X-User-ID: user-uuid
```

**Response:**
```json
{
  "status": "success",
  "message": "URLs retrieved successfully",
  "data": [
    {
      "id": "uuid",
      "short_code": "abc123",
      "short_url": "https://your-domain.com/abc123",
      "original_url": "https://example.com/very/long/url",
      "custom_alias": "my-link",
      "click_count": 42,
      "expires_at": "2024-12-31T23:59:59Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### Update URL
```
PUT /api/tinyurl/v1/urls
```

**Request:**
```json
{
  "id": "url-uuid",
  "url": "https://updated-example.com",
  "custom_alias": "updated-link",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

#### Delete URL
```
DELETE /api/tinyurl/v1/urls/{id}
```

### Analytics

#### Get Click Analytics
```
GET /api/tinyurl/v1/analytics/{id}?limit=20&offset=0
```

**Response:**
```json
{
  "status": "success",
  "message": "Analytics retrieved successfully",
  "data": [
    {
      "id": "click-uuid",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "referer": "https://google.com",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### System

#### Health Check
```
GET /health
```

#### Metrics
```
GET /metrics
```

#### Metrics Dashboard
```
GET /metrics/dashboard
```

## Setup

### Prerequisites

- Go 1.23.3 or later
- PostgreSQL 13 or later
- Redis 6 or later (optional, for caching)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd tinyurl
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp configs/.env.example configs/.env
# Edit configs/.env with your configuration
```

4. Run database migrations:
```bash
psql -d your_database -f db-migration/v1.0.sql
```

5. Run the application:
```bash
go run cmd/main.go
```

### Docker

Build and run with Docker:

```bash
docker build -t tinyurl .
docker run -p 8080:8080 --env-file configs/.env tinyurl
```

## Configuration

Environment variables:

- `APP_NAME`: Application name
- `APP_PORT`: Server port (default: 8080)
- `BASE_URL`: Base URL for short links (e.g., https://tiny.ly)
- `DB_DSN`: PostgreSQL connection string
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port
- `REDIS_PASSWORD`: Redis password (optional)
- `REDIS_DB`: Redis database number
- `AUTH_SERVICE_URL`: Auth.NS service URL

## Database Schema

The service uses PostgreSQL with the following main tables:

- `urls`: Stores URL mappings, metadata, and statistics
- `url_clicks`: Tracks individual click events for analytics

See `db-migration/v1.0.sql` for the complete schema.

## Performance

- Supports high concurrent load with Fiber framework
- Database connection pooling
- Efficient short code generation with collision detection
- Indexed database queries for fast lookups
- Prometheus metrics for monitoring

## Architecture

Following microservice patterns:
- Clean architecture with separated concerns
- Repository pattern for data access
- Service layer for business logic
- HTTP handlers for API endpoints
- Dependency injection for testability