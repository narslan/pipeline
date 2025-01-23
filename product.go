package dataflow

import (
	"context"
)

// Product represents a product in the database.
// Products are created by a job pipeline.
type Product struct {
	ID          uint32  `json:"id,omitempty"`
	Title       string  `json:"title,omitempty"`
	Price       float32 `json:"price,omitempty"`
	Category    string  `json:"category,omitempty"`
	Brand       string  `json:"brand,omitempty"`
	URL         string  `json:"url,omitempty"`
	Description string  `json:"description,omitempty"`
}

// Validate returns an error if the product contains invalid fields.
// This only performs basic validation.
func (p *Product) Validate() error {
	if p.ID == 0 {
		return Errorf(EINVALID, "ID must be greater than 0.")
	}
	if p.Title == "" {
		return Errorf(EINVALID, "Title must not be empty.")
	}
	if p.Price <= 0.0 {
		return Errorf(EINVALID, "Price must be greater than zero.")
	}
	if p.Category == "" {
		return Errorf(EINVALID, "Category must not be empty.")
	}
	if p.Brand == "" {
		return Errorf(EINVALID, "Brand must not be empty.")
	}

	return nil
}

// ProductService represents a service for managing products.
type ProductService interface {

	// Retrieves a product by ID.
	// Returns ENOTFOUND if product does not exist.
	FindProductByID(ctx context.Context, id uint32) (*Product, error)

	// Creates a new product.
	CreateProduct(ctx context.Context, p *Product) error
}
