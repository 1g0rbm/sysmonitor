package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/1g0rbm/sysmonitor/internal/application"
	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

var (
	s     storage.Storage
	dbErr error
)

func main() {
	cfg := config.GetConfigServer()

	switch cfg.GetStorageDriverName() {
	case storage.DBStorageType:
		s, dbErr = storage.NewDBStorage("pgx", cfg.DBDsn)
		if dbErr != nil {
			log.Fatalf("Create DB storage error: %s", dbErr)
		}
	case storage.MemStorageType:
		s = storage.NewMemStorage()
	default:
		log.Fatalf("Invalid storage type %d", cfg.GetStorageDriverName())
	}

	app := application.NewApp(s, cfg)

	go func() {
		if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Application start error: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Stopping application...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopErr := app.Stop(ctx)
	if stopErr != nil {
		log.Fatalf("Application stop error: %s", stopErr)
	}

	log.Println("Application stopped")
}
