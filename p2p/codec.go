package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GobDecoder struct{}

func (dec GobDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	stream := buf[0] == IncomingStream
	if stream {
		msg.Stream = true
		return nil
	}

	buf = make([]byte, 4+maxMessageSize)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
