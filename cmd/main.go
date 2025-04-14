package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"product-service/internal/handler"
	"product-service/internal/repository"
	"product-service/pkg/database"
	"product-service/pkg/logger"
	customMiddleware "product-service/pkg/middleware"
)

// CustomValidator implements validator.Validate
type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements the Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// Initialize logger
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	zapLogger := logger.InitLogger(env)
	defer zapLogger.Sync()

	// Create Echo instance
	e := echo.New()

	// Custom Error Handler
	e.HTTPErrorHandler = customMiddleware.CustomErrorHandler

	// Middleware
	e.Use(customMiddleware.RecoverMiddleware())

	// Add zap logger middleware
	e.Use(logger.LoggerMiddleware(zapLogger))

	// CORS and Security Middleware
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.SecureWithConfig(echoMiddleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Request Timeout
	e.Use(echoMiddleware.TimeoutWithConfig(echoMiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Create database connection
	db := database.NewConnection()
	defer db.Close()

	// Initialize database schema
	if err := database.InitSchema(db); err != nil {
		zapLogger.Fatal("Failed to initialize database schema",
			zap.Error(err),
			zap.String("action", "database_schema_init"),
		)
	}

	// Create repository and handler
	productRepo := repository.NewProductRepository(db)
	productHandler := handler.NewProductHandler(productRepo)

	// Create validator
	validate := validator.New()

	// Set custom validator
	e.Validator = &CustomValidator{validator: validate}

	// Add validation middleware
	// e.Use(customMiddleware.ValidationMiddleware(validate))

	// Routes
	v1 := e.Group("/api/v1")

	// Product routes
	v1.POST("/products", productHandler.CreateProduct)
	v1.GET("/products", productHandler.ListProducts)
	v1.GET("/products/all", productHandler.GetAllProducts)
	v1.GET("/products/:id", productHandler.GetProduct)
	v1.PUT("/products/:id", productHandler.UpdateProduct)
	v1.DELETE("/products/:id", productHandler.DeleteProduct)

	// Bulk operations routes
	v1.POST("/products/bulk/generate", productHandler.BulkGenerateProducts)
	v1.DELETE("/products/bulk", productHandler.DeleteAllProducts)

	// New route to get total product count
	v1.GET("/products/count", productHandler.GetProductCount)

	// Prometheus metrics route (placeholder for future implementation)
	e.GET("/metrics", func(c echo.Context) error {
		return c.String(http.StatusOK, "Metrics endpoint")
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		dbStatus := "healthy"
		if err := db.Ping(); err != nil {
			dbStatus = "unhealthy"
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"version":   "1.0.0",
			"env":       env,
			"database":  dbStatus,
			"timestamp": time.Now().UTC(),
		})
	})

	// Readiness probe
	e.GET("/ready", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ready",
		})
	})

	// Liveness probe
	e.GET("/live", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "alive",
		})
	})

	// Configure server
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	// Print routes for development
	if env == "development" {
		for _, route := range e.Routes() {
			zapLogger.Info("Registered Route",
				zap.String("method", route.Method),
				zap.String("path", route.Path),
			)
		}
	}

	// Start server
	zapLogger.Info("Starting server",
		zap.String("port", port),
		zap.String("environment", env),
	)

	// Graceful shutdown
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- e.Start(":" + port)
	}()

	// Block main and wait for shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for either a server error or an interrupt signal
	select {
	case err := <-serverErrors:
		zapLogger.Error("Server error",
			zap.Error(err),
			zap.String("action", "server_shutdown"),
		)
	case <-shutdown:
		zapLogger.Info("Starting graceful shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			zapLogger.Error("Graceful shutdown failed",
				zap.Error(err),
			)
		}
	}
}
