package listener

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/bdim404/SockRacer/src/config"
	"github.com/bdim404/SockRacer/src/racer"
	"github.com/bdim404/SockRacer/src/relay"
	"github.com/bdim404/SockRacer/src/socks5"
)

func handleConnection(ctx context.Context, clientConn net.Conn, upstreams []config.UpstreamConfig) {
	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("→ new connection from %s", clientAddr)
	defer func() {
		clientConn.Close()
		log.Printf("← connection closed from %s", clientAddr)
	}()

	if err := socks5.HandleNegotiation(clientConn); err != nil {
		log.Printf("negotiation failed from %s: %v", clientAddr, err)
		return
	}

	target, err := socks5.ParseRequest(clientConn)
	if err != nil {
		log.Printf("parse request failed from %s: %v", clientAddr, err)
		socks5.SendReply(clientConn, socks5.RepGeneralFailure, nil)
		return
	}

	log.Printf("request from %s to %s", clientAddr, target)

	r := racer.New(upstreams, 10*time.Second)
	upstreamConn, err := r.Race(ctx, target)
	if err != nil {
		log.Printf("race failed for %s from %s: %v", target, clientAddr, err)
		socks5.SendReply(clientConn, socks5.RepHostUnreachable, nil)
		return
	}
	defer upstreamConn.Close()

	if err := socks5.SendReply(clientConn, socks5.RepSuccess, upstreamConn.LocalAddr()); err != nil {
		log.Printf("send reply failed to %s: %v", clientAddr, err)
		return
	}

	log.Printf("relaying data for %s -> %s", clientAddr, target)
	relay.Bidirectional(ctx, clientConn, upstreamConn)
}
