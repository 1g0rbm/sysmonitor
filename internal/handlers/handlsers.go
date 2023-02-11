package handlers

import (
	"errors"
	"github.com/1g0rbm/sysmonitor/internal/metric/names"
	"github.com/1g0rbm/sysmonitor/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

func RegisterUpdateHandler(f *http.ServeMux) {
	f.HandleFunc("/update/", updateHandler)
}

var saveError error

func updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	mType, mName, mValue, err := extractMetricFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch mType {
	case "gauge":
		saveError = updateGauge(mName, mValue)
	case "counter":
		saveError = updateCounter(mName, mValue)
	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if saveError != nil {
		http.Error(w, saveError.Error(), http.StatusInternalServerError)
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

func extractMetricFromPath(path string) (mType string, mName string, mValue string, err error) {
	p := strings.Split(path, "/")

	if len(p) != 5 {
		return "", "", "", errors.New("can not extract data from path")
	}

	mType = p[2]
	mName = p[3]
	mValue = p[4]

	return
}
