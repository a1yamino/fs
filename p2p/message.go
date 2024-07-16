package p2p

const (
	IncomingMessage = iota
	IncomingStream
)

// RPC holds the data that is sent between peers in the network.
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
