package cmd

import (
	"fmt"
	"net"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/bdim404/parallel-socks/src/config"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *stringSlice) Type() string {
	return "string"
}

type Flags struct {
	ConfigPath string
	LogLevel   string
	Config     *config.Config
}

func printHelp() {
	fmt.Fprintf(os.Stderr, "parallel-socks - SOCKS5 parallel racing aggregator\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  parallel-socks [options]\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  Config file mode:\n")
	fmt.Fprintf(os.Stderr, "    parallel-socks --config /path/to/config.json\n")
	fmt.Fprintf(os.Stderr, "    parallel-socks -c /path/to/config.json\n")
	fmt.Fprintf(os.Stderr, "    parallel-socks (uses ./config.json by default)\n\n")
	fmt.Fprintf(os.Stderr, "  Command line mode:\n")
	fmt.Fprintf(os.Stderr, "    parallel-socks --listen-address ::1 --listen-port 1080 --socks upstream1:1081 --socks upstream2:1082\n")
	fmt.Fprintf(os.Stderr, "    parallel-socks -a ::1 -p 1080 -s upstream1:1081 -s upstream2:1082\n")
	fmt.Fprintf(os.Stderr, "    parallel-socks -a ::1 -p 1080 -s upstream1:1081 -l debug\n")
}

func ParseFlags() (*Flags, error) {
	var configPath string
	var listenAddr string
	var listenPort string
	var socks stringSlice
	var logLevel string
	var help bool

	flag.StringVarP(&configPath, "config", "c", "config.json", "Path to config file")
	flag.StringVarP(&listenAddr, "listen-address", "a", "::1", "Listen address")
	flag.StringVarP(&listenPort, "listen-port", "p", "", "Listen port")
	flag.VarP(&socks, "socks", "s", "Upstream SOCKS5 proxy (can be specified multiple times)")
	flag.StringVarP(&logLevel, "log-level", "l", "", "Log level: debug or info")
	flag.BoolVarP(&help, "help", "h", false, "Show help message")
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

		if logLevel != "" {
			cfg.LogLevel = logLevel
		}

		if err := cfg.Validate(); err != nil {
			return nil, err
		}

		return &Flags{Config: cfg}, nil
	}

	flags := &Flags{ConfigPath: configPath}
	if logLevel != "" {
		flags.LogLevel = logLevel
	}
	return flags, nil
}
