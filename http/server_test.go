package http_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	dataflowhttp "github.com/narslan/dataflow/http"
	"github.com/narslan/dataflow/mock"
)

// Server represents a test wrapper for dataflowhttp.Server.
// It attaches product mock to the server & initializes on a random port.
type Server struct {
	*dataflowhttp.Server

	// Mock services.
	ProductService mock.ProductService
}

// MustOpenServer is  test helper function for starting a new test HTTP server.
// Fail on error.
func MustOpenServer(tb testing.TB) *Server {
	tb.Helper()

	// Initialize wrapper.
	s := &Server{Server: dataflowhttp.NewServer()}
	s.Address = ":8080"
	// Assign mocks to actual server's services.
	s.Server.ProductService = &s.ProductService

	// Begin running test server.
	if err := s.Open(); err != nil {
		tb.Fatal(err)
	}
	return s

}

// MustCloseServer shut downs the test erver.
// Fail on error.
func MustCloseServer(tb testing.TB, s *Server) {
	tb.Helper()
	if err := s.Close(); err != nil {
		tb.Fatal(err)
	}
}

// MustNewRequest creates a new HTTP request.
func (s *Server) MustNewRequest(tb testing.TB, ctx context.Context, method, url string, body io.Reader) *http.Request {
	tb.Helper()

	// Create new net/http request with server's base URL.
	r, err := http.NewRequest(method, "http://localhost:8080/product/1", body)
	if err != nil {
		tb.Fatal(err)
	}

	return r
}
