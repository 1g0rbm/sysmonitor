package storage

import (
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type Type int

type Storage interface {
	Get(name string) (metric.IMetric, error)
	Find(limit int, offset int) (map[string]metric.IMetric, error)
	Update(m metric.IMetric) (metric.IMetric, error)
	BatchUpdate(sm []metric.IMetric) error
}

var ErrMetricNotFound error

func NewStorage(dsn string) (Storage, error) {
	if dsn != "" {
		s, dbErr := NewDBStorage("pgx", dsn)
		if dbErr != nil {
			return nil, dbErr
		}
		return s, nil
	} else {
		return NewMemStorage(), nil
	}
}
