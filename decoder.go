package wsrpc

import "fmt"

// CommandPalette maps type names to actual Command types.
// Use the zero value of your command types as the values.
type CommandPalette map[string]fromRawable

// Decode reads and decodes a single command from the websocket in a blocking
// manner.
// Decode can be called safely from multiple threads at once.
func (c *CommandDecoder) Decode() (Errable, error) {
	rawCommand := RawCommand{}

	err := c.wsconn.recvRaw(&rawCommand)
	if err != nil {
		return nil, err
	}

	rawCommand.wsconn = c.wsconn

	cmdType, ok := c.palette[rawCommand.Command]
	if !ok {
		rawCommand.Err("unknown command")
		return nil, fmt.Errorf("unknown command: %s", rawCommand.Command)
	}

	cmd, err := cmdType.FromRaw(rawCommand)
	if err != nil {
		return nil, fmt.Errorf("from raw: %w", err)
	}

	return cmd, nil
}

// CommandDecoder handles decoding the commands.
type CommandDecoder struct {
	wsconn  *Conn
	palette CommandPalette
}
