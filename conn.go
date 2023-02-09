package wsrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Conn struct {
	w *wsutil.Writer
	r *wsutil.Reader

	encoder *json.Encoder
	decoder *json.Decoder

	conn net.Conn

	CommandDecoder

	readLock  sync.Mutex
	writeLock sync.Mutex
}

// NewConn upgrades the connection to a WebSocket and returns a *Conn which can
// be used for further communication. The palette argument is used to register
// command types that the server is supposed to understand.
func NewConn(w http.ResponseWriter, r *http.Request, palette CommandPalette) (*Conn, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)

	if err != nil {
		return nil, fmt.Errorf("ws upgrade: %w", err)
	}

	sw := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
	sr := wsutil.NewReader(conn, ws.StateServerSide)
	encoder := json.NewEncoder(sw)
	decoder := json.NewDecoder(sr)

	wsconn := &Conn{}

	cmdDecoder := CommandDecoder{
		palette: palette,
		wsconn:  wsconn,
	}

	wsconn.conn = conn
	wsconn.w = sw
	wsconn.r = sr
	wsconn.encoder = encoder
	wsconn.decoder = decoder
	wsconn.CommandDecoder = cmdDecoder

	return wsconn, nil
}

// Close closes the underlying websocket connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// SendMessage pushes a message (any type implementing the Message interface) to
// the client.
// SendMessage can be used from multiple threads simultaneously.
func (c *Conn) SendMessage(m Message) error {
	return c.sendRaw(messageWrapper{
		Type:    m.Type(),
		Message: m,
	})
}

func (c *Conn) recvRaw(obj interface{}) error {
	c.readLock.Lock()
	defer c.readLock.Unlock()

	hdr, err := c.r.NextFrame()
	if err != nil {
		return fmt.Errorf("ws NextFrame: %w", err)
	}
	if hdr.OpCode == ws.OpClose {
		return io.EOF
	}

	err = c.decoder.Decode(obj)
	if err != nil {
		return fmt.Errorf("json decode: %w", err)
	}

	return nil
}

func (c *Conn) sendRaw(obj interface{}) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	err := c.encoder.Encode(obj)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	if err = c.w.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}
