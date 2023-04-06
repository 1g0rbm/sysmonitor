package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/watcher"
)

func main() {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	agentConfig := config.GetConfigAgent()
	w := watcher.NewWatcher(agentConfig, l)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		log.Fatal(w.Run(ctx))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info().Msg("Stopping agent...")

	cancel()

	l.Info().Msg("Agent was stopped")
}
