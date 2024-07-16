package p2p

type HandshakeFunc func(Peer) error

func NOPHandshake(p Peer) error {
	return nil
}
