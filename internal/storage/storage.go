package storage

import (
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type Storage interface {
	Set(m metric.IMetric)
	Get(name string) (metric.IMetric, bool)
	All() map[string]metric.IMetric
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

func newMemStorage() MemStorage {
	return make(MemStorage)
}

func NewStorage() Storage {
	return newMemStorage()
}
