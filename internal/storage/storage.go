package storage

import (
	"github.com/1g0rbm/sysmonitor/internal/config"
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

type CloseStorage func() error

func NewStorage(cfg *config.ServerConfig) (Storage, CloseStorage, error) {
	if cfg.DBDsn != "" {
		s, cls, dbErr := NewDBStorage("pgx", cfg.DBDsn)
		if dbErr != nil {
			return nil, nil, dbErr
		}
		return s, cls, nil
	} else {
		return NewMemStorage(), func() error { return nil }, nil
	}
}
