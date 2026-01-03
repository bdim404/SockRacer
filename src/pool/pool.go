package pool

import (
	"context"
	"fmt"
	"log"
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
	duration  time.Duration
}

func (p *Pool) GetConn(ctx context.Context, target *socks5.TargetAddress) (net.Conn, error) {
	return p.race(ctx, target)
}

func (p *Pool) race(ctx context.Context, target *socks5.TargetAddress) (net.Conn, error) {
	raceCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	resultCh := make(chan *result, len(p.upstreams))
	raceStartTime := time.Now()

	for _, upstream := range p.upstreams {
		go func(u config.UpstreamConfig) {
			start := time.Now()
			conn, err := socks5.DialSOCKS5(raceCtx, u.Address, target)
			duration := time.Since(start)

			select {
			case resultCh <- &result{
				conn:     conn,
				upstream: u,
				err:      err,
				duration: duration,
			}:
			case <-raceCtx.Done():
				if conn != nil {
					conn.Close()
				}
			}
		}(upstream)
	}

	var winnerConn net.Conn
	var winnerUpstream config.UpstreamConfig
	var winnerDuration time.Duration

	for i := 0; i < len(p.upstreams); i++ {
		select {
		case res := <-resultCh:
			if res.err == nil && res.conn != nil && winnerConn == nil {
				winnerConn = res.conn
				winnerUpstream = res.upstream
				winnerDuration = res.duration

				p.mu.Lock()
				p.lastWinner = &winnerUpstream
				p.winnerExpiry = time.Now().Add(30 * time.Second)
				p.mu.Unlock()

				winnerName := winnerUpstream.Address
				if winnerUpstream.Name != "" {
					winnerName = fmt.Sprintf("%s (%s)", winnerUpstream.Name, winnerUpstream.Address)
				}
				log.Printf("✓ %s -> %s (%dms)", target, winnerName, winnerDuration.Milliseconds())

				go p.collectRaceStats(resultCh, len(p.upstreams)-i-1, raceStartTime, &winnerUpstream, target)

				return winnerConn, nil
			}

			if winnerConn != nil && res.conn != nil {
				res.conn.Close()
			}

		case <-raceCtx.Done():
			log.Printf("✗ %s race timeout after %dms", target, time.Since(raceStartTime).Milliseconds())
			return nil, fmt.Errorf("race timeout")
		}
	}

	log.Printf("✗ %s all upstreams failed", target)
	return nil, fmt.Errorf("all upstreams failed")
}

func (p *Pool) collectRaceStats(resultCh chan *result, remaining int, raceStart time.Time, winner *config.UpstreamConfig, target *socks5.TargetAddress) {
	for i := 0; i < remaining; i++ {
		res := <-resultCh
		if res.conn != nil {
			res.conn.Close()
		}
	}
}
