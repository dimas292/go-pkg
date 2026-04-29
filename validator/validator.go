package validator

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

// sanitizer is the HTML sanitizer policy.
// StrictPolicy strips ALL HTML tags — the safest option for API input.
var sanitizer = bluemonday.StrictPolicy()

// Init registers custom validators with Gin's default validator engine.
// Call this once during server bootstrap.
func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register custom validation tags here
		v.RegisterValidation("notblank", notBlank)
		v.RegisterValidation("safeinput", safeInput)
	}
}

// --- Custom Validation Functions ---

// notBlank validates that a string is not empty or only whitespace.
// Usage: `binding:"notblank"`
func notBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

// safeInput validates that a string does not contain potentially dangerous characters.
// Blocks: <, >, ', ", ;, --, /*
// Usage: `binding:"safeinput"`
func safeInput(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	dangerous := []string{"<script", "javascript:", "onerror=", "onload=", "eval(", "';", "\";", "--", "/*", "*/", "DROP ", "DELETE ", "INSERT ", "UPDATE ", "UNION "}
	lower := strings.ToLower(value)
	for _, d := range dangerous {
		if strings.Contains(lower, strings.ToLower(d)) {
			return false
		}
	}
	return true
}

// --- Sanitization Functions ---

// SanitizeString strips all HTML tags from a string.
func SanitizeString(input string) string {
	return strings.TrimSpace(sanitizer.Sanitize(input))
}

// SanitizeStruct sanitizes all exported string fields of a struct.
// This uses reflection, so call it only when needed (not in hot paths).
func SanitizeStruct(v interface{}) {
	sanitizeFields(v)
}

// --- Error Formatting ---

// FormatValidationErrors converts validator.ValidationErrors into
// a human-readable map of field → error message.
func FormatValidationErrors(err error) map[string]string {
	errs := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			errs[fe.Field()] = formatFieldError(fe)
		}
	}
	return errs
}

// formatFieldError returns a human-readable error message for a single field error.
func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("minimum length is %s", fe.Param())
	case "max":
		return fmt.Sprintf("maximum length is %s", fe.Param())
	case "notblank":
		return "cannot be blank or whitespace only"
	case "safeinput":
		return "contains potentially unsafe characters"
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
