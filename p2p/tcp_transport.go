package p2p

import (
	"errors"
	"log"
	"net"
	"sync"
)

const (
	maxMessageSize = 1024
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// The underlying TCP connection.
	net.Conn
	// true if the connection is outgoing, false if it is incoming.
	outbound bool

	wg *sync.WaitGroup
}

var _ = Peer(&TCPPeer{})

func NewTCPPeer(conn net.Conn, out bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: out,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(data []byte) error {
	_, err := p.Write(data)
	return err
}

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

type TCPTransportOpts struct {
	ListenAddr string
	Handshake  HandshakeFunc
	Decoder    Decoder
	OnPeer     func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listner net.Listener
	rpcch   chan RPC
}

var _ = Transport(&TCPTransport{})

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, maxMessageSize),
	}
}

func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) Close() error {
	return t.listner.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) ListenAndServe() error {
	var err error

	t.listner, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAccepting()

	log.Printf("TCP tracsport listening on port: %s\n", t.ListenAddr)

	return nil
}

func (t *TCPTransport) startAccepting() {
	for {
		conn, err := t.listner.Accept()

		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}

		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		log.Printf("dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.Handshake(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		rpc.From = conn.LocalAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			log.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			log.Printf("[%s] stream closed\n", conn.RemoteAddr())
			continue
		}

		t.rpcch <- rpc
	}
}
