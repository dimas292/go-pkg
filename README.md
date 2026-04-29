# go-pkg

Shared Go module — reusable building blocks untuk semua project Go.

Menyediakan generic CRUD, auth, structured logging, error handling, dan lainnya out-of-the-box.

## Install

```bash
go get github.com/dimas292/go-pkg@latest
```

## Package Overview

| Package | Deskripsi |
|---------|-----------|
| `apperror` | Structured error — pisahkan internal error dari client message |
| `auth` | JWT service + Auth/Role middleware untuk Gin |
| `config` | YAML config loader |
| `database` | PostgreSQL (GORM) + Redis connection setup |
| `handler` | Generic CRUD HTTP handler (Go generics) |
| `logger` | Structured logging via zerolog |
| `model` | Base model: UUID, timestamps, soft-delete |
| `repository` | Generic GORM repository (CRUD) |
| `response` | Standardized API response envelope |
| `router` | Module interface untuk plug-in architecture |
| `server` | Server bootstrap + graceful shutdown |
| `service` | Generic service layer |
| `validator` | Input validation + HTML sanitization |

## Quick Start

### 1. Define Model

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

Untuk entity yang hanya butuh CRUD standar:

```go
package category

import (
    "github.com/dimas292/go-pkg/handler"
    "github.com/dimas292/go-pkg/model"
    "github.com/dimas292/go-pkg/repository"
    "github.com/dimas292/go-pkg/service"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type Category struct {
    model.BaseModel
    Name string `json:"name" binding:"required"`
}

type CategoryModule struct{ db *gorm.DB }

func NewCategoryModule(db *gorm.DB) *CategoryModule {
    db.AutoMigrate(&Category{})
    return &CategoryModule{db: db}
}

func (m *CategoryModule) RegisterRoutes(rg *gin.RouterGroup) {
    repo := repository.NewBaseRepository[Category, *Category](m.db)
    svc := service.NewBaseService[Category, *Category](repo)
    hdl := handler.NewBaseHandler[Category, *Category](svc)
    hdl.RegisterCRUD(rg.Group("/categories"))
    // 5 endpoints langsung tersedia:
    // POST   /categories      → Create
    // GET    /categories      → FindAll (paginated)
    // GET    /categories/:id  → FindByID
    // PUT    /categories/:id  → Update
    // DELETE /categories/:id  → Delete
}
```

### 3. Custom Logic Module

Untuk fitur dengan business logic khusus (auth, url shortener, dll):

```go
package auth

import (
    "github.com/dimas292/go-pkg/apperror"
    "github.com/dimas292/go-pkg/logger"
    "github.com/dimas292/go-pkg/response"
)

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
    // AppError: safe message ke client, internal error di-log
    user, err := s.findByEmail(req.Email)
    if err != nil {
        return nil, apperror.Internal("failed to authenticate", err)
    }

    // Domain error: langsung tampil ke client
    if !checkPassword(user, req.Password) {
        return nil, apperror.Unauthorized("invalid email or password")
    }

    logger.Info().Str("user_id", user.ID).Msg("user logged in")
    return &AuthResponse{User: user.ToResponse(), Token: token}, nil
}

// Di handler, cukup 1 baris untuk handle semua error type:
func (m *AuthModule) handleLogin(c *gin.Context) {
    result, err := m.service.Login(req)
    if err != nil {
        response.HandleError(c, err) // auto-route berdasarkan error type
        return
    }
    response.Success(c, "login successful", result)
}
```

### 4. Server Bootstrap

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

```go
apperror.BadRequest("invalid input")          // 400
apperror.Unauthorized("invalid credentials")  // 401
apperror.Forbidden("insufficient permissions")// 403
apperror.NotFound("user not found")           // 404
apperror.Conflict("email already registered") // 409
apperror.Internal("something went wrong", err)// 500 (err di-log, tidak ke client)
```

## API Response Format

Semua endpoint mengembalikan format yang konsisten:

```json
{
  "status": 200,
  "message": "retrieved successfully",
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 50,
    "total_page": 5
  }
}
```

## Config Format

```yaml
app:
  name: your-app
  port: ":8080"
  jwt:
    secret: "your-secret-key"
    expiration: 24
  db:
    postgres:
      dbhost: localhost
      dbuser: postgres
      dbpassword: your-password
      dbname: your-db
    redis:
      host: localhost
      port: 6379
```

## License

MIT
