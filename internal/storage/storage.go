package storage

import (
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type Storage interface {
	Set(m metric.IMetric)
	Get(name string) (metric.IMetric, bool)
	All() map[string]metric.IMetric
	GetCounter(name string) (metric.CounterMetric, bool)
	GetGauge(name string) (metric.GaugeMetric, bool)
}

type MemStorage map[string]metric.IMetric

func (ms MemStorage) All() map[string]metric.IMetric {
	return ms
}

func (ms MemStorage) Set(m metric.IMetric) {
	ms[m.Name()] = m
}

func (ms MemStorage) Get(name string) (metric.IMetric, bool) {
	v, ok := ms[name]
	if !ok {
		return nil, false
	}

	return v, true
}

func (ms MemStorage) GetCounter(name string) (metric.CounterMetric, bool) {
	v, ok := ms[name]
	if !ok {
		return metric.CounterMetric{}, false
	}

	if v.Type() != metric.CounterType {
		return metric.CounterMetric{}, false
	}

	t, _ := v.(metric.CounterMetric)

	return t, true
}

func (ms MemStorage) GetGauge(name string) (metric.GaugeMetric, bool) {
	v, ok := ms[name]
	if !ok {
		return metric.GaugeMetric{}, false
	}

	if v.Type() != metric.GaugeType {
		return metric.GaugeMetric{}, false
	}

	t, _ := v.(metric.GaugeMetric)

	return t, true
}

func newMemStorage() MemStorage {
	return make(MemStorage)
}

func NewStorage() Storage {
	return newMemStorage()
}
