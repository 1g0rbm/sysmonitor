package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/1g0rbm/sysmonitor/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func GetAllHandler(ts storage.TStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := map[string]string{}
		sg, ok := ts.SType("gauge")
		if !ok {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		for key, v := range sg.All() {
			nv, _ := strconv.ParseFloat(string(v), 64)
			a[key] = fmt.Sprintf("%v", nv)
		}

		sc, ok := ts.SType("counter")
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
}

func UpdateHandler(ts storage.TStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/plain")

		mType := chi.URLParam(r, "Type")
		mName := chi.URLParam(r, "Name")
		mValue := chi.URLParam(r, "Value")

		s, ok := ts.SType(mType)
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
}

func GetOneHandler(ts storage.TStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		mType := chi.URLParam(r, "Type")
		mName := chi.URLParam(r, "Name")

		s, sOk := ts.SType(mType)
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
