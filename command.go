package wsrpc

import (
	"encoding/json"
	"fmt"
)

// CommandMeta describes command metadata - request ID and the type name
type CommandMeta struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

// Command is used in tandem with CommandPalette to define your commands
type Command[RequestT any, ReplyT any] struct {
	CommandMeta
	Request RequestT `json:"request"`

	wsconn *Conn
}

// RawCommand's request won't be decoded, and it's response can be anything
type RawCommand = Command[json.RawMessage, any]

type fromRawable interface {
	FromRaw(r RawCommand) (Errable, error)
}

type Errable interface {
	Err(format string, args ...interface{}) error
}

// FromRaw tries to create a command with the same type as the receiver, by
// decoding a RawCommand struct. This is mostly an implementation detail which
// makes it possible to instantiate commands with the correct types, based on a
// CommandPalette.
func (c Command[RequestT, ReplyT]) FromRaw(r RawCommand) (Errable, error) {
	cmd := Command[RequestT, ReplyT]{
		CommandMeta: r.CommandMeta,

		wsconn: r.wsconn,
	}

	err := json.Unmarshal(r.Request, &cmd.Request)
	if err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	return cmd, nil
}

// OK sends a success response, taking care of setting the correct ID and
// Command values.
func (c Command[RequestT, ReplyT]) OK(reply ReplyT) error {
	response := Response[ReplyT]{
		CommandMeta: c.CommandMeta,
		OK:          true,
		Response:    reply,
	}
	return c.wsconn.sendRaw(response)
}

// Err sends an error response, taking care of setting the correct ID and
// Command values.
func (c Command[RequestT, ReplyT]) Err(format string, args ...interface{}) error {
	response := Response[any]{
		CommandMeta: c.CommandMeta,
		OK:          false,
		Error:       fmt.Sprintf(format, args...),
	}

	return c.wsconn.sendRaw(response)
}
