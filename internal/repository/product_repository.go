package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"product-service/internal/models"
	"product-service/pkg/utils"
)

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new repository instance
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products 
		(id, name, description, price, created_at, updated_at) 
		VALUES (:id, :name, :description, :price, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

// CreateBulk inserts multiple products in a single transaction
func (r *ProductRepository) CreateBulk(ctx context.Context, products []models.Product) error {
	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	// Prepare the query
	query := `
		INSERT INTO products 
		(id, name, description, price, created_at, updated_at) 
		VALUES (:id, :name, :description, :price, :created_at, :updated_at)
	`

	// Execute bulk insert
	for _, product := range products {
		_, err := tx.NamedExecContext(ctx, query, product)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// GetByID retrieves a product by its UUID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	query := `SELECT * FROM products WHERE id = $1`

	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// List retrieves all products with optional pagination
func (r *ProductRepository) List(ctx context.Context, page, pageSize int) ([]models.Product, error) {
	var products []models.Product
	query := `
		SELECT * FROM products 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &products, query, pageSize, offset)
	return products, err
}

// GetAll retrieves all products without pagination
func (r *ProductRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	query := `SELECT * FROM products ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &products, query)
	return products, err
}

// Update modifies an existing product
func (r *ProductRepository) Update(ctx context.Context, id uuid.UUID, req *models.ProductRequest) error {
	query := `
		UPDATE products 
		SET name = $1, 
			description = $2, 
			price = $3, 
			updated_at = $4 
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		req.Name,
		req.Description,
		req.Price,
		time.Now(),
		id,
	)
	return err
}

// Delete removes a product by its ID
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no product found with the given ID")
	}

	return nil
}

// DeleteAll removes all products from the database
func (r *ProductRepository) DeleteAll(ctx context.Context) error {
	query := `DELETE FROM products`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// GenerateAndSaveBulkProducts creates a specified number of random products
func (r *ProductRepository) GenerateAndSaveBulkProducts(ctx context.Context, count int) error {
	// Generate random products
	randomProducts := utils.GenerateRandomProducts(count)

	// Convert to models.Product
	products := make([]models.Product, len(randomProducts))
	for i, rp := range randomProducts {
		products[i] = models.Product{
			ID:          rp.ID,
			Name:        rp.Name,
			Description: rp.Description,
			Price:       rp.Price,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	// Bulk insert
	return r.CreateBulk(ctx, products)
}

// Count returns the total number of products in the database
func (r *ProductRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("error counting products: %v", err)
	}
	return count, nil
}
