package listener

import (
	"context"
	"log"
	"net"

	"github.com/bdim404/SockRacer/src/pool"
	"github.com/bdim404/SockRacer/src/relay"
	"github.com/bdim404/SockRacer/src/socks5"
)

func handleConnection(ctx context.Context, clientConn net.Conn, p *pool.Pool) {
	defer clientConn.Close()

	if err := socks5.HandleNegotiation(clientConn); err != nil {
		log.Printf("negotiation failed: %v", err)
		return
	}

	target, err := socks5.ParseRequest(clientConn)
	if err != nil {
		log.Printf("parse request failed: %v", err)
		socks5.SendReply(clientConn, socks5.RepGeneralFailure, nil)
		return
	}

	upstreamConn, err := p.GetConn(ctx, target)
	if err != nil {
		log.Printf("failed to get upstream conn for %s: %v", target, err)
		socks5.SendReply(clientConn, socks5.RepHostUnreachable, nil)
		return
	}

	if err := socks5.SendReply(clientConn, socks5.RepSuccess, upstreamConn.LocalAddr()); err != nil {
		log.Printf("send reply failed: %v", err)
		upstreamConn.Close()
		return
	}

	relay.Bidirectional(ctx, clientConn, upstreamConn)
	upstreamConn.Close()
}
