package storage

import (
	"errors"
	"fmt"
	"github.com/1g0rbm/sysmonitor/internal/metric/names"
)

type gaugeMetricStorage map[string]names.Gauge
type counterMetricStorage map[string]names.Counter

type MemStorage struct {
	gaugeMetrics   gaugeMetricStorage
	counterMetrics counterMetricStorage
}

var storage *MemStorage

func GetMemStorage() *MemStorage {
	if storage != nil {
		return storage
	}

	return &MemStorage{
		gaugeMetrics:   map[string]names.Gauge{},
		counterMetrics: map[string]names.Counter{},
	}
}

func (ms *MemStorage) SetGauge(name string, value names.Gauge) error {
	if name == "" {
		return errors.New(fmt.Sprintf("name can't be blank"))
	}

	ms.gaugeMetrics[name] = value

	return nil
}

func (ms *MemStorage) SetCounter(name string, value names.Counter) error {
	if name == "" {
		return errors.New(fmt.Sprintf("name can't be blank"))
	}

	ms.counterMetrics[name] = value

	return nil
}

func (ms *MemStorage) GetCounter(name string) (names.Counter, bool) {
	v, ok := ms.counterMetrics[name]
	if !ok {
		return 0, false
	}

	return v, true
}

func (ms *MemStorage) GetGauge(name string) (names.Gauge, bool) {
	v, ok := ms.gaugeMetrics[name]
	if !ok {
		return 0, false
	}

	return v, true
}
