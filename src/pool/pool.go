package pool

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bdim404/SockRacer/src/config"
	"github.com/bdim404/SockRacer/src/socks5"
)

type Pool struct {
	upstreams    []config.UpstreamConfig
	timeout      time.Duration
	lastWinner   *config.UpstreamConfig
	winnerExpiry time.Time
	mu           sync.RWMutex
}

func New(upstreams []config.UpstreamConfig, timeout time.Duration) *Pool {
	return &Pool{
		upstreams: upstreams,
		timeout:   timeout,
	}
}

type result struct {
	conn      net.Conn
	upstream  config.UpstreamConfig
	err       error
	isFastest bool
}

func (p *Pool) GetConn(ctx context.Context, target *socks5.TargetAddress) (net.Conn, error) {
	p.mu.RLock()
	winner := p.lastWinner
	expired := time.Now().After(p.winnerExpiry)
	p.mu.RUnlock()

	if winner != nil && !expired {
		conn, err := socks5.DialSOCKS5(ctx, winner.Address, target)
		if err == nil {
			return conn, nil
		}
	}

	return p.race(ctx, target)
}

func (p *Pool) race(ctx context.Context, target *socks5.TargetAddress) (net.Conn, error) {
	raceCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	resultCh := make(chan *result, len(p.upstreams))

	for _, upstream := range p.upstreams {
		go func(u config.UpstreamConfig) {
			conn, err := socks5.DialSOCKS5(raceCtx, u.Address, target)

			select {
			case resultCh <- &result{
				conn:     conn,
				upstream: u,
				err:      err,
			}:
			case <-raceCtx.Done():
				if conn != nil {
					conn.Close()
				}
			}
		}(upstream)
	}

	var errors []string

	for i := 0; i < len(p.upstreams); i++ {
		select {
		case res := <-resultCh:
			if res.err == nil && res.conn != nil {
				p.mu.Lock()
				p.lastWinner = &res.upstream
				p.winnerExpiry = time.Now().Add(30 * time.Second)
				p.mu.Unlock()

				go func() {
					for j := i + 1; j < len(p.upstreams); j++ {
						select {
						case r := <-resultCh:
							if r.conn != nil {
								r.conn.Close()
							}
						case <-raceCtx.Done():
							return
						}
					}
				}()

				return res.conn, nil
			}

			if res.err != nil {
				displayName := res.upstream.Address
				if res.upstream.Name != "" {
					displayName = res.upstream.Name
				}
				errors = append(errors, displayName)
			}

		case <-raceCtx.Done():
			return nil, fmt.Errorf("race timeout")
		}
	}

	return nil, fmt.Errorf("all upstreams failed")
}
