package listener

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/bdim404/SockRacer/src/config"
)

type Listener struct {
	cfg *config.ListenerConfig
	ln  net.Listener
	wg  sync.WaitGroup
}

func New(cfg *config.ListenerConfig) (*Listener, error) {
	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return nil, err
	}

	return &Listener{
		cfg: cfg,
		ln:  ln,
	}, nil
}

func (l *Listener) Serve(ctx context.Context) error {
	defer l.ln.Close()

	go func() {
		<-ctx.Done()
		l.ln.Close()
	}()

	log.Printf("listening on %s with %d upstreams", l.cfg.Listen, len(l.cfg.Socks))

	for {
		conn, err := l.ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				l.wg.Wait()
				return nil
			default:
				log.Printf("accept error: %v", err)
				continue
			}
		}

		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			handleConnection(ctx, conn, l.cfg.Socks)
		}()
	}
}
