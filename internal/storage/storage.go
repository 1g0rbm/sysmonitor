package storage

import (
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type Type int

type Storage interface {
	Set(m metric.IMetric)
	Get(name string) (metric.IMetric, error)
	All() map[string]metric.IMetric
	Update(m metric.IMetric) (metric.IMetric, error)
	GetCounter(name string) (metric.CounterMetric, error)
	GetGauge(name string) (metric.GaugeMetric, error)
}

const (
	DBStorageType Type = iota
	MemStorageType
)
