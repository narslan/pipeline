package http

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/nevroz/dataflow"
)

// Error prints & optionally logs an error message.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// Extract error code & message.
	code, message := dataflow.ErrorCode(err), dataflow.ErrorMessage(err)

	// Log internal errors. Here is no-op
	if code == dataflow.EINTERNAL {
		LogError(r, err)
	}

	// Print user message to response.
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(ErrorStatusCode(code))

	if err := json.NewEncoder(w).Encode(&ErrorResponse{Error: message}); err != nil {
		log.Println("failed to send reposnse:", err)
	}

}

// ErrorResponse represents a JSON structure for error output.
type ErrorResponse struct {
	Error string `json:"error"`
}

// LogError logs an error with the HTTP route information.
func LogError(r *http.Request, err error) {
	log.Printf("[http] error: %s %s: %s", r.Method, r.URL.Path, err)
}

// lookup of application error codes to HTTP status codes.
var codes = map[string]int{

	dataflow.EINVALID:  http.StatusBadRequest,
	dataflow.ENOTFOUND: http.StatusNotFound,
	dataflow.EINTERNAL: http.StatusInternalServerError,
}

// ErrorStatusCode returns the associated HTTP status code for a dataflow error code.
func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// FromErrorStatusCode returns the associated dataflow code for an HTTP status code.
func FromErrorStatusCode(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return dataflow.EINTERNAL
}
