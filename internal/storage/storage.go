package storage

import (
	"errors"
	"fmt"

	"github.com/1g0rbm/sysmonitor/internal/metric"
)

var ErrMetricNotFound error

type Storage interface {
	Set(m metric.IMetric)
	Get(name string) (metric.IMetric, error)
	All() map[string]metric.IMetric
	Update(m metric.IMetric) error
	GetCounter(name string) (metric.CounterMetric, error)
	GetGauge(name string) (metric.GaugeMetric, error)
}

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

func (ms MemStorage) Update(m metric.IMetric) error {
	switch m.Type() {
	case metric.CounterType:
		cm, ok := m.(metric.CounterMetric)
		if !ok {
			return fmt.Errorf("impossible to cast to target type")
		}

		em, emErr := ms.GetCounter(m.Name())
		if emErr != nil {
			if errors.Is(ErrMetricNotFound, emErr) {
				ms.Set(m)
				return nil
			} else {
				return emErr
			}
		}

		updM, updErr := cm.Update(em)
		if updErr != nil {
			return updErr
		}

		ms.Set(updM)
		return nil
	case metric.GaugeType:
		ms.Set(m)
		return nil
	default:
		return fmt.Errorf("undefined metric type '%s'", m.Type())
	}
}

func newMemStorage() MemStorage {
	return make(MemStorage)
}

func NewStorage() Storage {
	return newMemStorage()
}
