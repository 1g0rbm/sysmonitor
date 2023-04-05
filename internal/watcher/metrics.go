package watcher

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/1g0rbm/sysmonitor/internal/metric"
	"github.com/shirou/gopsutil/v3/mem"
)

type gMetrics struct {
	m  map[string]metric.Gauge
	mu sync.RWMutex
}

func newGMetrics() gMetrics {
	return gMetrics{
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
	}
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

func newCMetrics() cMetrics {
	return cMetrics{
		"PollCount": 0,
	}
}

func (cm cMetrics) update() {
	cm["PollCount"] += 1
}
