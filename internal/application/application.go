package application

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/1g0rbm/sysmonitor/internal/metric"
	"github.com/1g0rbm/sysmonitor/internal/storage"
	"github.com/1g0rbm/sysmonitor/internal/tmp"
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

	if err := t.Execute(w, app.storage.All()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app App) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	mName := chi.URLParam(r, "Name")
	mType := chi.URLParam(r, "Type")
	mValue := chi.URLParam(r, "Value")

	fmt.Println(mValue)

	if mName == "" || mType == "" {
		http.Error(w, "invalid path params", http.StatusBadRequest)
		return
	}

	if ok, _ := regexp.MatchString("[0-9]+(.[0-9]+)?", mValue); !ok {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}

	m, mErr := metric.NewMetric(mName, mType, []byte(mValue))
	if mErr != nil {
		http.Error(w, mErr.Error(), http.StatusNotImplemented)
		return
	}

	updErr := app.storage.Update(m)
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

func (app App) getRouter() chi.Router {
	return app.router
}
