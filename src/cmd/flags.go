package cmd

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/bdim404/SockRacer/src/config"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type Flags struct {
	ConfigPath string
	Config     *config.Config
}

func printHelp() {
	fmt.Fprintf(os.Stderr, "SockRacer - SOCKS5 parallel racing aggregator\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  sockracer [options]\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  Config file mode:\n")
	fmt.Fprintf(os.Stderr, "    sockracer --config /path/to/config.json\n")
	fmt.Fprintf(os.Stderr, "    sockracer (uses ./config.json by default)\n\n")
	fmt.Fprintf(os.Stderr, "  Command line mode:\n")
	fmt.Fprintf(os.Stderr, "    sockracer --listen-address 127.0.0.1 --listen-port 1080 --socks upstream1:1081 --socks upstream2:1082\n")
}

func ParseFlags() (*Flags, error) {
	var configPath string
	var listenAddr string
	var listenPort string
	var socks stringSlice
	var help bool

	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.StringVar(&configPath, "c", "config.json", "Path to config file (shorthand)")
	flag.StringVar(&listenAddr, "listen-address", "127.0.0.1", "Listen address")
	flag.StringVar(&listenPort, "listen-port", "", "Listen port")
	flag.Var(&socks, "socks", "Upstream SOCKS5 proxy (can be specified multiple times)")
	flag.BoolVar(&help, "help", false, "Show help message")
	flag.BoolVar(&help, "h", false, "Show help message (shorthand)")
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	if listenPort != "" {
		if len(socks) == 0 {
			return nil, fmt.Errorf("at least one --socks upstream must be specified")
		}

		listen := net.JoinHostPort(listenAddr, listenPort)

		upstreams := make([]config.UpstreamConfig, len(socks))
		for i, addr := range socks {
			upstreams[i] = config.UpstreamConfig{
				Address: addr,
			}
		}

		cfg := &config.Config{
			Listeners: []config.ListenerConfig{
				{
					Listen: listen,
					Socks:  upstreams,
				},
			},
		}

		if err := cfg.Validate(); err != nil {
			return nil, err
		}

		return &Flags{Config: cfg}, nil
	}

	return &Flags{ConfigPath: configPath}, nil
}
