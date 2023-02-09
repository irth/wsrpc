package wsrpc

import (
	"encoding/json"
	"net/http"
)

// Response defines the structure of the RPC response.
type Response[ResponseT any] struct {
	CommandMeta
	OK       bool      `json:"ok"`
	Error    string    `json:"error,omitempty"`
	Response ResponseT `json:"response,omitempty"`
}

// WriteHTTPError can be used to write a HTTP error response with structure that
// matches the one used by the websocket.
func WriteHTTPError(w http.ResponseWriter, statusCode int, message string) error {
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(Response[any]{
		OK:    false,
		Error: message,
	})
}
