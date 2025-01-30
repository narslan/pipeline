package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/narslan/pipeline"
)

// Ensure the HTTP server can return the product.
func TestGetProduct(t *testing.T) {
	// Start the mocked HTTP test server.
	s := MustOpenServer(t)
	defer MustCloseServer(t, s)

	//Mock product data.

	p := &dataflow.Product{
		ID:          uint32(1),
		Title:       "title1",
		Price:       42.01,
		Category:    "bilgisayar",
		Brand:       "brand1",
		URL:         "https://url1.com",
		Description: "a description",
	}

	// Mock the fetch of product.
	s.ProductService.FindProductByIDFn = func(ctx context.Context, id uint32) (*dataflow.Product, error) {
		return p, nil
	}

	// Ensure server can generate JSON output.
	t.Run("JSON", func(t *testing.T) {
		// Mock the fetch of product.
		s.ProductService.FindProductByIDFn = func(ctx context.Context, id uint32) (*dataflow.Product, error) {
			return p, nil
		}

		// Issue request with product id.
		resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.TODO(), "GET", "/product/1", nil))
		if err != nil {
			t.Fatal(err)
		} else if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Fatalf("StatusCode=%v, want %v", got, want)
		}
	})

}
