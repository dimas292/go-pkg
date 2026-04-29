# Shared Go module reusable building blocks for all Go projects.

Provides generic CRUD, JWT authentication, structured logging, error handling, input validation, and more out-of-the-box.

## Install

```bash
go get github.com/dimas292/go-pkg@v1.0.0
```

## Package Overview

| Package      | Description                                                    |
|--------------|----------------------------------------------------------------|
| `apperror`   | Structured errors — separates internal errors from client-safe messages |
| `auth`       | JWT service + Auth/Role middleware for Gin                      |
| `config`     | YAML config loader                                             |
| `database`   | PostgreSQL (GORM) + Redis connection setup                     |
| `handler`    | Generic CRUD HTTP handler (Go generics)                        |
| `logger`     | Structured logging via zerolog                                 |
| `model`      | Base model: UUID primary key, timestamps, soft-delete          |
| `repository` | Generic GORM repository (CRUD)                                 |
| `response`   | Standardized API response envelope                             |
| `router`     | Module interface for plug-in architecture                      |
| `server`     | Server bootstrap + graceful shutdown                           |
| `service`    | Generic service layer                                          |
| `validator`  | Input validation + HTML sanitization                           |

## Quick Start

### 1. Define a Model

Any model must embed `model.BaseModel` to get a UUID primary key, timestamps, and soft-delete support.

```go
package url

import "github.com/dimas292/go-pkg/model"

type Url struct {
    model.BaseModel
    ShortUrl    string `json:"short_url" gorm:"type:varchar(10);not null"`
    OriginalUrl string `json:"original_url" gorm:"type:varchar(255);not null"`
}
```

### 2. Full CRUD (Zero Boilerplate)

For entities that only need standard CRUD — no custom logic required:

```go
package category

import (
    "github.com/dimas292/go-pkg/handler"
    "github.com/dimas292/go-pkg/model"
    "github.com/dimas292/go-pkg/repository"
    "github.com/dimas292/go-pkg/router"
    "github.com/dimas292/go-pkg/service"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type Category struct {
    model.BaseModel
    Name string `json:"name" binding:"required"`
}

type CategoryModule struct{ db *gorm.DB }

// Ensure CategoryModule implements router.Module
var _ router.Module = (*CategoryModule)(nil)

func NewCategoryModule(db *gorm.DB) *CategoryModule {
    db.AutoMigrate(&Category{})
    return &CategoryModule{db: db}
}

func (m *CategoryModule) RegisterRoutes(rg *gin.RouterGroup) {
    repo := repository.NewBaseRepository[Category, *Category](m.db)
    svc := service.NewBaseService[Category, *Category](repo)
    hdl := handler.NewBaseHandler[Category, *Category](svc)
    hdl.RegisterCRUD(rg.Group("/categories"))
    // 5 endpoints are registered automatically:
    // POST   /categories      → Create
    // GET    /categories      → FindAll (paginated)
    // GET    /categories/:id  → FindByID
    // PUT    /categories/:id  → Update
    // DELETE /categories/:id  → Delete (soft-delete)
}
```

### 3. Custom Logic Module

For features with custom business logic (auth, URL shortener, etc.):

```go
package auth

import (
    "github.com/dimas292/go-pkg/apperror"
    "github.com/dimas292/go-pkg/logger"
    "github.com/dimas292/go-pkg/response"
    "github.com/gin-gonic/gin"
)

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
    // AppError: safe message sent to client, internal error is logged
    user, err := s.findByEmail(req.Email)
    if err != nil {
        return nil, apperror.Internal("failed to authenticate", err)
    }

    // Domain error: displayed directly to client
    if !checkPassword(user, req.Password) {
        return nil, apperror.Unauthorized("invalid email or password")
    }

    logger.Info().Str("user_id", user.GetID()).Msg("user logged in")
    return &AuthResponse{User: user.ToResponse(), Token: token}, nil
}

// In the handler, a single call handles all error types:
func (m *AuthModule) handleLogin(c *gin.Context) {
    result, err := m.service.Login(req)
    if err != nil {
        response.HandleError(c, err) // auto-routes based on error type
        return
    }
    response.Success(c, "login successful", result)
}
```

### 4. Server Bootstrap

`server.New()` automatically initializes: logger, validators, config, PostgreSQL, Redis, and JWT.

A health check endpoint is also registered at `GET /api/v1/health`.

```go
package main

import (
    "github.com/dimas292/go-pkg/server"
    "your-project/modules/auth"
    "your-project/modules/url"
)

func main() {
    srv := server.New("config.yml")

    srv.RegisterModules(
        auth.NewAuthModule(srv.DB, srv.Redis, srv.JWT),
        url.NewUrlModule(srv.DB, srv.Redis, srv.JWT),
    )

    srv.Run() // graceful shutdown included
}
```

## Error Handling

Use `apperror` to create structured errors with appropriate HTTP status codes. For `Internal` errors, the wrapped `err` is logged but never sent to the client.

```go
apperror.BadRequest("invalid input")           // 400
apperror.Unauthorized("invalid credentials")   // 401
apperror.Forbidden("insufficient permissions") // 403
apperror.NotFound("user not found")            // 404
apperror.Conflict("email already registered")  // 409
apperror.Internal("something went wrong", err) // 500 (err is logged, not sent to client)
apperror.Wrap(code, "custom message", err)     // any status code
```

In handlers, use `response.HandleError()` to automatically route errors:

```go
response.HandleError(c, err)
// - AppError       → uses its Code and Message
// - ValidationError → 400 with field-level error details
// - Other errors   → 500 with generic message (real error is logged)
```

## Auth Middleware

Protect routes with JWT authentication and role-based access control:

```go
import "github.com/dimas292/go-pkg/auth"

// Require a valid JWT token
rg.Use(auth.AuthMiddleware(jwtService))

// Require specific roles (must be used after AuthMiddleware)
rg.Use(auth.RoleMiddleware("admin", "editor"))

// Extract user info from context in handlers
userID := auth.GetUserID(c)   // string
email  := auth.GetEmail(c)    // string
role   := auth.GetRole(c)     // string
```

## Input Validation & Sanitization

Custom validation tags are registered automatically by `server.New()`. Use them in your struct bindings:

```go
type CreateRequest struct {
    Name  string `json:"name" binding:"required,notblank,safeinput"`
    Email string `json:"email" binding:"required,email"`
    URL   string `json:"url" binding:"required,url"`
}
```

| Tag         | Description                                           |
|-------------|-------------------------------------------------------|
| `notblank`  | Rejects empty strings or whitespace-only values       |
| `safeinput` | Blocks XSS patterns, SQL injection keywords           |

All string fields are automatically sanitized (HTML tags stripped) on `Create` and `Update` via the generic handler.

## API Response Format

All endpoints return a consistent JSON envelope:

**Success response:**

```json
{
  "status": 200,
  "message": "retrieved successfully",
  "data": { "..." }
}
```

**Paginated response:**

```json
{
  "status": 200,
  "message": "retrieved successfully",
  "data": [ "..." ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 50,
    "total_page": 5
  }
}
```

**Error response:**

```json
{
  "status": 400,
  "message": "validation failed",
  "errors": {
    "name": "field is required",
    "email": "must be a valid email address"
  }
}
```

**Pagination query parameters:**

| Parameter  | Default | Constraints    |
|------------|---------|----------------|
| `page`     | `1`     | min: 1         |
| `per_page` | `10`    | min: 1, max: 100 |

## Config Format

```yaml
app:
  name: your-app
  port: ":8080"
  jwt:
    secret: "your-secret-key"
    expiration: 24  # hours
  db:
    postgres:
      dbhost: localhost
      dbuser: postgres
      dbpassword: your-password
      dbname: your-db
    redis:
      host: localhost
      port: "6379"
```

## Project Structure

```
go-pkg/
├── apperror/       # Structured error types (AppError)
├── auth/           # JWT service, Auth & Role middleware
├── config/         # YAML config loader & structs
├── database/       # PostgreSQL & Redis initialization
├── handler/        # Generic CRUD HTTP handlers
├── logger/         # Zerolog-based structured logging
├── model/          # BaseModel with UUID, timestamps, soft-delete
├── repository/     # Generic GORM repository (CRUD)
├── response/       # Standardized API response envelope
├── router/         # Module interface for plug-in architecture
├── server/         # Server bootstrap & graceful shutdown
├── service/        # Generic service layer
└── validator/      # Input validation & HTML sanitization
```

