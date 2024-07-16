package p2p

import "net"

// Peer is an interface that represents the remote node in the network.
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport is anything that handles the communication between peers in the network.
// It could be a TCP connection, a WebRTC connection, or anything else.
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndServe() error
	Consume() <-chan RPC
	Close() error
}
