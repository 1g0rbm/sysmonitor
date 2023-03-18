package main

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1g0rbm/sysmonitor/internal/application"
	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.GetConfigServer()

	s, cls, dbErr := storage.NewStorage(cfg)
	if dbErr != nil {
		l.Fatal().Msg(dbErr.Error())
	}

	defer func() {
		if err := cls(); err != nil {
			l.Fatal().Msg(err.Error())
		}
	}()

	app := application.NewApp(s, cfg, l)

	go func() {
		if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Fatal().Msgf("Application start error: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info().Msg("Stopping application...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopErr := app.Stop(ctx)
	if stopErr != nil {
		l.Fatal().Msgf("Application stop error: %s", stopErr)
	}

	l.Info().Msg("Application stopped")
}
