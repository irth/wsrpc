# wsrpc

[![Go Reference](https://pkg.go.dev/badge/github.com/irth/wsrpc.svg)](https://pkg.go.dev/github.com/irth/wsrpc)

A simple package to make it easier to handle RPC over websockets.

Reads/writes should be thread safe.

## Usage

Websocket clients can send commands as JSON messages (see the end of readme for
info about how to push messages to the clients (otherwise you could just use
HTTP requests...)):

```json
{
  "id": "<message id, optional, will appear in the response if set>",
  "command": "<command name>",
  "request": {
    /* request contents, depending on the command, see golang examples below */
  }
}
```

The server responds with:

```json
// success
{
    "id": "<message id from the request>",
    "command": "<command name>",
    "ok": true,
    "response": {
        /* response content, depending on the command, see Golang examples below */
    }
}

// failure
{

    "id": "<message id from the request>",
    "command": "<command name>",
    "ok": false,
    "error": "<error message>"
}
```

To use the package, first, define the commands (this code is also awailable as a
[runnable example](./examples/from_readme/main.go)):

```go
type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// doesn't have to be a struct, it can be anything that serializes to JSON
type SumResponse struct {
	Sum int `json:"sum"`
}

type SumCommand = wsrpc.Command[SumRequest, SumResponse]
```

```go
// the request and response types do not have to be structs
type NegateCommand = wsrpc.Command[int, int]

// define the mapping from type names to Golang types
var wsCommands = wsrpc.CommandPalette{
	"sum":    SumCommand{},
	"negate": NegateCommand{},
}
```

Then, set up an HTTP handler:

```go
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
```

Aaaand, you're done!

```go
func main() {
	http.HandleFunc("/ws", WebsocketHandler)
	panic(http.ListenAndServe(":8080", nil))
}
```

Example communication:

```
client: {"command": "negate", "request": 2137}
server: {"id":"","command":"negate","ok":true,"response":-2137}
```

```
client: {"command": "sum", "request": {"a": 1, "b": 2}}
server: {"id":"","command":"sum","ok":true,"response":{"sum":3}}
```

```
client: {"command": "sum", "request": {"a": 0, "b": 167317762}}
server: {"id":"","command":"sum","ok":false,"error":"illegal number detected"}
```

You can also push messages without a request from the client:

```go

type HelloMessage struct {
    Hiiii string `json:"hiiii"`
}

// implement the wsrpc.Message interface
func (h HelloMessage) Type() string { return "hello" }

// ...in the ws handler function
conn.SendMessage(HelloMessage{Hiiii: ":3"})
```

This would send the following JSON message to the client:

```json
{ "type": "hello", "message": { "hiiii": ":3" } }
```
