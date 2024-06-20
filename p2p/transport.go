package p2p

// Peer is an interface that represents the remote node in the network.
type Peer interface {
}

// Transport is anything that handles the communication between peers in the network.
// It could be a TCP connection, a WebRTC connection, or anything else.
type Transport interface {
}
