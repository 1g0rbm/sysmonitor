package main

import (
	"github.com/1g0rbm/sysmonitor/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"

	"github.com/1g0rbm/sysmonitor/internal/handlers"
)

const addr string = ":8080"

func main() {
	r := InitRouter()
	log.Fatal(http.ListenAndServe(addr, r))
}

func InitRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := storage.NewStorage()

	handlers.RegisterUpdateHandler(r, s)
	handlers.RegisterGetOneHandler(r, s)
	handlers.RegisterGetAllHandler(r, s)

	return r
}
