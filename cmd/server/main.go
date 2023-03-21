package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"

	"github.com/1g0rbm/sysmonitor/internal/application"
	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

func main() {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.GetConfigServer()

	s, dbErr := storage.NewStorage(cfg.DBDsn)
	if dbErr != nil {
		l.Fatal().Msg(dbErr.Error())
	}

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

	stopErr := app.Shutdown(ctx)
	if stopErr != nil {
		l.Fatal().Msgf("Application stop error: %s", stopErr)
	}

	l.Info().Msg("Application stopped")
}
