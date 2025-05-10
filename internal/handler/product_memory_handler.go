package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"product-service/internal/models"
	"product-service/internal/repository"
	"product-service/pkg/logger"
)

// ProductMemoryHandler handles HTTP requests for in-memory products
type ProductMemoryHandler struct {
	repo   *repository.ProductMemoryRepository
	logger *zap.Logger
}

// NewProductMemoryHandler creates a new instance of ProductMemoryHandler
func NewProductMemoryHandler(repo *repository.ProductMemoryRepository) *ProductMemoryHandler {
	return &ProductMemoryHandler{
		repo:   repo,
		logger: logger.GetLogger(),
	}
}

// CreateProduct handles POST request to create a new product in memory
func (h *ProductMemoryHandler) CreateProduct(c echo.Context) error {
	var req models.ProductRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("Failed to bind product request",
			zap.Error(err),
			zap.String("handler", "CreateProduct (Memory)"),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.logger.Warn("Product validation failed",
			zap.Error(err),
			zap.String("handler", "CreateProduct (Memory)"),
			zap.Any("request", req),
		)
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	// Convert request to product
	product := req.ToProduct()

	// Save to memory
	if err := h.repo.Create(c.Request().Context(), &product); err != nil {
		h.logger.Error("Failed to create product in memory",
			zap.Error(err),
			zap.String("handler", "CreateProduct (Memory)"),
			zap.Any("product", product),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create product"})
	}

	h.logger.Info("Product created successfully in memory",
		zap.String("product_id", product.ID.String()),
		zap.String("product_name", product.Name),
	)

	return c.JSON(http.StatusCreated, product)
}

// GetProduct handles GET request to retrieve a specific product from memory
func (h *ProductMemoryHandler) GetProduct(c echo.Context) error {
	// Parse product ID from URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid product ID",
			zap.Error(err),
			zap.String("handler", "GetProduct (Memory)"),
			zap.String("input_id", idStr),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// Retrieve product from memory
	product, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to retrieve product from memory",
			zap.Error(err),
			zap.String("handler", "GetProduct (Memory)"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
	}

	h.logger.Info("Product retrieved successfully from memory",
		zap.String("product_id", product.ID.String()),
		zap.String("product_name", product.Name),
	)

	return c.JSON(http.StatusOK, product)
}

// ListProducts handles GET request to list products from memory
func (h *ProductMemoryHandler) ListProducts(c echo.Context) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Retrieve products from memory
	products, err := h.repo.List(c.Request().Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to retrieve products from memory",
			zap.Error(err),
			zap.String("handler", "ListProducts (Memory)"),
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve products"})
	}

	// Get total count for pagination metadata
	totalCount, err := h.repo.Count(c.Request().Context())
	if err != nil {
		h.logger.Warn("Failed to retrieve total product count from memory",
			zap.Error(err),
			zap.String("handler", "ListProducts (Memory)"),
		)
		totalCount = 0
	}

	h.logger.Info("Products listed successfully from memory",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.Int("total_count", totalCount),
		zap.Int("returned_count", len(products)),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products":   products,
		"page":       page,
		"pageSize":   pageSize,
		"totalCount": totalCount,
	})
}

// GetAllProducts handles GET request to retrieve all products without pagination from memory
func (h *ProductMemoryHandler) GetAllProducts(c echo.Context) error {
	// Retrieve all products from memory
	products, err := h.repo.GetAll(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to retrieve all products from memory",
			zap.Error(err),
			zap.String("handler", "GetAllProducts (Memory)"),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve products"})
	}

	totalCount := len(products)

	h.logger.Info("All products retrieved successfully from memory",
		zap.Int("total_count", totalCount),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products":   products,
		"page":       1,
		"pageSize":   totalCount,
		"totalCount": totalCount,
	})
}

// UpdateProduct handles PUT request to update a product in memory
func (h *ProductMemoryHandler) UpdateProduct(c echo.Context) error {
	// Parse product ID from URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid product ID",
			zap.Error(err),
			zap.String("handler", "UpdateProduct (Memory)"),
			zap.String("input_id", idStr),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// Parse request body
	var req models.ProductRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("Failed to bind update request",
			zap.Error(err),
			zap.String("handler", "UpdateProduct (Memory)"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.logger.Warn("Product update validation failed",
			zap.Error(err),
			zap.String("handler", "UpdateProduct (Memory)"),
			zap.String("product_id", id.String()),
			zap.Any("request", req),
		)
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	// Update product in memory
	if err := h.repo.Update(c.Request().Context(), id, &req); err != nil {
		h.logger.Error("Failed to update product in memory",
			zap.Error(err),
			zap.String("handler", "UpdateProduct (Memory)"),
			zap.String("product_id", id.String()),
			zap.Any("request", req),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update product"})
	}

	// Retrieve updated product
	updatedProduct, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to retrieve updated product from memory",
			zap.Error(err),
			zap.String("handler", "UpdateProduct (Memory)"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated product"})
	}

	h.logger.Info("Product updated successfully in memory",
		zap.String("product_id", updatedProduct.ID.String()),
		zap.String("product_name", updatedProduct.Name),
	)

	return c.JSON(http.StatusOK, updatedProduct)
}

// DeleteProduct handles DELETE request to remove a product from memory
func (h *ProductMemoryHandler) DeleteProduct(c echo.Context) error {
	// Parse product ID from URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid product ID",
			zap.Error(err),
			zap.String("handler", "DeleteProduct (Memory)"),
			zap.String("input_id", idStr),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// Delete product from memory
	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		h.logger.Error("Failed to delete product from memory",
			zap.Error(err),
			zap.String("handler", "DeleteProduct (Memory)"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete product"})
	}

	h.logger.Info("Product deleted successfully from memory",
		zap.String("product_id", id.String()),
	)

	return c.JSON(http.StatusOK, map[string]string{"message": "Product deleted successfully from memory"})
}

// BulkGenerateProducts handles POST request to generate random products in memory
func (h *ProductMemoryHandler) BulkGenerateProducts(c echo.Context) error {
	// Parse number of products to generate
	count, err := strconv.Atoi(c.QueryParam("count"))
	if err != nil || count < 1 || count > 10000 {
		count = 1000 // Default to 1000 if invalid
	}

	h.logger.Info("Generating bulk products in memory",
		zap.Int("product_count", count),
	)

	// Generate and save products in memory
	if err := h.repo.GenerateAndSaveBulkProducts(c.Request().Context(), count); err != nil {
		h.logger.Error("Failed to generate products in memory",
			zap.Error(err),
			zap.String("handler", "BulkGenerateProducts (Memory)"),
			zap.Int("product_count", count),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate products"})
	}

	// Get total count after generation
	totalCount, err := h.repo.Count(c.Request().Context())
	if err != nil {
		h.logger.Warn("Failed to retrieve total product count after bulk generation from memory",
			zap.Error(err),
			zap.String("handler", "BulkGenerateProducts (Memory)"),
		)
		totalCount = 0
	}

	h.logger.Info("Bulk product generation completed in memory",
		zap.Int("generated_count", count),
		zap.Int("total_count", totalCount),
	)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":    "Products generated successfully in memory",
		"count":      count,
		"totalCount": totalCount,
	})
}

// DeleteAllProducts handles DELETE request to remove all products from memory
func (h *ProductMemoryHandler) DeleteAllProducts(c echo.Context) error {
	h.logger.Warn("Attempting to delete all products from memory")

	// Delete all products from memory
	if err := h.repo.DeleteAll(c.Request().Context()); err != nil {
		h.logger.Error("Failed to delete all products from memory",
			zap.Error(err),
			zap.String("handler", "DeleteAllProducts (Memory)"),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete all products"})
	}

	h.logger.Info("All products deleted successfully from memory")

	return c.JSON(http.StatusOK, map[string]string{"message": "All products deleted successfully from memory"})
}

// GetProductCount handles GET request to retrieve the total number of products in memory
func (h *ProductMemoryHandler) GetProductCount(c echo.Context) error {
	totalCount, err := h.repo.Count(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to retrieve product count from memory",
			zap.Error(err),
			zap.String("handler", "GetProductCount (Memory)"),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve product count"})
	}

	h.logger.Info("Product count retrieved successfully from memory",
		zap.Int("total_count", totalCount),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"totalCount": totalCount,
	})
}
