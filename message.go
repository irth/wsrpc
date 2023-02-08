package wsrpc

type messageWrapper struct {
	Type    string      `json:"type"`
	Message interface{} `json:"message"`
}

type Message interface {
	Type() string
}
