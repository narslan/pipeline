package cassandra

import (
	"context"
	"errors"

	"bitbucket.org/nevroz/dataflow"
	"github.com/gocql/gocql"
)

// Ensure service implements interface.
var _ dataflow.ProductService = (*ProductService)(nil)

// ProductService represents a service for managing products.
type ProductService struct {
	db *DB
}

// NewProductService returns a new instance of ProductService.
func NewProductService(db *DB) *ProductService {
	return &ProductService{db: db}
}

// FindProductByID retrieves a product by ID.
// Returns ENOTFOUND if product does not exist.
func (s *ProductService) FindProductByID(ctx context.Context, id uint32) (*dataflow.Product, error) {

	// Prepare query string.
	qryStmt := "SELECT id, title, price, category, brand, url, description  FROM products WHERE id = ?"

	var (
		pid         uint32
		title       string
		price       float32
		category    string
		brand       string
		url         string
		description string
	)

	// Execute query to fetch user rows.
	err := s.db.session.Query(qryStmt, id).WithContext(ctx).Scan(&pid, &title, &price, &category, &brand, &url, &description)

	if err != nil {

		if errors.Is(err, gocql.ErrNotFound) {
			return nil, dataflow.Errorf(dataflow.ENOTFOUND, "product with id: %d is not found", id)
		} else {
			return nil, dataflow.Errorf(dataflow.EINTERNAL, "query by %d failed with err: %v", id, err)
		}

	}

	return &dataflow.Product{
		ID:          pid,
		Title:       title,
		Price:       price,
		Category:    category,
		Brand:       brand,
		URL:         url,
		Description: description,
	}, nil

}

// CreateProduct creates a new product.
func (s *ProductService) CreateProduct(ctx context.Context, p *dataflow.Product) error {

	//Validate input
	err := p.Validate()
	if err != nil {
		return err
	}

	// Prepare insert statement.
	insSt := `INSERT INTO products(id, title, price, category, brand, url, description) VALUES(?,?,?,?,?,?,?)`

	// Execute query to insert product values.
	return s.db.session.Query(insSt, p.ID, p.Title, p.Price, p.Category, p.Brand, p.URL, p.Description).WithContext(ctx).Exec()
}
