package storage

import (
	"errors"
	"fmt"

	"github.com/1g0rbm/sysmonitor/internal/metric"
)

var ErrMetricNotFound error

// MemStorage ToDo Data Race
type MemStorage map[string]metric.IMetric

func (ms MemStorage) All() map[string]metric.IMetric {
	return ms
}

func (ms MemStorage) Set(m metric.IMetric) {
	ms[m.Name()] = m
}

func (ms MemStorage) Get(name string) (metric.IMetric, error) {
	v, ok := ms[name]
	if !ok {
		ErrMetricNotFound = fmt.Errorf("metric not found by name '%s'", name)
		return nil, ErrMetricNotFound
	}

	return v, nil
}

func (ms MemStorage) GetCounter(name string) (metric.CounterMetric, error) {
	v, ok := ms[name]
	if !ok {
		ErrMetricNotFound = fmt.Errorf("metric not found by name '%s'", name)
		return metric.CounterMetric{}, ErrMetricNotFound
	}

	if v.Type() != metric.CounterType {
		err := fmt.Errorf("metric should be a counter type, but a '%s' was found", v.Type())
		return metric.CounterMetric{}, err
	}

	t, _ := v.(metric.CounterMetric)

	return t, nil
}

func (ms MemStorage) GetGauge(name string) (metric.GaugeMetric, error) {
	v, ok := ms[name]
	if !ok {
		ErrMetricNotFound = fmt.Errorf("metric not found by name '%s'", name)
		return metric.GaugeMetric{}, ErrMetricNotFound
	}

	if v.Type() != metric.GaugeType {
		err := fmt.Errorf("metric should be a gauge type, but a '%s' was found", v.Type())
		return metric.GaugeMetric{}, err
	}

	t, _ := v.(metric.GaugeMetric)

	return t, nil
}

func (ms MemStorage) Update(m metric.IMetric) (metric.IMetric, error) {
	switch m.Type() {
	case metric.CounterType:
		em, emErr := ms.Get(m.Name())
		if emErr != nil && errors.Is(ErrMetricNotFound, emErr) {
			ms.Set(m)
			return m, nil
		}
		if emErr != nil {
			return nil, emErr
		}

		updM, updErr := m.Update(em)
		if updErr != nil {
			return nil, updErr
		}

		ms.Set(updM)
		return updM, nil
	case metric.GaugeType:
		ms.Set(m)
		return m, nil
	default:
		return nil, fmt.Errorf("undefined metric type '%s'", m.Type())
	}
}

func (ms MemStorage) Health() error {
	return nil
}

func newMemStorage() MemStorage {
	return make(MemStorage)
}

func NewMemStorage() Storage {
	return newMemStorage()
}
