package listener

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bdim404/SockRacer/src/config"
	"github.com/bdim404/SockRacer/src/pool"
)

type Listener struct {
	cfg  *config.ListenerConfig
	ln   net.Listener
	pool *pool.Pool
	wg   sync.WaitGroup
}

func New(cfg *config.ListenerConfig) (*Listener, error) {
	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return nil, err
	}

	p := pool.New(cfg.Socks, 5*time.Second)

	return &Listener{
		cfg:  cfg,
		ln:   ln,
		pool: p,
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

		log.Printf("accepted connection from %s", conn.RemoteAddr())

		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			handleConnection(ctx, conn, l.pool)
		}()
	}
}
