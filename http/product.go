package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bitbucket.org/nevroz/dataflow"
)

func (s *Server) getProductById(w http.ResponseWriter, r *http.Request) {

	// Parse ID from path.
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		Error(w, r, dataflow.Errorf(dataflow.EINVALID, "Invalid ID format"))
		return
	}

	// Fetch product from the database.
	p, err := s.ProductService.FindProductByID(r.Context(), uint32(id))
	if err != nil {
		Error(w, r, err)
		return
	}

	// Provide content type in response header.
	w.Header().Set("Content-Type", "aplication/json")

	// Encode the model as a JSON string.
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		LogError(r, err)
		return
	}

}
