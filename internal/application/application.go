package application

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/1g0rbm/sysmonitor/internal/metric"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

const addr = ":8080"

type config struct {
	addr string
}

type App struct {
	storage storage.Storage
	router  *chi.Mux
	config  config
}

func NewApp(s storage.Storage, r *chi.Mux) *App {
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
	for _, m := range app.storage.All() {
		a[m.Name()] = m.ValueAsString()
	}

	d, _ := json.Marshal(a)

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func (app App) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mValue := chi.URLParam(r, "Value")
	_, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}

	m, mErr := metric.NewMetric(
		chi.URLParam(r, "Name"),
		chi.URLParam(r, "Type"),
		[]byte(mValue),
	)

	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusNotImplemented)
		return
	}

	app.storage.Set(m)

	w.WriteHeader(http.StatusOK)
}

func (app App) getMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mName := chi.URLParam(r, "Name")

	m, vOk := app.storage.Get(mName)
	if !vOk {
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(m.ValueAsString()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
