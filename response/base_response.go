package response

import (
	"errors"
	"net/http"

	"github.com/dimas292/go-pkg/apperror"
	"github.com/dimas292/go-pkg/logger"
	"github.com/dimas292/go-pkg/validator"
	"github.com/gin-gonic/gin"
	govalidator "github.com/go-playground/validator/v10"
)

// Response is the standardized API response envelope.
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Meta holds pagination metadata.
type Meta struct {
	Page      int   `json:"page"`
	PerPage   int   `json:"per_page"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

// PaginationQuery represents incoming pagination parameters.
type PaginationQuery struct {
	Page    int `form:"page,default=1" binding:"min=1"`
	PerPage int `form:"per_page,default=10" binding:"min=1,max=100"`
}

// Offset calculates the database offset from page and per_page.
func (p PaginationQuery) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Success sends a 200 JSON response.
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 JSON response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	})
}

// Paginated sends a 200 JSON response with pagination metadata.
func Paginated(c *gin.Context, message string, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Response{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

// Error sends an error JSON response with the given status code.
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Status:  statusCode,
		Message: message,
	})
}

// HandleError inspects the error type and sends the appropriate response.
// - AppError: uses its Code and Message, logs internal Err if present.
// - ValidationErrors: formats field-level errors into a readable map.
// - Other errors: returns 500 with a generic message, logs the real error.
func HandleError(c *gin.Context, err error) {
	// Check for AppError
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		// Log internal error if present
		if appErr.Err != nil {
			logger.Error().
				Err(appErr.Err).
				Int("status", appErr.Code).
				Str("path", c.FullPath()).
				Str("method", c.Request.Method).
				Msg(appErr.Message)
		}
		c.JSON(appErr.Code, Response{
			Status:  appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	// Check for validation errors
	var ve govalidator.ValidationErrors
	if errors.As(err, &ve) {
		fieldErrors := validator.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, Response{
			Status:  http.StatusBadRequest,
			Message: "validation failed",
			Errors:  fieldErrors,
		})
		return
	}

	// Unknown error — log it, return generic message
	logger.Error().
		Err(err).
		Str("path", c.FullPath()).
		Str("method", c.Request.Method).
		Msg("unexpected error")

	c.JSON(http.StatusInternalServerError, Response{
		Status:  http.StatusInternalServerError,
		Message: "internal server error",
	})
}
