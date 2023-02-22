package storage

import (
	"fmt"
	"github.com/1g0rbm/sysmonitor/internal/metric"
	"strconv"
)

type Storage interface {
	Set(m metric.IMetric)
	Get(name string) (metric.IMetric, bool)
	All() map[string]metric.IMetric
	Update(m metric.IMetric) error
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

func (ms MemStorage) Update(m metric.IMetric) error {
	switch m.Type() {
	case metric.CounterType:
		cm, ok := m.(metric.CounterMetric)
		if !ok {
			return fmt.Errorf("impossible to cast ")
		}

		curV, curVErr := cm.NormalizeValue()
		if curVErr != nil {
			return curVErr
		}

		em, emOk := ms.GetCounter(m.Name())
		if !emOk {
			ms.Set(m)
			return nil
		}

		emv, emvErr := em.NormalizeValue()
		if emvErr != nil {
			return emvErr
		}

		updM, updErr := metric.NewMetric(m.Name(), m.Type(), []byte(strconv.FormatInt(int64(curV+emv), 10)))
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
