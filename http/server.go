package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/narslan/pipeline"
)

// ShutdownTimeout is the period for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

// Server represents an HTTP server.
type Server struct {
	server *http.Server

	// Bind address for the server's listener as in ":8080".
	Address string

	// Service used by the HTTP routes.
	ProductService dataflow.ProductService
}

// NewServer returns a new instance of Server.
func NewServer() *Server {

	// Create a router using standard library.
	mux := http.NewServeMux()

	// Create a new server that wraps the net/http server and standard mux.
	s := &Server{
		server: &http.Server{
			Handler: mux,
		},
	}

	// Setup our handler that gets product from .
	mux.HandleFunc("GET /product/{id}", s.getProductById)
	return s
}

// Open begins listening on the bind address.
func (s *Server) Open() (err error) {

	// Assign server port.
	s.server.Addr = s.Address

	go func() {
		log.Println("microservice server listens on", s.server.Addr)
		err := s.server.ListenAndServe()

		if err != http.ErrServerClosed {
			// it is fine to use Fatal here because it is not main gorutine
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()
	return nil
}

// Close shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	log.Print("shutting down server")
	return s.server.Shutdown(ctx)
}
