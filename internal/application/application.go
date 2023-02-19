package application

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"net/http"
	"path"
	"runtime"
	"strconv"

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
	_, filename, _, ok := runtime.Caller(0)

	if !ok {
		http.Error(w, "can not find template", http.StatusInternalServerError)
	}

	tmpl := template.Must(template.ParseFiles(path.Dir(filename) + "/../template/metrics.html"))

	type data struct {
		Metrics map[string]metric.IMetric
	}

	err := tmpl.Execute(w, data{Metrics: app.storage.All()})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app App) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mName := chi.URLParam(r, "Name")
	mType := chi.URLParam(r, "Type")
	mValue := chi.URLParam(r, "Value")

	if mValue == "" || mName == "" || mType == "" {
		http.Error(w, "invalid path params", http.StatusBadRequest)
		return
	}

	_, vErr := strconv.ParseFloat(mValue, 64)
	if vErr != nil {
		http.Error(w, "value should be a numeric", http.StatusBadRequest)
		return
	}

	m, mErr := metric.NewMetric(mName, mType, []byte(mValue))
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
	if mName == "" {
		http.Error(w, "invalid path params", http.StatusBadRequest)
		return
	}

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
