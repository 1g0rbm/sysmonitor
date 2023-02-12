package handlers

import (
	"fmt"
	"github.com/1g0rbm/sysmonitor/internal/metric/names"
	"github.com/1g0rbm/sysmonitor/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func RegisterUpdateHandler(r *chi.Mux) {
	r.Post("/update/{Type}/{Name}/{Value}", updateHandler)
}

var saveError error

func updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	u := r.URL
	fmt.Println(u)

	mType := chi.URLParam(r, "Type")
	mName := chi.URLParam(r, "Name")
	mValue := chi.URLParam(r, "Value")

	switch mType {
	case "gauge":
		saveError = updateGauge(mName, mValue)
	case "counter":
		saveError = updateCounter(mName, mValue)
	default:
		http.Error(w, "invalid metric type", http.StatusNotImplemented)
		return
	}

	if saveError != nil {
		http.Error(w, saveError.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updateGauge(name string, value string) error {
	v, err := strToGauge(value)
	if err != nil {
		return err
	}
	return storage.GetMemStorage().SetGauge(name, v)
}

func updateCounter(name string, value string) error {
	v, err := strToCounter(value)
	if err != nil {
		return err
	}

	curVal, _ := storage.GetMemStorage().GetCounter(name)

	return storage.GetMemStorage().SetCounter(name, curVal+v)
}

func strToGauge(value string) (names.Gauge, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return names.Gauge(v), nil
}

func strToCounter(value string) (names.Counter, error) {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return names.Counter(v), nil
}
