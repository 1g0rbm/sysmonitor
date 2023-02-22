package watcher

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/1g0rbm/sysmonitor/internal/metric"
)

const (
	scheme string = "http"
	host   string = "localhost:8080"
)

type gMetrics map[string]metric.Gauge

func (gm gMetrics) update(m runtime.MemStats) {
	gm["Alloc"] = metric.Gauge(m.Alloc)
	gm["BuckHashSys"] = metric.Gauge(m.BuckHashSys)
	gm["Frees"] = metric.Gauge(m.Frees)
	gm["GCCPUFraction"] = metric.Gauge(m.GCCPUFraction)
	gm["GCSys"] = metric.Gauge(m.GCSys)
	gm["HeapAlloc"] = metric.Gauge(m.HeapAlloc)
	gm["HeapIdle"] = metric.Gauge(m.HeapIdle)
	gm["HeapInuse"] = metric.Gauge(m.HeapInuse)
	gm["HeapObjects"] = metric.Gauge(m.HeapObjects)
	gm["HeapReleased"] = metric.Gauge(m.HeapReleased)
	gm["HeapSys"] = metric.Gauge(m.HeapSys)
	gm["LastGC"] = metric.Gauge(m.LastGC)
	gm["Lookups"] = metric.Gauge(m.Lookups)
	gm["MCacheInuse"] = metric.Gauge(m.MCacheInuse)
	gm["MCacheSys"] = metric.Gauge(m.MCacheSys)
	gm["MSpanInuse"] = metric.Gauge(m.MSpanInuse)
	gm["MSpanSys"] = metric.Gauge(m.MSpanSys)
	gm["Mallocs"] = metric.Gauge(m.Mallocs)
	gm["NextGC"] = metric.Gauge(m.NextGC)
	gm["NumForcedGC"] = metric.Gauge(m.NumForcedGC)
	gm["NumGC"] = metric.Gauge(m.NumGC)
	gm["OtherSys"] = metric.Gauge(m.OtherSys)
	gm["PauseTotalNs"] = metric.Gauge(m.PauseTotalNs)
	gm["StackInuse"] = metric.Gauge(m.StackInuse)
	gm["StackSys"] = metric.Gauge(m.StackInuse)
	gm["Sys"] = metric.Gauge(m.Sys)
	gm["TotalAlloc"] = metric.Gauge(m.TotalAlloc)
	gm["RandomValue"] = metric.Gauge(rand.Float64())
}

type cMetrics map[string]metric.Counter

func (cm cMetrics) update() {
	cm["PollCounter"] += 1
}

type Watcher struct {
	cm cMetrics
	gm gMetrics
}

func NewWatcher() Watcher {
	return Watcher{
		gm: gMetrics{
			"Alloc":         0,
			"BuckHashSys":   0,
			"Frees":         0,
			"GCCPUFraction": 0,
			"GCSys":         0,
			"HeapAlloc":     0,
			"HeapIdle":      0,
			"HeapInuse":     0,
			"HeapObjects":   0,
			"HeapReleased":  0,
			"HeapSys":       0,
			"LastGC":        0,
			"Lookups":       0,
			"MCacheInuse":   0,
			"MCacheSys":     0,
			"MSpanInuse":    0,
			"MSpanSys":      0,
			"Mallocs":       0,
			"NextGC":        0,
			"NumForcedGC":   0,
			"NumGC":         0,
			"OtherSys":      0,
			"PauseTotalNs":  0,
			"StackInuse":    0,
			"StackSys":      0,
			"Sys":           0,
			"TotalAlloc":    0,
			"RandomValue":   0,
		},
		cm: cMetrics{
			"PollCounter": 0,
		},
	}
}

func (w Watcher) Run(updMetricsDuration int, sendMetricsDuration int) error {
	if updMetricsDuration >= sendMetricsDuration {
		errMsg := fmt.Sprintf(
			"update duration (%d) should be less than send duration (%d)",
			updMetricsDuration,
			sendMetricsDuration)
		return errors.New(errMsg)
	}

	updMetricsTicker := time.NewTicker(time.Second * time.Duration(updMetricsDuration))
	sendMetricsTicker := time.NewTicker(time.Second * time.Duration(sendMetricsDuration))

	updURL := url.URL{
		Scheme: scheme,
		Host:   host,
	}

	var rms runtime.MemStats

	for {
		select {
		case <-updMetricsTicker.C:
			runtime.ReadMemStats(&rms)

			w.cm.update()
			w.gm.update(rms)
		case <-sendMetricsTicker.C:
			for name, val := range w.gm {
				updURL.Path = fmt.Sprintf("/update/gauge/%v/%f", name, val)
				err := sendMetrics(updURL.String())
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Metric %s was sent successfull\n", name)
			}
			for name, val := range w.cm {
				updURL.Path = fmt.Sprintf("/update/counter/%v/%d", name, val)
				err := sendMetrics(updURL.String())
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Metric %s was sent successfull\n", name)
			}
		}
	}
}

func sendMetrics(url string) error {
	client := &http.Client{}

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response code %d", response.StatusCode)
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
