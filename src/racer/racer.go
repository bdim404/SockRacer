package racer

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/bdim404/SockRacer/src/config"
	"github.com/bdim404/SockRacer/src/socks5"
)

type Racer struct {
	upstreams []config.UpstreamConfig
	timeout   time.Duration
}

func New(upstreams []config.UpstreamConfig, timeout time.Duration) *Racer {
	return &Racer{
		upstreams: upstreams,
		timeout:   timeout,
	}
}

type raceResult struct {
	conn      net.Conn
	proxyAddr string
	proxyName string
	err       error
	duration  time.Duration
}

func (r *Racer) Race(ctx context.Context, target *socks5.TargetAddress) (net.Conn, error) {
	raceCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	resultCh := make(chan *raceResult, len(r.upstreams))
	startTime := time.Now()

	log.Printf("racing %d upstreams for %s", len(r.upstreams), target)

	for _, upstream := range r.upstreams {
		go func(u config.UpstreamConfig) {
			connStart := time.Now()
			conn, err := socks5.DialSOCKS5(raceCtx, u.Address, target)
			duration := time.Since(connStart)

			select {
			case resultCh <- &raceResult{
				conn:      conn,
				proxyAddr: u.Address,
				proxyName: u.Name,
				err:       err,
				duration:  duration,
			}:
			case <-raceCtx.Done():
				if conn != nil {
					conn.Close()
				}
			}
		}(upstream)
	}

	var errors []string

	for i := 0; i < len(r.upstreams); i++ {
		select {
		case result := <-resultCh:
			if result.err == nil {
				displayName := result.proxyAddr
				if result.proxyName != "" {
					displayName = fmt.Sprintf("%s (%s)", result.proxyName, result.proxyAddr)
				}

				log.Printf("✓ winner: %s (%dms)",
					displayName, result.duration.Milliseconds())

				go func() {
					for j := i + 1; j < len(r.upstreams); j++ {
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

				return result.conn, nil
			}

			displayName := result.proxyAddr
			if result.proxyName != "" {
				displayName = fmt.Sprintf("%s (%s)", result.proxyName, result.proxyAddr)
			}
			log.Printf("✗ failed: %s (%dms) - %v",
				displayName, result.duration.Milliseconds(), result.err)
			errors = append(errors, fmt.Sprintf("%s: %v", displayName, result.err))

		case <-raceCtx.Done():
			return nil, fmt.Errorf("race timeout after %dms", time.Since(startTime).Milliseconds())
		}
	}

	return nil, fmt.Errorf("all upstreams failed: %s", strings.Join(errors, "; "))
}
