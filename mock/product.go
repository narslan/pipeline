package mock

import (
	"context"

	"bitbucket.org/nevroz/dataflow"
)

var _ dataflow.ProductService = (*ProductService)(nil)

type ProductService struct {
	FindProductByIDFn func(ctx context.Context, id uint32) (*dataflow.Product, error)
	CreateProductFn   func(ctx context.Context, p *dataflow.Product) error
}

func (s *ProductService) FindProductByID(ctx context.Context, id uint32) (*dataflow.Product, error) {
	return s.FindProductByIDFn(ctx, id)
}

func (s *ProductService) CreateProduct(ctx context.Context, p *dataflow.Product) error {
	return s.CreateProductFn(ctx, p)
}
