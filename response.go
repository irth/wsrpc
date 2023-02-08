package wsrpc

import (
	"encoding/json"
	"net/http"
)

type Response[ResponseT any] struct {
	CommandMeta
	OK       bool      `json:"ok"`
	Error    string    `json:"error,omitempty"`
	Response ResponseT `json:"response,omitempty"`
}

func WriteHTTPError(w http.ResponseWriter, statusCode int, message string) error {
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(Response[any]{
		OK:    false,
		Error: message,
	})
}
