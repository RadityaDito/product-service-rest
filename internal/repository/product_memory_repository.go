package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"product-service/internal/models"
	"product-service/pkg/logger"
	"product-service/pkg/utils"
)

// ProductMemoryRepository handles in-memory operations for products
type ProductMemoryRepository struct {
	products []models.Product
	mutex    sync.RWMutex
	logger   *zap.Logger
}

// NewProductMemoryRepository creates a new in-memory repository instance
func NewProductMemoryRepository() *ProductMemoryRepository {
	return &ProductMemoryRepository{
		products: make([]models.Product, 0),
		logger:   logger.GetLogger(),
	}
}

// Create adds a new product to the in-memory storage
func (r *ProductMemoryRepository) Create(ctx context.Context, product *models.Product) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.products = append(r.products, *product)
	return nil
}

// CreateBulk adds multiple products to the in-memory storage
func (r *ProductMemoryRepository) CreateBulk(ctx context.Context, products []models.Product) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.products = append(r.products, products...)
	return nil
}

// GetByID retrieves a product by its UUID
func (r *ProductMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, product := range r.products {
		if product.ID == id {
			productCopy := product // Create a copy to avoid race conditions
			return &productCopy, nil
		}
	}
	return nil, errors.New("product not found")
}

// List retrieves products with pagination
func (r *ProductMemoryRepository) List(ctx context.Context, page, pageSize int) ([]models.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Calculate start and end indices for pagination
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	// Check if startIndex is valid
	if startIndex >= len(r.products) {
		return []models.Product{}, nil
	}

	// Check if endIndex is valid
	if endIndex > len(r.products) {
		endIndex = len(r.products)
	}

	// Create a copy of the slice to prevent race conditions
	result := make([]models.Product, endIndex-startIndex)
	copy(result, r.products[startIndex:endIndex])
	return result, nil
}

// GetAll retrieves all products without pagination
func (r *ProductMemoryRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Return a copy of the products slice to prevent race conditions
	result := make([]models.Product, len(r.products))
	copy(result, r.products)
	return result, nil
}

// Update modifies an existing product
func (r *ProductMemoryRepository) Update(ctx context.Context, id uuid.UUID, req *models.ProductRequest) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, product := range r.products {
		if product.ID == id {
			// Update product fields
			r.products[i].Name = req.Name
			r.products[i].Description = req.Description
			r.products[i].Price = req.Price
			r.products[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return errors.New("product not found")
}

// Delete removes a product by its ID
func (r *ProductMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, product := range r.products {
		if product.ID == id {
			// Remove the product by swapping with the last element and truncating
			r.products[i] = r.products[len(r.products)-1]
			r.products = r.products[:len(r.products)-1]
			return nil
		}
	}
	return errors.New("product not found")
}

// DeleteAll removes all products
func (r *ProductMemoryRepository) DeleteAll(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.products = make([]models.Product, 0)
	return nil
}

// GenerateAndSaveBulkProducts creates a specified number of random products
func (r *ProductMemoryRepository) GenerateAndSaveBulkProducts(ctx context.Context, count int) error {
	// Generate random products
	randomProducts := utils.GenerateRandomProducts(count)

	// Convert to models.Product
	products := make([]models.Product, len(randomProducts))
	now := time.Now()
	for i, rp := range randomProducts {
		products[i] = models.Product{
			ID:          rp.ID,
			Name:        rp.Name,
			Description: rp.Description,
			Price:       rp.Price,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	// Add to in-memory storage
	return r.CreateBulk(ctx, products)
}

// Count returns the total number of products
func (r *ProductMemoryRepository) Count(ctx context.Context) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.products), nil
}
