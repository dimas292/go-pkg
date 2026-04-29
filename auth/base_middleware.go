package auth

import (
	"net/http"
	"strings"

	"github.com/dimas292/go-pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	// ContextUserID is the key for user ID in gin context.
	ContextUserID = "user_id"
	// ContextEmail is the key for email in gin context.
	ContextEmail = "email"
	// ContextRole is the key for role in gin context.
	ContextRole = "role"
)

// AuthMiddleware returns a Gin middleware that validates JWT tokens.
// It extracts the token from the Authorization header (Bearer <token>)
// and sets user_id, email, and role in the gin context.
func AuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "invalid authorization format, use: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		// Set user info in context for downstream handlers
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextRole, claims.Role)

		c.Next()
	}
}

// RoleMiddleware returns a Gin middleware that checks if the user
// has one of the allowed roles. Must be used after AuthMiddleware.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextRole)
		if !exists {
			response.Error(c, http.StatusForbidden, "role not found in context")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Error(c, http.StatusForbidden, "invalid role type")
			c.Abort()
			return
		}

		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "insufficient permissions")
		c.Abort()
	}
}

// GetUserID extracts the user ID from gin context. Returns "" if not found.
func GetUserID(c *gin.Context) string {
	id, exists := c.Get(ContextUserID)
	if !exists {
		return ""
	}
	userID, ok := id.(string)
	if !ok {
		return ""
	}
	return userID
}

// GetEmail extracts the email from gin context. Returns "" if not found.
func GetEmail(c *gin.Context) string {
	email, exists := c.Get(ContextEmail)
	if !exists {
		return ""
	}
	e, ok := email.(string)
	if !ok {
		return ""
	}
	return e
}

// GetRole extracts the role from gin context. Returns "" if not found.
func GetRole(c *gin.Context) string {
	role, exists := c.Get(ContextRole)
	if !exists {
		return ""
	}
	r, ok := role.(string)
	if !ok {
		return ""
	}
	return r
}
