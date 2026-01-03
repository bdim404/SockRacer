package listener

import (
	"context"
	"io"
	"log"
	"net"

	"github.com/bdim404/SockRacer/src/pool"
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

	if err := socks5.SendReply(clientConn, socks5.RepSuccess, clientConn.LocalAddr()); err != nil {
		log.Printf("send reply failed: %v", err)
		return
	}

	upstreamConn, _, err := p.GetConn(ctx, target)
	if err != nil {
		return
	}
	defer upstreamConn.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(upstreamConn, clientConn)
		upstreamConn.Close()
		done <- struct{}{}
	}()

	go func() {
		io.Copy(clientConn, upstreamConn)
		clientConn.Close()
		done <- struct{}{}
	}()

	select {
	case <-done:
		<-done
	case <-ctx.Done():
		clientConn.Close()
		upstreamConn.Close()
		<-done
		<-done
	}
}
