package application

import (
	"github.com/1g0rbm/sysmonitor/internal/tmp"
	"html/template"
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

func NewApp(s storage.Storage) *App {
	app := new(App)
	app.storage = s
	app.router = chi.NewRouter()
	app.config = config{
		addr: addr,
	}

	app.router.Use(middleware.RequestID)
	app.router.Use(middleware.RealIP)
	app.router.Use(middleware.Logger)
	app.router.Use(middleware.Recoverer)

	app.router.Get("/", app.getAllMetricsHandler)
	app.router.Post("/update/{Type}/{Name}/{Value}", app.updateMetricHandler)
	app.router.Get("/value/{Type}/{Name}", app.getMetricHandler)

	return app
}

func (app App) Run() error {
	return http.ListenAndServe(app.config.addr, app.router)
}

func (app App) getAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	t, tErr := template.New("metrics").Parse(tmp.AllMetricsTmp)
	if tErr != nil {
		http.Error(w, tErr.Error(), http.StatusInternalServerError)
	}

	type data struct {
		Metrics map[string]metric.IMetric
	}

	if err := t.Execute(w, data{Metrics: app.storage.All()}); err != nil {
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
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}

	m, mErr := metric.NewMetric(mName, mType, []byte(mValue))
	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusNotImplemented)
		return
	}

	updErr := updateMetric(app.storage, m)
	if updErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

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

func updateMetric(s storage.Storage, m metric.IMetric) error {
	switch m.Type() {
	case metric.CounterType:
		curV, curVErr := metric.NormalizeCounterMetricValue(m)
		if curVErr != nil {
			return curVErr
		}

		cm, ok := s.Get(m.Name())
		if !ok {
			s.Set(m)
			return nil
		}

		newV, newVErr := metric.NormalizeCounterMetricValue(cm)
		if newVErr != nil {
			return newVErr
		}

		updM, updErr := metric.NewMetric(m.Name(), m.Type(), []byte(strconv.FormatInt(int64(curV+newV), 10)))
		if updErr != nil {
			return updErr
		}

		s.Set(updM)
		return nil
	default:
		s.Set(m)
		return nil
	}
}
