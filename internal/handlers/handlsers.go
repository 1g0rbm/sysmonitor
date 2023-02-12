package handlers

import (
	"fmt"
	"github.com/1g0rbm/sysmonitor/internal/metric/names"
	"github.com/1g0rbm/sysmonitor/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func RegisterUpdateHandler(r *chi.Mux, s *storage.MemStorage) {
	r.Post("/update/{Type}/{Name}/{Value}", func(w http.ResponseWriter, r *http.Request) {
		var saveError error

		w.Header().Set("Content-Type", "text/plain")

		mType := chi.URLParam(r, "Type")
		mName := chi.URLParam(r, "Name")
		mValue := chi.URLParam(r, "Value")

		switch mType {
		case "gauge":
			saveError = updateGauge(mName, mValue, s)
		case "counter":
			saveError = updateCounter(mName, mValue, s)
		default:
			http.Error(w, "invalid metric type", http.StatusNotImplemented)
			return
		}

		if saveError != nil {
			http.Error(w, saveError.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func RegisterGetOneHandler(r *chi.Mux, s *storage.MemStorage) {
	r.Get("/value/{Type}/{Name}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		mType := chi.URLParam(r, "Type")
		mName := chi.URLParam(r, "Name")

		switch mType {
		case "gauge":
			val, ok := s.GetGauge(mName)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			_, err := w.Write([]byte(fmt.Sprintf("%v", val)))
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		case "counter":
			val, ok := s.GetCounter(mName)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			_, err := w.Write([]byte(fmt.Sprintf("%d", val)))
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "invalid metric type", http.StatusNotImplemented)
			return
		}
	})
}

func updateGauge(name string, value string, s *storage.MemStorage) error {
	v, err := strToGauge(value)
	if err != nil {
		return err
	}
	return s.SetGauge(name, v)
}

func updateCounter(name string, value string, s *storage.MemStorage) error {
	v, err := strToCounter(value)
	if err != nil {
		return err
	}

	curVal, _ := s.GetCounter(name)

	return s.SetCounter(name, curVal+v)
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
