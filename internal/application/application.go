package application

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/1g0rbm/sysmonitor/internal/fs"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

type App struct {
	storage storage.Storage
	router  *chi.Mux
	config  config.ServerConfig
	server  *http.Server
}

func NewApp(s storage.Storage, cfg config.ServerConfig) (app *App) {
	r := chi.NewRouter()

	app = &App{
		storage: s,
		config:  cfg,
		router:  r,
		server: &http.Server{
			Addr:    config.Address,
			Handler: r,
		},
	}

	app.router.Use(middleware.RequestID)
	app.router.Use(middleware.RealIP)
	app.router.Use(middleware.Logger)
	app.router.Use(middleware.Recoverer)

	app.router.Get("/", app.getAllMetricsHandler)
	app.router.Post("/update/{Type}/{Name}/{Value}", app.updateMetricHandler)
	app.router.Get("/value/{Type}/{Name}", app.getMetricHandler)

	app.router.Post("/update/", app.updateJSONMetricHandler)
	app.router.Post("/value/", app.getJSONMetricHandler)

	return app
}

func (app App) Run() (err error) {
	if app.config.StoreInterval > 0 && app.config.StoreFile != "" {
		go func() {
			dumpTicker := time.NewTicker(app.config.StoreInterval)
			defer dumpTicker.Stop()
			for {
				select {
				case <-dumpTicker.C:
					dErr := fs.DumpStorage(app.storage.All(), app.config.StoreFile)
					if dErr != nil && err == nil {
						err = dErr
					}
				}
			}
		}()
	}

	err = app.server.ListenAndServe()

	return
}

func (app App) Stop(ctx context.Context) error {
	dErr := fs.DumpStorage(app.storage.All(), app.config.StoreFile)
	if dErr != nil {
		return dErr
	}

	return app.server.Shutdown(ctx)
}

func (app App) getAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	t, tErr := template.New("metrics").Parse(AllMetricsTemplate)
	if tErr != nil {
		http.Error(w, tErr.Error(), http.StatusInternalServerError)
	}

	m := app.storage.All()

	if err := t.Execute(w, m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app App) updateJSONMetricHandler(w http.ResponseWriter, r *http.Request) {
	var m metric.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric structure"))
		return
	}

	if m.MType != metric.GaugeType && m.MType != metric.CounterType {
		sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric type"))
		return
	}

	if m.Delta == nil && m.Value == nil {
		sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric value"))
		return
	}

	updM, updErr := app.storage.Update(m)
	if updErr != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("update error"))
		return
	}

	b, err := json.Marshal(updM)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("error while response creation"))
		return
	}

	sendJSONResponse(w, http.StatusOK, b)
}

func (app App) getJSONMetricHandler(w http.ResponseWriter, r *http.Request) {
	var rm metric.Metrics

	if err := json.NewDecoder(r.Body).Decode(&rm); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric structure"))
		return
	}

	if rm.MType != metric.GaugeType && rm.MType != metric.CounterType {
		sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric type"))
		return
	}

	m, err := app.storage.Get(rm.Name())
	if err != nil && errors.Is(storage.ErrMetricNotFound, err) {
		sendJSONResponse(w, http.StatusNotFound, []byte(err.Error()))
		return
	}
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("internal server error"))
		log.Fatalf("Error during getting metric from storage: %s", err)
		return
	}

	b, mErr := json.Marshal(m)
	if mErr != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("internal server error"))
		log.Fatalf("Metric marshaling error: %s", mErr)
		return
	}

	sendJSONResponse(w, http.StatusOK, b)
}

func (app App) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mName := chi.URLParam(r, "Name")
	mType := chi.URLParam(r, "Type")
	mValue := chi.URLParam(r, "Value")

	if mName == "" || mType == "" {
		http.Error(w, "invalid path params", http.StatusBadRequest)
		return
	}

	m, mErr := metric.NewMetric(mName, mType, mValue)
	if errors.Is(metric.ErrInvalidValue, mErr) {
		http.Error(w, mErr.Error(), http.StatusBadRequest)
		return
	}
	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusNotImplemented)
		return
	}

	_, updErr := app.storage.Update(m)
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

	m, vErr := app.storage.Get(mName)
	if vErr != nil {
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(m.ValueAsString()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app App) getRouter() chi.Router {
	return app.router
}

func sendJSONResponse(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Fatalf("Error while sending response: %s", err)
	}
}
