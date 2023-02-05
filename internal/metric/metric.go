package metric

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

type gauge float64
type counter int64

type MetricInterface interface {
	getPath() string
}

type Metric struct {
	Name  string
	Value gauge
}

func (m Metric) getPath() string {
	return fmt.Sprintf("update/guage/%v/%f", m.Name, m.Value)
}

type Counter struct {
	Name  string
	Value counter
}

func (c Counter) getPath() string {
	return fmt.Sprintf("update/counter/%v/%d", c.Name, c.Value)
}

type stats struct {
	MemStats    runtime.MemStats
	PollCounter counter
}

var metricValueExtractors = map[string]func(runtime.MemStats) gauge{
	"Alloc":         func(m runtime.MemStats) gauge { return gauge(m.Alloc) },
	"BuckHashSys":   func(m runtime.MemStats) gauge { return gauge(m.BuckHashSys) },
	"Frees":         func(m runtime.MemStats) gauge { return gauge(m.Frees) },
	"GCCPUFraction": func(m runtime.MemStats) gauge { return gauge(m.GCCPUFraction) },
	"GCSys":         func(m runtime.MemStats) gauge { return gauge(m.GCSys) },
	"HeapAlloc":     func(m runtime.MemStats) gauge { return gauge(m.HeapAlloc) },
	"HeapIdle":      func(m runtime.MemStats) gauge { return gauge(m.HeapIdle) },
	"HeapInuse":     func(m runtime.MemStats) gauge { return gauge(m.HeapInuse) },
	"HeapObjects":   func(m runtime.MemStats) gauge { return gauge(m.HeapObjects) },
	"HeapReleased":  func(m runtime.MemStats) gauge { return gauge(m.HeapReleased) },
	"HeapSys":       func(m runtime.MemStats) gauge { return gauge(m.HeapSys) },
	"LastGC":        func(m runtime.MemStats) gauge { return gauge(m.LastGC) },
	"Lookups":       func(m runtime.MemStats) gauge { return gauge(m.Lookups) },
	"MCacheInuse":   func(m runtime.MemStats) gauge { return gauge(m.MCacheInuse) },
	"MCacheSys":     func(m runtime.MemStats) gauge { return gauge(m.MCacheSys) },
	"MSpanInuse":    func(m runtime.MemStats) gauge { return gauge(m.MSpanInuse) },
	"MSpanSys":      func(m runtime.MemStats) gauge { return gauge(m.MSpanSys) },
	"Mallocs":       func(m runtime.MemStats) gauge { return gauge(m.Mallocs) },
	"NextGC":        func(m runtime.MemStats) gauge { return gauge(m.NextGC) },
	"NumForcedGC":   func(m runtime.MemStats) gauge { return gauge(m.NumForcedGC) },
	"NumGC":         func(m runtime.MemStats) gauge { return gauge(m.NumGC) },
	"OtherSys":      func(m runtime.MemStats) gauge { return gauge(m.OtherSys) },
	"PauseTotalNs":  func(m runtime.MemStats) gauge { return gauge(m.PauseTotalNs) },
	"StackInuse":    func(m runtime.MemStats) gauge { return gauge(m.StackInuse) },
	"StackSys":      func(m runtime.MemStats) gauge { return gauge(m.StackSys) },
	"Sys":           func(m runtime.MemStats) gauge { return gauge(m.Sys) },
	"TotalAlloc":    func(m runtime.MemStats) gauge { return gauge(m.TotalAlloc) },
	"RandomValue":   func(m runtime.MemStats) gauge { return gauge(rand.Float64()) },
}

func Update() {
	mc := &[]MetricInterface{}

	s := stats{PollCounter: 0}

	updMetricsTicker := time.NewTicker(time.Second * 2)
	sendMetricsticer := time.NewTicker(time.Second * 10)

	url := url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
	}

	for {
		select {
		case <-updMetricsTicker.C:
			runtime.ReadMemStats(&s.MemStats)
			s.PollCounter += 1

			mc = &[]MetricInterface{}
			for name, extract := range metricValueExtractors {
				*mc = append(*mc, Metric{
					Name:  name,
					Value: extract(s.MemStats),
				})
			}
			*mc = append(*mc, Counter{
				Name:  "PollCounter",
				Value: s.PollCounter,
			})
		case <-sendMetricsticer.C:
			for _, m := range *mc {
				url.Path = m.getPath()
				fmt.Printf("Encoded URL is %q\n", url.String())

				response := sendMetrics(url.String())

				fmt.Printf("Response status code: %d\n", response.StatusCode)
				response.Body.Close()
			}
		}
	}
}

func sendMetrics(url string) *http.Response {
	client := &http.Client{}

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	return response
}
