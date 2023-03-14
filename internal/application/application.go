package application

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/fs"
	"github.com/1g0rbm/sysmonitor/internal/metric"
	localmiddleware "github.com/1g0rbm/sysmonitor/internal/middleware"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

type App struct {
	storage storage.Storage
	router  *chi.Mux
	config  *config.ServerConfig
	server  *http.Server
	db      *sql.DB
}

func NewApp(s storage.Storage, cfg *config.ServerConfig, db *sql.DB) (app *App) {
	r := chi.NewRouter()

	app = &App{
		storage: s,
		config:  cfg,
		router:  r,
		server: &http.Server{
			Addr:    cfg.Address,
			Handler: r,
		},
		db: db,
	}

	app.router.Use(middleware.RequestID)
	app.router.Use(middleware.RealIP)
	app.router.Use(middleware.Logger)
	app.router.Use(middleware.Recoverer)
	app.router.Use(localmiddleware.Gzip)

	app.router.Get("/", app.getAllMetricsHandler)
	app.router.Post("/update/{Type}/{Name}/{Value}", app.updateMetricHandler)
	app.router.Get("/value/{Type}/{Name}", app.getMetricHandler)

	app.router.Post("/update/", app.updateJSONMetricHandler)
	app.router.Post("/value/", app.getJSONMetricHandler)

	app.router.Get("/ping", app.dbHealthCheckHandler)

	return app
}

func (app App) Run() (err error) {
	if app.config.Restore {
		err = fs.RestoreStorage(app.storage, app.config.StoreFile)
		if err != nil {
			return err
		}
		log.Println("Metrics restored from file")
	}

	if app.config.NeedPeriodicalStore() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func(ctx context.Context) {
			dumpTicker := time.NewTicker(app.config.StoreInterval)
			defer dumpTicker.Stop()

			log.Printf("Metrics will be dumped every %d seconds", app.config.StoreInterval)

			for {
				select {
				case <-dumpTicker.C:
					dErr := fs.DumpStorage(app.storage.All(), app.config.StoreFile)
					if dErr != nil && err == nil {
						err = dErr
					}
				case <-ctx.Done():
					return
				}
			}
		}(ctx)
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

func (app App) dbHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := app.db.PingContext(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (app App) getAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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

	if decodeErr := m.Decode(r.Body); decodeErr != nil {
		sendJSONResponse(w, http.StatusBadRequest, []byte(decodeErr.Error()))
		return
	}

	if m.Delta == nil && m.Value == nil {
		sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric value"))
		return
	}

	if app.config.NeedCheckSign() {
		ok, signErr := m.CheckSign(app.config.Key)
		if signErr != nil {
			sendJSONResponse(w, http.StatusInternalServerError, []byte("check sign error"))
			return
		}
		if !ok {
			sendJSONResponse(w, http.StatusBadRequest, []byte("wrong sign"))
			return
		}
	}

	im, convertErr := m.ToIMetric()
	if convertErr != nil {
		sendJSONResponse(w, http.StatusBadRequest, []byte(convertErr.Error()))
		return
	}

	updM, updErr := app.storage.Update(im)
	if updErr != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("update error"))
		return
	}

	rm, rmErr := metric.NewMetricsFromIMetric(updM)
	if rmErr != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("update error"))
		return
	}

	if app.config.NeedCheckSign() {
		if rmSignErr := rm.Sign(app.config.Key); rmSignErr != nil {
			sendJSONResponse(w, http.StatusInternalServerError, []byte("check sign error"))
			return
		}
	}

	b, err := rm.Encode()
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("error while response creation"))
		return
	}

	sendJSONResponse(w, http.StatusOK, b)
}

func (app App) getJSONMetricHandler(w http.ResponseWriter, r *http.Request) {
	var rm metric.Metrics

	if decodeErr := rm.Decode(r.Body); decodeErr != nil {
		sendJSONResponse(w, http.StatusBadRequest, []byte(decodeErr.Error()))
		return
	}

	m, err := app.storage.Get(rm.ID)
	if err != nil && errors.Is(storage.ErrMetricNotFound, err) {
		sendJSONResponse(w, http.StatusNotFound, []byte(err.Error()))
		return
	}

	resM, rmErr := metric.NewMetricsFromIMetric(m)
	if rmErr != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("update error"))
		return
	}

	if app.config.NeedCheckSign() {
		_ = resM.Sign(app.config.Key)
	}

	if app.config.NeedCheckSign() {
		if resSignErr := resM.Sign(app.config.Key); resSignErr != nil {
			sendJSONResponse(w, http.StatusInternalServerError, []byte("check sign error"))
			return
		}
	}

	b, mErr := resM.Encode()
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
