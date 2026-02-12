<p align="center">
  <img src="https://img.shields.io/badge/A4AD-Post%20Service-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="A4AD Post Service">
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/PostgreSQL-16+-4169E1?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/Status-Early%20Development-orange?style=flat-square" alt="Status">
</p>

---

## 📝 Overview

**Post Service** is the core content management service for the A4AD Forum. It handles post creation, editing, deletion, and retrieval with support for pagination, sorting (new/hot/top), and real-time view/like counters.

---

## ✨ Features

- 📝 **Post Management** — Create, read, update, and delete forum posts
- 👍 **Like System** — Like/unlike posts with atomic counter updates
- 👁️ **View Tracking** — Real-time view counter with deduplication
- 🔥 **Sorting Options** — Sort by new, hot (trending), or top (most liked)
- 🔍 **Full-text Search** — Search posts by title and content
- 🏷️ **Tagging** — Categorize posts with tags
- 📄 **Pagination** — Cursor-based pagination for optimal performance
- 📊 **Analytics** — Post engagement metrics and statistics

---

## 🛠 Tech Stack

- **Language:** Go 1.23+
- **Framework:** Standard library + Chi router (or Fiber)
- **Database:** PostgreSQL 16+
- **ORM/Query Builder:** sqlc or GORM
- **Migrations:** golang-migrate
- **Full-text Search:** PostgreSQL tsvector/trigram
- **Caching:** Redis for hot posts and counters
- **Events:** Kafka or Redis Pub/Sub for real-time updates
- **Configuration:** Viper

---

## 📊 Architecture

```mermaid
flowchart LR
    Gateway["API Gateway<br/>Go + Fiber"] --> Post["Post Service<br/>Go"]
    Post --> PostgreSQL[("PostgreSQL<br/>Posts Table")]
    Post --> Redis[("Redis<br/>Hot Posts Cache")]
    Post --> Kafka[("Kafka<br/>Post Events")]
    
    Kafka -.->|"post.created"| Notification["Notification Service<br/>NestJS"]
    Kafka -.->|"post.liked"| Notification
    Profile["Profile Service<br/>Go"] -.->|"Stats sync"| Post
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 16+
- Redis 7.x (optional, for caching)
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/A4AD-team/post-service.git
   cd post-service
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Start infrastructure:**
   ```bash
   docker compose up -d postgres redis
   ```

4. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

5. **Run database migrations:**
   ```bash
   make migrate-up
   # Or: migrate -path ./migrations -database "postgres://user:pass@localhost:5432/post_db?sslmode=disable" up
   ```

6. **Run the service:**
   ```bash
   # Development mode
   go run ./cmd/server

   # Or with custom config
   go run ./cmd/server -config=./config/local.yaml
   ```

The service will be available at `http://localhost:8083`

---

## 🔧 Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `APP_ENV` | Environment (development, staging, production) | `development` | No |
| `APP_PORT` | HTTP server port | `8083` | No |
| `APP_HOST` | HTTP server host | `0.0.0.0` | No |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/post_db?sslmode=disable` | Yes |
| `DATABASE_MAX_CONNS` | Max database connections | `25` | No |
| `REDIS_ADDR` | Redis connection string | `localhost:6379` | No |
| `REDIS_PASSWORD` | Redis password | `` | No |
| `REDIS_DB` | Redis database number | `0` | No |
| `CACHE_TTL_SECONDS` | Cache TTL in seconds | `300` | No |
| `HOT_POSTS_CACHE_MINUTES` | Hot posts cache duration | `10` | No |
| `MAX_POST_TITLE_LENGTH` | Max post title length | `300` | No |
| `MAX_POST_CONTENT_LENGTH` | Max post content length | `50000` | No |
| `PAGINATION_DEFAULT_LIMIT` | Default posts per page | `20` | No |
| `PAGINATION_MAX_LIMIT` | Max posts per page | `100` | No |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` | No |
| `LOG_LEVEL` | Logging level | `info` | No |
| `METRICS_ENABLED` | Enable Prometheus metrics | `true` | No |

---

## 📡 API Endpoints

### Public Routes

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/posts` | List posts with pagination and sorting |
| `GET` | `/api/v1/posts/:id` | Get single post by ID |
| `GET` | `/api/v1/posts/search` | Search posts (query: `?q=keyword`) |
| `GET` | `/api/v1/posts/hot` | Get trending/hot posts |
| `GET` | `/api/v1/posts/top` | Get top-rated posts |
| `GET` | `/api/v1/posts/:id/comments` | Get comments for a post |

### Protected Routes (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/posts` | Create a new post |
| `PATCH` | `/api/v1/posts/:id` | Update own post |
| `DELETE` | `/api/v1/posts/:id` | Delete own post |
| `POST` | `/api/v1/posts/:id/like` | Like a post |
| `DELETE` | `/api/v1/posts/:id/like` | Unlike a post |

### Query Parameters for List Endpoints

| Parameter | Description | Default |
|-----------|-------------|---------|
| `limit` | Number of posts per page | `20` |
| `cursor` | Pagination cursor (from previous response) | `` |
| `sort` | Sort order: `new`, `hot`, `top` | `new` |
| `author` | Filter by author username | `` |
| `tag` | Filter by tag | `` |

### Example Post Response

```json
{
  "id": "post-uuid",
  "title": "Getting Started with Go 1.23",
  "content": "Go 1.23 introduces several exciting features...",
  "authorId": "user-uuid",
  "author": {
    "username": "johndoe",
    "avatarUrl": "https://cdn.example.com/avatars/johndoe.png"
  },
  "tags": ["golang", "programming", "tutorial"],
  "views": 1523,
  "likesCount": 42,
  "commentsCount": 15,
  "isLikedByMe": false,
  "createdAt": "2026-02-10T14:30:00Z",
  "updatedAt": "2026-02-10T14:30:00Z"
}
```

---

## 🩺 Health Checks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | `GET` | Overall health status |
| `/health/live` | `GET` | Liveness probe |
| `/health/ready` | `GET` | Readiness probe (checks database) |
| `/metrics` | `GET` | Prometheus metrics |

### Example Response

```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T15:30:00Z",
  "version": "0.1.0",
  "checks": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestPostService ./...

# Run integration tests
go test -tags=integration ./...
```

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <strong>Built with ❤️ by A4AD Team</strong>
</p>
