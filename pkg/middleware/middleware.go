package middleware

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"product-service/pkg/logger"
)

// ValidationError represents a detailed validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse is a standard error response structure
type ErrorResponse struct {
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ValidationMiddleware creates a middleware for request validation
func ValidationMiddleware(validate *validator.Validate) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// // Get the logger
			// log := logger.GetLogger()

			// // Skip validation for methods that don't require a body
			// req := c.Request()
			// if req.Method == http.MethodGet || req.Method == http.MethodDelete || req.Method == http.MethodHead || req.Method == http.MethodOptions {
			// 	return next(c)
			// }

			// // Check if the request body is empty
			// if req.ContentLength == 0 {
			// 	log.Warn("Empty request body",
			// 		zap.String("path", req.URL.Path),
			// 		zap.String("method", req.Method),
			// 	)
			// 	return c.JSON(http.StatusBadRequest, ErrorResponse{
			// 		Message: "Request body cannot be empty",
			// 	})
			// }

			return next(c)
		}
	}
}

// CustomErrorHandler creates a custom error handler for Echo
func CustomErrorHandler(err error, c echo.Context) {
	// Get the logger
	log := logger.GetLogger()

	// Default error response
	var (
		code    = http.StatusInternalServerError
		message = "Internal Server Error"
		errors  []ValidationError
	)

	// Handle different types of errors
	switch e := err.(type) {
	case *echo.HTTPError:
		// Handle HTTP errors (like 404, 405, etc.)
		code = e.Code
		message = fmt.Sprintf("%v", e.Message)
		log.Warn("HTTP Error",
			zap.Int("code", code),
			zap.String("message", message),
		)

	case validator.ValidationErrors:
		// Handle validation errors
		code = http.StatusUnprocessableEntity
		message = "Validation failed"
		errors = make([]ValidationError, len(e))

		for i, fe := range e {
			errors[i] = ValidationError{
				Field:   fe.Field(),
				Message: getValidationErrorMessage(fe),
			}
		}

		log.Warn("Validation error",
			zap.String("message", message),
			zap.Any("validation_errors", errors),
		)

	default:
		// Log unexpected errors
		log.Error("Unhandled error",
			zap.Error(err),
			zap.String("path", c.Path()),
		)
	}

	// Prepare and send error response
	responseErr := ErrorResponse{
		Message: message,
		Errors:  errors,
	}

	// Ensure we don't override an already sent response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, responseErr)
		}
	}
}

// getValidationErrorMessage generates a human-readable error message
func getValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// RecoverMiddleware creates a middleware to recover from panics
func RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Get the logger
					log := logger.GetLogger()

					// Log the panic
					log.Error("Panic recovered",
						zap.Any("recover", r),
						zap.String("path", c.Path()),
					)

					// Send a 500 response
					err := c.JSON(http.StatusInternalServerError, ErrorResponse{
						Message: "Internal server error",
					})
					if err != nil {
						log.Error("Failed to send panic response", zap.Error(err))
					}
				}
			}()
			return next(c)
		}
	}
}
