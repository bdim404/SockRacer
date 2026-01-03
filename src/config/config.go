package config

import (
	"fmt"
	"net"
)

type Config struct {
	Listeners []ListenerConfig `json:"listeners"`
}

type UpstreamConfig struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
}

type ListenerConfig struct {
	Listen string           `json:"listen"`
	Socks  []UpstreamConfig `json:"socks"`
}

func (c *Config) Validate() error {
	if len(c.Listeners) == 0 {
		return fmt.Errorf("no listeners configured")
	}

	for i, listener := range c.Listeners {
		if err := listener.Validate(); err != nil {
			return fmt.Errorf("listener %d: %w", i, err)
		}
	}

	return nil
}

func (lc *ListenerConfig) Validate() error {
	if lc.Listen == "" {
		return fmt.Errorf("listen address is empty")
	}

	host, port, err := net.SplitHostPort(lc.Listen)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}

	if host == "" {
		return fmt.Errorf("listen host is empty")
	}

	if port == "" {
		return fmt.Errorf("listen port is empty")
	}

	if len(lc.Socks) == 0 {
		return fmt.Errorf("no socks upstreams configured")
	}

	for i, sock := range lc.Socks {
		if sock.Address == "" {
			return fmt.Errorf("socks upstream %d: address is empty", i)
		}
		_, _, err := net.SplitHostPort(sock.Address)
		if err != nil {
			return fmt.Errorf("invalid socks upstream %d (%s): %w", i, sock.Address, err)
		}
	}

	return nil
}
