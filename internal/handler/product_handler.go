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

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	repo   *repository.ProductRepository
	logger *zap.Logger
}

// NewProductHandler creates a new instance of ProductHandler
func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		repo:   repo,
		logger: logger.GetLogger(),
	}
}

// CreateProduct handles POST request to create a new product
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	var req models.ProductRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("Failed to bind product request",
			zap.Error(err),
			zap.String("handler", "CreateProduct"),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.logger.Warn("Product validation failed",
			zap.Error(err),
			zap.String("handler", "CreateProduct"),
			zap.Any("request", req),
		)
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	// Convert request to product
	product := req.ToProduct()

	// Save to database
	if err := h.repo.Create(c.Request().Context(), &product); err != nil {
		h.logger.Error("Failed to create product",
			zap.Error(err),
			zap.String("handler", "CreateProduct"),
			zap.Any("product", product),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create product"})
	}

	h.logger.Info("Product created successfully",
		zap.String("product_id", product.ID.String()),
		zap.String("product_name", product.Name),
	)

	return c.JSON(http.StatusCreated, product)
}

// GetProduct handles GET request to retrieve a specific product
func (h *ProductHandler) GetProduct(c echo.Context) error {
	// Parse product ID from URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid product ID",
			zap.Error(err),
			zap.String("handler", "GetProduct"),
			zap.String("input_id", idStr),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// Retrieve product
	product, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to retrieve product",
			zap.Error(err),
			zap.String("handler", "GetProduct"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
	}

	h.logger.Info("Product retrieved successfully",
		zap.String("product_id", product.ID.String()),
		zap.String("product_name", product.Name),
	)

	return c.JSON(http.StatusOK, product)
}

// ListProducts handles GET request to list products
func (h *ProductHandler) ListProducts(c echo.Context) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Retrieve products
	products, err := h.repo.List(c.Request().Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to retrieve products",
			zap.Error(err),
			zap.String("handler", "ListProducts"),
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve products"})
	}

	// // Get total count for pagination metadata
	// totalCount, err := h.repo.Count(c.Request().Context())
	// if err != nil {
	// 	h.logger.Warn("Failed to retrieve total product count",
	// 		zap.Error(err),
	// 		zap.String("handler", "ListProducts"),
	// 	)
	// 	totalCount = 0
	// }

	h.logger.Info("Products listed successfully",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		// zap.Int("total_count", totalCount),
		zap.Int("returned_count", len(products)),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products": products,
		"page":     page,
		"pageSize": pageSize,
		// "totalCount": totalCount,
	})
}

// GetAllProducts handles GET request to retrieve all products without pagination
func (h *ProductHandler) GetAllProducts(c echo.Context) error {
	// Retrieve all products
	products, err := h.repo.List(c.Request().Context(), 1, 100000) // Arbitrary large page size
	if err != nil {
		h.logger.Error("Failed to retrieve all products",
			zap.Error(err),
			zap.String("handler", "GetAllProducts"),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve products"})
	}

	h.logger.Info("All products retrieved successfully",
		zap.Int("total_count", len(products)),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products": products,
	})
}

// UpdateProduct handles PUT request to update a product
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	// Parse product ID from URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid product ID",
			zap.Error(err),
			zap.String("handler", "UpdateProduct"),
			zap.String("input_id", idStr),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// Parse request body
	var req models.ProductRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("Failed to bind update request",
			zap.Error(err),
			zap.String("handler", "UpdateProduct"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.logger.Warn("Product update validation failed",
			zap.Error(err),
			zap.String("handler", "UpdateProduct"),
			zap.String("product_id", id.String()),
			zap.Any("request", req),
		)
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	// Update product
	if err := h.repo.Update(c.Request().Context(), id, &req); err != nil {
		h.logger.Error("Failed to update product",
			zap.Error(err),
			zap.String("handler", "UpdateProduct"),
			zap.String("product_id", id.String()),
			zap.Any("request", req),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update product"})
	}

	// Retrieve updated product
	updatedProduct, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to retrieve updated product",
			zap.Error(err),
			zap.String("handler", "UpdateProduct"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated product"})
	}

	h.logger.Info("Product updated successfully",
		zap.String("product_id", updatedProduct.ID.String()),
		zap.String("product_name", updatedProduct.Name),
	)

	return c.JSON(http.StatusOK, updatedProduct)
}

// DeleteProduct handles DELETE request to remove a product
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	// Parse product ID from URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid product ID",
			zap.Error(err),
			zap.String("handler", "DeleteProduct"),
			zap.String("input_id", idStr),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// Delete product
	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		h.logger.Error("Failed to delete product",
			zap.Error(err),
			zap.String("handler", "DeleteProduct"),
			zap.String("product_id", id.String()),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete product"})
	}

	h.logger.Info("Product deleted successfully",
		zap.String("product_id", id.String()),
	)

	return c.JSON(http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}

// BulkGenerateProducts handles POST request to generate random products
func (h *ProductHandler) BulkGenerateProducts(c echo.Context) error {
	// Parse number of products to generate
	count, err := strconv.Atoi(c.QueryParam("count"))
	if err != nil || count < 1 || count > 10000 {
		count = 1000 // Default to 1000 if invalid
	}

	h.logger.Info("Generating bulk products",
		zap.Int("product_count", count),
	)

	// Generate and save products
	if err := h.repo.GenerateAndSaveBulkProducts(c.Request().Context(), count); err != nil {
		h.logger.Error("Failed to generate products",
			zap.Error(err),
			zap.String("handler", "BulkGenerateProducts"),
			zap.Int("product_count", count),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate products"})
	}

	// Get total count after generation
	totalCount, err := h.repo.Count(c.Request().Context())
	if err != nil {
		h.logger.Warn("Failed to retrieve total product count after bulk generation",
			zap.Error(err),
			zap.String("handler", "BulkGenerateProducts"),
		)
		totalCount = 0
	}

	h.logger.Info("Bulk product generation completed",
		zap.Int("generated_count", count),
		zap.Int("total_count", totalCount),
	)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":    "Products generated successfully",
		"count":      count,
		"totalCount": totalCount,
	})
}

// DeleteAllProducts handles DELETE request to remove all products
func (h *ProductHandler) DeleteAllProducts(c echo.Context) error {
	h.logger.Warn("Attempting to delete all products")

	// Delete all products
	if err := h.repo.DeleteAll(c.Request().Context()); err != nil {
		h.logger.Error("Failed to delete all products",
			zap.Error(err),
			zap.String("handler", "DeleteAllProducts"),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete all products"})
	}

	h.logger.Info("All products deleted successfully")

	return c.JSON(http.StatusOK, map[string]string{"message": "All products deleted successfully"})
}
