package main

import (
	"github.com/1g0rbm/sysmonitor/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"

	"github.com/1g0rbm/sysmonitor/internal/handlers"
)

func main() {
	r := InitRouter()
	log.Fatal(http.ListenAndServe(":8080", r))
}

func InitRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := storage.NewMemStorage()

	handlers.RegisterUpdateHandler(r, s)
	handlers.RegisterGetOneHandler(r, s)
	handlers.RegisterGetAllHandler(r, s)

	return r
}
