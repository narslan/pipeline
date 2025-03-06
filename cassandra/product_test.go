package cassandra_test

import (
	"context"
	"reflect"
	"testing"

	dataflow "github.com/narslan/pipeline"
	"github.com/narslan/pipeline/cassandra"
	"github.com/narslan/pipeline/container"
)

func TestProductService_CreateProduct(t *testing.T) {
	// Start containers for test.
	ctx := context.Background()
	cdbc, cassandraConnectionHost := container.MustDeployCassandra(ctx)
	defer container.MustCleanCassandraContainer(ctx, cdbc)
	//Ensure product can be created.
	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t, cassandraConnectionHost)
		defer MustCloseDB(t, db)

		c := cassandra.NewProductService(db)

		p := &dataflow.Product{
			ID:          1,
			Title:       "title1",
			Price:       42.01,
			Category:    "bilgisayar",
			Brand:       "brand1",
			URL:         "https://url1.com",
			Description: "a description",
		}

		// Create new product.
		err := c.CreateProduct(context.Background(), p)
		if err != nil {
			t.Fatal(err)
		}

		p2 := &dataflow.Product{
			ID:       2,
			Title:    "title2",
			Price:    42.0,
			Category: "bilgisayar",
			Brand:    "brand2",
		}

		// Create second product.
		err = c.CreateProduct(context.Background(), p2)
		if err != nil {
			t.Fatal(err)
		}

		// Fetch product from database and compare.
		if other, err := c.FindProductByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(p, other) {
			t.Fatalf("mismatch: %#v != %#v", p, other)
		}
	})

	// Ensure an error is returned if product id is not set.
	t.Run("ErrIDRequired", func(t *testing.T) {
		db := MustOpenDB(t, cassandraConnectionHost)
		defer MustCloseDB(t, db)

		s := cassandra.NewProductService(db)
		err := s.CreateProduct(context.Background(), &dataflow.Product{})

		if err == nil {
			t.Fatal("expected error")
		} else if dataflow.ErrorCode(err) != dataflow.EINVALID || dataflow.ErrorMessage(err) != `ID must be greater than 0.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure an error is returned if product title is not set.
	t.Run("ErrTitleRequired", func(t *testing.T) {
		db := MustOpenDB(t, cassandraConnectionHost)
		defer MustCloseDB(t, db)

		s := cassandra.NewProductService(db)
		err := s.CreateProduct(context.Background(), &dataflow.Product{ID: 4})

		if err == nil {
			t.Fatal("expected error")
		} else if dataflow.ErrorCode(err) != dataflow.EINVALID || dataflow.ErrorMessage(err) != `Title must not be empty.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure an error is returned if product price is negative.
	t.Run("ErrPriceNegative", func(t *testing.T) {
		db := MustOpenDB(t, cassandraConnectionHost)
		defer MustCloseDB(t, db)

		s := cassandra.NewProductService(db)
		err := s.CreateProduct(context.Background(), &dataflow.Product{ID: 3, Title: "title3", Price: -4.2})

		if err == nil {
			t.Fatal("expected error")
		} else if dataflow.ErrorCode(err) != dataflow.EINVALID || dataflow.ErrorMessage(err) != `Price must be greater than zero.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestProductService_FindProduct(t *testing.T) {
	// Ensure an error is returned if fetching a non-existent product.
	// Start containers for test.
	ctx := context.Background()
	cdbc, cassandraConnectionHost := container.MustDeployCassandra(ctx)
	defer container.MustCleanCassandraContainer(ctx, cdbc)

	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t, cassandraConnectionHost)
		defer MustCloseDB(t, db)
		s := cassandra.NewProductService(db)
		_, err := s.FindProductByID(context.Background(), 20)

		if dataflow.ErrorCode(err) != dataflow.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
