package listener

import (
	"context"
	"net"

	"github.com/bdim404/SockRacer/src/pool"
	"github.com/bdim404/SockRacer/src/relay"
	"github.com/bdim404/SockRacer/src/socks5"
)

func handleConnection(ctx context.Context, clientConn net.Conn, p *pool.Pool) {
	defer clientConn.Close()

	if err := socks5.HandleNegotiation(clientConn); err != nil {
		return
	}

	target, err := socks5.ParseRequest(clientConn)
	if err != nil {
		socks5.SendReply(clientConn, socks5.RepGeneralFailure, nil)
		return
	}

	upstreamConn, err := p.GetConn(ctx, target)
	if err != nil {
		socks5.SendReply(clientConn, socks5.RepHostUnreachable, nil)
		return
	}
	defer upstreamConn.Close()

	if err := socks5.SendReply(clientConn, socks5.RepSuccess, upstreamConn.LocalAddr()); err != nil {
		return
	}

	relay.Bidirectional(ctx, clientConn, upstreamConn)
}
