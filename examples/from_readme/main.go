package main

import (
	"log"
	"net/http"

	"github.com/irth/wsrpc"
)

type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// doesn't have to be a struct, it can be anything that serializes to JSON
type SumResponse struct {
	Sum int `json:"sum"`
}

type SumCommand = wsrpc.Command[SumRequest, SumResponse]

// the request and response types do not have to be structs
type NegateCommand = wsrpc.Command[int, int]

// define the mapping from type names to Golang types
var wsCommands = wsrpc.CommandPalette{
	"sum":    SumCommand{},
	"negate": NegateCommand{},
}

// now you can use wsrpc.NewConn in a HTTP request handler
// the connection upgrade to a websocket will be handled for you
func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsrpc.NewConn(w, r, wsCommands)
	if err != nil {
		wsrpc.WriteHTTPError(w, 500, "failed to upgrade connection")
		return
	}
	defer conn.Close() // make sure to clean up after yourself :)

	for {
		cmd, err := conn.Decode()
		if err != nil {
			log.Printf("error: %s", err.Error())
			return
		}

		switch cmd := cmd.(type) {
		case SumCommand:
			sum := cmd.Request.A + cmd.Request.B
			if sum == 0x09F91102 {
				// 09 F9 11 02 is the beginning of the AACS encryption
				// key[1], which is considered to be an illegal number[2]

				// 1. https://en.wikipedia.org/wiki/AACS_encryption_key_controversy
				// 2. https://en.wikipedia.org/wiki/Illegal_number

				// cmd.Err will reply with all the IDs etc set correctly, as
				// described in the JSON examples above
				cmd.Err("illegal number detected")
				continue
			}

			// cmd.Ok also fills the IDs and such automatically
			cmd.OK(SumResponse{sum})
		case NegateCommand:
			cmd.OK(-cmd.Request)

		// if a command is not defined in the wsCommands palette, wsrpc will
		// respond with an error, so we don't need a default case, but one
		// might include it anyway to ensure unhandled commands that are in
		// the palette also get a response
		default:
			cmd.Err("not implemented")
		}
	}
}

func main() {
	http.HandleFunc("/ws", WebsocketHandler)
	panic(http.ListenAndServe(":8080", nil))
}
