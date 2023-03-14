package main

import (
	"context"
	"database/sql"
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

func main() {
	cfg := config.GetConfigServer()

	db, dbErr := sql.Open("pgx", cfg.DbDsn)
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	s := storage.NewStorage()
	app := application.NewApp(s, cfg, db)

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
