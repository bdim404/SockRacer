package relay

import (
	"context"
	"io"
	"net"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

func Bidirectional(ctx context.Context, a, b net.Conn) {
	errCh := make(chan error, 2)

	go func() {
		buf := bufferPool.Get().([]byte)
		_, err := io.CopyBuffer(a, b, buf)
		bufferPool.Put(buf)
		if tcpConn, ok := a.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		errCh <- err
	}()

	go func() {
		buf := bufferPool.Get().([]byte)
		_, err := io.CopyBuffer(b, a, buf)
		bufferPool.Put(buf)
		if tcpConn, ok := b.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		errCh <- err
	}()

	select {
	case <-errCh:
		<-errCh
	case <-ctx.Done():
		<-errCh
		<-errCh
	}
}
