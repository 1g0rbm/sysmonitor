package watcher

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

const (
	scheme         string = "http"
	clientTimeout         = 10 * time.Second
	requestTimeout        = 5 * time.Second
)

type gMetrics struct {
	m  map[string]metric.Gauge
	mu sync.RWMutex
}

func (gm *gMetrics) update(m runtime.MemStats) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.m["Alloc"] = metric.Gauge(m.Alloc)
	gm.m["BuckHashSys"] = metric.Gauge(m.BuckHashSys)
	gm.m["Frees"] = metric.Gauge(m.Frees)
	gm.m["GCCPUFraction"] = metric.Gauge(m.GCCPUFraction)
	gm.m["GCSys"] = metric.Gauge(m.GCSys)
	gm.m["HeapAlloc"] = metric.Gauge(m.HeapAlloc)
	gm.m["HeapIdle"] = metric.Gauge(m.HeapIdle)
	gm.m["HeapInuse"] = metric.Gauge(m.HeapInuse)
	gm.m["HeapObjects"] = metric.Gauge(m.HeapObjects)
	gm.m["HeapReleased"] = metric.Gauge(m.HeapReleased)
	gm.m["HeapSys"] = metric.Gauge(m.HeapSys)
	gm.m["LastGC"] = metric.Gauge(m.LastGC)
	gm.m["Lookups"] = metric.Gauge(m.Lookups)
	gm.m["MCacheInuse"] = metric.Gauge(m.MCacheInuse)
	gm.m["MCacheSys"] = metric.Gauge(m.MCacheSys)
	gm.m["MSpanInuse"] = metric.Gauge(m.MSpanInuse)
	gm.m["MSpanSys"] = metric.Gauge(m.MSpanSys)
	gm.m["Mallocs"] = metric.Gauge(m.Mallocs)
	gm.m["NextGC"] = metric.Gauge(m.NextGC)
	gm.m["NumForcedGC"] = metric.Gauge(m.NumForcedGC)
	gm.m["NumGC"] = metric.Gauge(m.NumGC)
	gm.m["OtherSys"] = metric.Gauge(m.OtherSys)
	gm.m["PauseTotalNs"] = metric.Gauge(m.PauseTotalNs)
	gm.m["StackInuse"] = metric.Gauge(m.StackInuse)
	gm.m["StackSys"] = metric.Gauge(m.StackInuse)
	gm.m["Sys"] = metric.Gauge(m.Sys)
	gm.m["TotalAlloc"] = metric.Gauge(m.TotalAlloc)
	gm.m["RandomValue"] = metric.Gauge(rand.Float64())
}

func (gm *gMetrics) updateAdditional(v *mem.VirtualMemoryStat) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.m["TotalMemory"] = metric.Gauge(v.Total)
	gm.m["FreeMemory"] = metric.Gauge(v.Free)
	gm.m["CPUutilization1"] = metric.Gauge(v.UsedPercent)
}

type cMetrics map[string]metric.Counter

func (cm cMetrics) update() {
	cm["PollCount"] += 1
}

type Watcher struct {
	cm     cMetrics
	gm     gMetrics
	config *config.AgentConfig
}

func NewWatcher(cfg *config.AgentConfig) Watcher {
	return Watcher{
		gm: gMetrics{
			m: map[string]metric.Gauge{
				"Alloc":           0,
				"BuckHashSys":     0,
				"Frees":           0,
				"GCCPUFraction":   0,
				"GCSys":           0,
				"HeapAlloc":       0,
				"HeapIdle":        0,
				"HeapInuse":       0,
				"HeapObjects":     0,
				"HeapReleased":    0,
				"HeapSys":         0,
				"LastGC":          0,
				"Lookups":         0,
				"MCacheInuse":     0,
				"MCacheSys":       0,
				"MSpanInuse":      0,
				"MSpanSys":        0,
				"Mallocs":         0,
				"NextGC":          0,
				"NumForcedGC":     0,
				"NumGC":           0,
				"OtherSys":        0,
				"PauseTotalNs":    0,
				"StackInuse":      0,
				"StackSys":        0,
				"Sys":             0,
				"TotalAlloc":      0,
				"RandomValue":     0,
				"TotalMemory":     0,
				"FreeMemory":      0,
				"CPUutilization1": 0,
			},
		},
		cm: cMetrics{
			"PollCount": 0,
		},
		config: cfg,
	}
}

func (w *Watcher) update() {
	var rms runtime.MemStats
	runtime.ReadMemStats(&rms)

	w.cm.update()
	w.gm.update(rms)
}

func (w *Watcher) updateAdditional() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	w.gm.updateAdditional(v)

	return nil
}

func (w *Watcher) getAll() (metric.MetricsBatch, error) {
	w.gm.mu.RLock()
	defer w.gm.mu.RUnlock()

	var mb metric.MetricsBatch

	for name, value := range w.gm.m {
		v := float64(value)
		m, _ := metric.NewMetrics(name, metric.GaugeType, nil, &v)
		if w.config.NeedSign() {
			sgnErr := m.Sign(w.config.Key)
			if sgnErr != nil {
				return metric.MetricsBatch{}, sgnErr
			}
		}
		mb.Metrics = append(mb.Metrics, m)
	}

	for name, value := range w.cm {
		v := int64(value)
		m, _ := metric.NewMetrics(name, metric.CounterType, &v, nil)
		if w.config.NeedSign() {
			sgnErr := m.Sign(w.config.Key)
			if sgnErr != nil {
				return metric.MetricsBatch{}, sgnErr
			}
		}
		mb.Metrics = append(mb.Metrics, m)
	}

	return mb, nil
}

func (w *Watcher) Run() error {
	if w.config.PollInterval >= w.config.ReportInterval {
		errMsg := fmt.Sprintf(
			"update duration (%d) should be less than send duration (%d)",
			w.config.PollInterval,
			w.config.ReportInterval)
		return errors.New(errMsg)
	}

	updMetricsTicker := time.NewTicker(w.config.PollInterval)
	sendMetricsTicker := time.NewTicker(w.config.ReportInterval)

	metricChan := make(chan metric.MetricsBatch)
	errChan := make(chan error)

	var updGroup sync.WaitGroup

	for {
		select {
		case <-updMetricsTicker.C:
			updGroup.Add(2)
			go func() {
				w.update()
				updGroup.Done()
			}()
			go func() {
				if err := w.updateAdditional(); err != nil {
					errChan <- err
				}
				updGroup.Done()
			}()
			updGroup.Wait()
			go w.poll(metricChan, errChan)
		case <-sendMetricsTicker.C:
			for i := 0; i <= w.config.RateLimit; i += 1 {
				go w.send(metricChan)
			}
		case err := <-errChan:
			fmt.Printf("ERROR: %s \n", err)
		}
	}
}

func (w *Watcher) poll(mbChan chan<- metric.MetricsBatch, errChan chan<- error) {
	mb, err := w.getAll()
	if err != nil {
		errChan <- err
	}

	mbChan <- mb
}

func (w *Watcher) send(mb <-chan metric.MetricsBatch) {
	batch := <-mb

	updUrl := url.URL{
		Scheme: scheme,
		Host:   w.config.Address,
	}

	updUrl.Path = "/updates/"
	if err := sendMetrics(updUrl.String(), batch); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%d metrics was sent successfull\n", len(batch.Metrics))
	}
}

func sendMetrics(url string, b metric.MetricsBatch) error {
	client := &http.Client{
		Timeout: clientTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	d, mErr := b.Encode()
	if mErr != nil {
		return mErr
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")

	response, rErr := client.Do(request.WithContext(ctx))
	if rErr != nil {
		return rErr
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
