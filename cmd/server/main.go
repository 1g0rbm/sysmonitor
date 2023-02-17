package main

import (
	"log"

	"github.com/go-chi/chi/v5"

	"github.com/1g0rbm/sysmonitor/internal/application"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

func main() {
	app := application.NewApp(storage.NewStorage(), chi.NewRouter())

	log.Fatal(app.Run())
}
