package wsrpc

import (
	"context"

	"github.com/irth/chanutil"
)

// Pump represents a pump which pumps commands from the websocket to a Go
// channel.
type Pump interface {
	// Ch returns the channel to which the commands get pumped. If an error is
	// encountered during decoding, the channel is closed and Pump.Err() returns
	// the error.
	Ch() <-chan Errable

	// Err returns the error encountered when decoding, if any.
	Err() error
}

type pump struct {
	decoder *CommandDecoder
	ch      chan Errable
	err     error
}

func newPump(decoder *CommandDecoder) *pump {
	return &pump{
		decoder: decoder,
		ch:      make(chan Errable),
		err:     nil,
	}
}

func (p *pump) run(ctx context.Context) {
	defer close(p.ch)

	for ctx.Err() == nil {
		cmd, err := p.decoder.Decode()
		if err != nil {
			p.err = err
			return
		}

		chanutil.Put(ctx, p.ch, cmd)
	}
	p.err = ctx.Err()
}

func (p *pump) Ch() <-chan Errable { return p.ch }
func (p *pump) Err() error         { return p.err }
