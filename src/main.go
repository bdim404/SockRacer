package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bdim404/SockRacer/src/cmd"
	"github.com/bdim404/SockRacer/src/config"
	"github.com/bdim404/SockRacer/src/listener"
)

func main() {
	flags, err := cmd.ParseFlags()
	if err != nil {
		log.Fatalf("parse flags: %v", err)
	}

	var cfg *config.Config
	if flags.Config != nil {
		cfg = flags.Config
	} else {
		cfg, err = config.LoadConfig(flags.ConfigPath)
		if err != nil {
			log.Fatalf("load config: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	for _, listenerCfg := range cfg.Listeners {
		l, err := listener.New(&listenerCfg)
		if err != nil {
			log.Fatalf("create listener for %s: %v", listenerCfg.Listen, err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Serve(ctx); err != nil {
				log.Printf("listener error: %v", err)
			}
		}()
	}

	<-sigCh
	log.Println("shutting down...")
	cancel()
	wg.Wait()
	log.Println("shutdown complete")
}
