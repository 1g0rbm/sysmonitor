package application

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
	localmiddleware "github.com/1g0rbm/sysmonitor/internal/middleware"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

const (
	metricOnPage = 50
	page         = 1
)

type App struct {
	storage storage.Storage
	router  *chi.Mux
	config  *config.ServerConfig
	server  *http.Server
	logger  zerolog.Logger
}

func NewApp(s storage.Storage, cfg *config.ServerConfig, l zerolog.Logger) (app *App) {
	r := chi.NewRouter()

	app = &App{
		storage: s,
		config:  cfg,
		router:  r,
		server: &http.Server{
			Addr:    cfg.Address,
			Handler: r,
		},
		logger: l,
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

	app.router.Post("/updates/", app.updateJSONMetricsHandler)

	app.router.Get("/ping", app.dbHealthCheckHandler)

	return app
}

func (app App) Run() (err error) {
	if app.config.NeedRestore() {
		mem, itIsMem := app.storage.(storage.MemStorage)
		if !itIsMem {
			return fmt.Errorf("try to restor non memstorage storage")
		}
		if err := mem.Restore(app.config.StoreFile); err != nil {
			return err
		}
		app.logger.Info().Msg("Metrics restored from file")
	}

	if app.config.NeedPeriodicalStore() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func(ctx context.Context) {
			dumpTicker := time.NewTicker(app.config.StoreInterval)
			defer dumpTicker.Stop()

			app.logger.Info().Msgf("Metrics will be dumped every %d seconds", app.config.StoreInterval/1000)

			for {
				select {
				case <-dumpTicker.C:
					mem, itIsMem := app.storage.(storage.MemStorage)
					if !itIsMem {
						err = fmt.Errorf("try store datra for non memstorage")
					}
					err = mem.BackupData(app.config.StoreFile)
				case <-ctx.Done():
					return
				}
			}
		}(ctx)
	}

	app.logger.Info().Msgf("Application started on host %s\n", app.config.Address)
	err = app.server.ListenAndServe()

	return
}

func (app App) Shutdown(ctx context.Context) error {
	db, itIsDB := app.storage.(storage.DBStorage)
	if itIsDB {
		if err := db.Close(); err != nil {
			return err
		}
	}

	mem, itIsMem := app.storage.(storage.MemStorage)
	if itIsMem {
		if backUpErr := mem.BackupData(app.config.StoreFile); backUpErr != nil {
			return backUpErr
		}
	}

	return app.server.Shutdown(ctx)
}

func (app App) dbHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	if app.config.DBDsn == "" {
		return
	}

	db, ok := app.storage.(storage.DBStorage)
	if !ok {
		app.logger.Error().Msg("Can not get db instance from storage interface")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		app.logger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (app App) getAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, tErr := template.New("metrics").Parse(AllMetricsTemplate)
	if tErr != nil {
		app.logger.Error().Msgf("Template creating error: %s", tErr)
		http.Error(w, tErr.Error(), http.StatusInternalServerError)
	}

	m, err := app.storage.Find(metricOnPage, 0)
	if err != nil {
		app.logger.Error().Msgf("Error while getting metrics list: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := t.Execute(w, m); err != nil {
		app.logger.Error().Msgf("Template render error: %s", tErr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app App) updateJSONMetricsHandler(w http.ResponseWriter, r *http.Request) {
	var b metric.MetricsBatch

	if decodeErr := b.Decode(r.Body); decodeErr != nil {
		app.logger.Error().Msgf("Body decode error: %s", decodeErr)
		sendJSONResponse(w, http.StatusBadRequest, []byte(decodeErr.Error()), app.logger)
		return
	}

	var s []metric.IMetric
	for _, m := range b.Metrics {
		if m.Delta == nil && m.Value == nil {
			app.logger.Error().Msg("Invalid metric. Delta and Value can't be nil at the same time.")
			sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric value"), app.logger)
			return
		}

		if app.config.NeedCheckSign() {
			ok, signErr := m.CheckSign(app.config.Key)
			if signErr != nil {
				app.logger.Error().Msgf("Sign check error: %s", signErr)
				sendJSONResponse(w, http.StatusInternalServerError, []byte("check sign error"), app.logger)
				return
			}
			if !ok {
				app.logger.Error().Msgf("Metric wrong sign %v", m)
				sendJSONResponse(w, http.StatusBadRequest, []byte("wrong sign"), app.logger)
				return
			}
		}

		im, convertErr := m.ToIMetric()
		if convertErr != nil {
			app.logger.Error().Msgf("Metric convert error %s", convertErr)
			sendJSONResponse(w, http.StatusBadRequest, []byte(convertErr.Error()), app.logger)
			return
		}

		s = append(s, im)
	}

	if updErr := app.storage.BatchUpdate(s); updErr != nil {
		app.logger.Error().Msgf("Update error %s", updErr)
		sendJSONResponse(w, http.StatusInternalServerError, []byte("update error"), app.logger)
		return
	}

	sendJSONResponse(w, http.StatusOK, []byte("{}"), app.logger)
}

func (app App) updateJSONMetricHandler(w http.ResponseWriter, r *http.Request) {
	var m metric.Metrics

	if decodeErr := m.Decode(r.Body); decodeErr != nil {
		app.logger.Error().Msgf("Decode metric error: %s", decodeErr)
		sendJSONResponse(w, http.StatusBadRequest, []byte(decodeErr.Error()), app.logger)
		return
	}

	if m.Delta == nil && m.Value == nil {
		app.logger.Error().Msg("Invalid metric. Delta and Value can't be nil at the same time.")
		sendJSONResponse(w, http.StatusBadRequest, []byte("invalid metric value"), app.logger)
		return
	}

	if app.config.NeedCheckSign() {
		ok, signErr := m.CheckSign(app.config.Key)
		if signErr != nil {
			app.logger.Error().Msgf("Sign check error: %s", signErr)
			sendJSONResponse(w, http.StatusInternalServerError, []byte("check sign error"), app.logger)
			return
		}
		if !ok {
			app.logger.Error().Msgf("Metric wrong sign %v", m)
			sendJSONResponse(w, http.StatusBadRequest, []byte("wrong sign"), app.logger)
			return
		}
	}

	im, convertErr := m.ToIMetric()
	if convertErr != nil {
		app.logger.Error().Msgf("Metric convert error %s", convertErr)
		sendJSONResponse(w, http.StatusBadRequest, []byte(convertErr.Error()), app.logger)
		return
	}

	updM, updErr := app.storage.Update(im)
	if updErr != nil {
		app.logger.Error().Msgf("Metric update error: %s", updErr)
		sendJSONResponse(w, http.StatusInternalServerError, []byte("update error"), app.logger)
		return
	}

	rm, rmErr := metric.NewMetricsFromIMetric(updM)
	if rmErr != nil {
		app.logger.Error().Msgf("Metric convert error %s", rmErr)
		sendJSONResponse(w, http.StatusInternalServerError, []byte("update error"), app.logger)
		return
	}

	if app.config.NeedCheckSign() {
		if rmSignErr := rm.Sign(app.config.Key); rmSignErr != nil {
			app.logger.Error().Msgf("Check sign error: %s", rmSignErr)
			sendJSONResponse(w, http.StatusInternalServerError, []byte("check sign error"), app.logger)
			return
		}
	}

	b, err := rm.Encode()
	if err != nil {
		app.logger.Error().Msgf("Metric json encode error %s", err)
		sendJSONResponse(w, http.StatusInternalServerError, []byte("error while response creation"), app.logger)
		return
	}

	sendJSONResponse(w, http.StatusOK, b, app.logger)
}

func (app App) getJSONMetricHandler(w http.ResponseWriter, r *http.Request) {
	var rm metric.Metrics

	if decodeErr := rm.Decode(r.Body); decodeErr != nil {
		app.logger.Error().Msgf("Body decode error: %s", decodeErr)
		sendJSONResponse(w, http.StatusBadRequest, []byte(decodeErr.Error()), app.logger)
		return
	}

	m, err := app.storage.Get(rm.ID)
	if err != nil && errors.Is(storage.ErrMetricNotFound, err) {
		app.logger.Error().Msgf("Metric find error %s", err)
		sendJSONResponse(w, http.StatusNotFound, []byte(err.Error()), app.logger)
		return
	}

	resM, rmErr := metric.NewMetricsFromIMetric(m)
	if rmErr != nil {
		app.logger.Error().Msgf("Metric convert error %s", rmErr)
		sendJSONResponse(w, http.StatusInternalServerError, []byte("internal error"), app.logger)
		return
	}

	if app.config.NeedCheckSign() {
		_ = resM.Sign(app.config.Key)
	}

	if app.config.NeedCheckSign() {
		if resSignErr := resM.Sign(app.config.Key); resSignErr != nil {
			app.logger.Error().Msgf("Check sign error: %s", resSignErr)
			sendJSONResponse(w, http.StatusInternalServerError, []byte("check sign error"), app.logger)
			return
		}
	}

	b, mErr := resM.Encode()
	if mErr != nil {
		sendJSONResponse(w, http.StatusInternalServerError, []byte("internal server error"), app.logger)
		app.logger.Error().Msgf("Metric marshaling error: %s", mErr)
		return
	}

	sendJSONResponse(w, http.StatusOK, b, app.logger)
}

func (app App) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mName := chi.URLParam(r, "Name")
	mType := chi.URLParam(r, "Type")
	mValue := chi.URLParam(r, "Value")

	if mName == "" || mType == "" {
		app.logger.Error().Msgf("Invalid path params. Name: %s, Type: %s", mName, mType)
		http.Error(w, "invalid path params", http.StatusBadRequest)
		return
	}

	m, mErr := metric.NewMetric(mName, mType, mValue)
	if errors.Is(metric.ErrInvalidValue, mErr) {
		app.logger.Error().Msgf("Metric invalid error value: %s", mErr)
		http.Error(w, mErr.Error(), http.StatusBadRequest)
		return
	}
	if mErr != nil {
		app.logger.Error().Msgf("Create metric error: %s", mErr)
		http.Error(w, mErr.Error(), http.StatusNotImplemented)
		return
	}

	_, updErr := app.storage.Update(m)
	if updErr != nil {
		app.logger.Error().Msgf("Update metric error: %s", updErr)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) getMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mName := chi.URLParam(r, "Name")
	if mName == "" {
		app.logger.Error().Msgf("Invalid path param. Name: %s", mName)
		http.Error(w, "invalid path params", http.StatusBadRequest)
		return
	}

	m, vErr := app.storage.Get(mName)
	if vErr != nil {
		app.logger.Error().Msgf("Metric not found by name: %s", mName)
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(m.ValueAsString()))
	if err != nil {
		app.logger.Error().Msgf("Create response error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app App) getRouter() chi.Router {
	return app.router
}

func sendJSONResponse(w http.ResponseWriter, status int, body []byte, logger zerolog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		logger.Fatal().Msgf("Error while sending response: %s", err)
	}
	logger.Debug().Msgf("Send response with headers %s and body %s", w.Header(), string(body))
}
