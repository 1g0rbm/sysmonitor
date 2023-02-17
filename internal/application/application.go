package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/1g0rbm/sysmonitor/internal/storage"
)

const addr = ":8080"

type config struct {
	addr string
}

type App struct {
	storage storage.TStorage
	router  *chi.Mux
	config  config
}

func NewApp(s storage.TStorage, r *chi.Mux) *App {
	app := new(App)
	app.storage = s
	app.router = r
	app.config = config{
		addr: addr,
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", app.getAllMetricsHandler)
	r.Post("/update/{Type}/{Name}/{Value}", app.updateMetricHandler)
	r.Get("/value/{Type}/{Name}", app.getMetricHandler)

	return app
}

func (app App) Run() error {
	return http.ListenAndServe(app.config.addr, app.router)
}

func (app App) getAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	a := map[string]string{}
	sg, ok := app.storage.SType("gauge")
	if !ok {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	for key, v := range sg.All() {
		nv, _ := strconv.ParseFloat(string(v), 64)
		a[key] = fmt.Sprintf("%v", nv)
	}

	sc, ok := app.storage.SType("counter")
	if !ok {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	for key, v := range sc.All() {
		nv, _ := strconv.ParseInt(string(v), 10, 64)
		a[key] = fmt.Sprintf("%d", nv)
	}

	d, _ := json.Marshal(a)

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func (app App) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mType := chi.URLParam(r, "Type")
	mName := chi.URLParam(r, "Name")
	mValue := chi.URLParam(r, "Value")

	s, ok := app.storage.SType(mType)
	if !ok {
		http.Error(w, "invalid metric type", http.StatusNotImplemented)
		return
	}

	err := s.Set(mName, []byte(mValue))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) getMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mType := chi.URLParam(r, "Type")
	mName := chi.URLParam(r, "Name")

	s, sOk := app.storage.SType(mType)
	if !sOk {
		http.Error(w, "invalid metric type", http.StatusNotImplemented)
		return
	}

	val, vOk := s.Get(mName)
	if !vOk {
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	err := writeMetricValue(w, mType, val)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeMetricValue(w http.ResponseWriter, mType string, b []byte) error {
	switch mType {
	case "gauge":
		v, _ := strconv.ParseFloat(string(b), 64)
		_, err := w.Write([]byte(fmt.Sprintf("%v", v)))
		if err != nil {
			return err
		}
	case "counter":
		v, _ := strconv.ParseInt(string(b), 10, 64)
		_, err := w.Write([]byte(fmt.Sprintf("%d", v)))
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid type")
	}

	return nil
}
