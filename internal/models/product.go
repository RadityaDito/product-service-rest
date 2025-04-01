package models

import (
	"time"

	"github.com/google/uuid"
)

// Product represents the product structure
type Product struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,min=3,max=255"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price" validate:"required,min=0"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ProductRequest represents the input for creating/updating a product
type ProductRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

// ToProduct converts ProductRequest to Product
func (pr *ProductRequest) ToProduct() Product {
	now := time.Now()
	return Product{
		ID:          uuid.New(),
		Name:        pr.Name,
		Description: pr.Description,
		Price:       pr.Price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
