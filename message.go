package wsrpc

type messageWrapper struct {
	Type    string      `json:"type"`
	Message interface{} `json:"message"`
}

// Message needs to be implemented by any type that the server wants to push to
// the client.
type Message interface {
	Type() string
}
