package storage

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/1g0rbm/sysmonitor/internal/fs"
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

// MemStorage ToDo Data Race
type MemStorage struct {
	data map[string]metric.IMetric
	mu   *sync.RWMutex
}

func (ms MemStorage) Find(limit int, offset int) (map[string]metric.IMetric, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if offset < 0 || offset >= len(ms.data) || limit < 0 {
		return nil, errors.New("invalid input parameters")
	}

	cnt := 0
	result := make(map[string]metric.IMetric)

	for key, m := range ms.data {
		if cnt >= offset {
			result[key] = m
			if len(result) == limit {
				break
			}
		}
		cnt++
	}

	return result, nil
}

func (ms MemStorage) Get(name string) (metric.IMetric, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	v, ok := ms.data[name]
	if !ok {
		ErrMetricNotFound = fmt.Errorf("metric not found by name '%s'", name)
		return nil, ErrMetricNotFound
	}

	return v, nil
}

func (ms MemStorage) GetCounter(name string) (metric.CounterMetric, error) {
	v, ok := ms.data[name]
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
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	v, ok := ms.data[name]
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
	ms.mu.Lock()
	defer ms.mu.Unlock()

	switch m.Type() {
	case metric.CounterType:
		em, emErr := ms.Get(m.Name())
		if emErr != nil && errors.Is(ErrMetricNotFound, emErr) {
			ms.data[m.Name()] = m
			return m, nil
		}
		if emErr != nil {
			return nil, emErr
		}

		updM, updErr := m.Update(em)
		if updErr != nil {
			return nil, updErr
		}

		ms.data[updM.Name()] = updM
		return updM, nil
	case metric.GaugeType:
		ms.data[m.Name()] = m
		return m, nil
	default:
		return nil, fmt.Errorf("undefined metric type '%s'", m.Type())
	}
}

func (ms MemStorage) BatchUpdate(sm []metric.IMetric) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, m := range sm {
		_, err := ms.Update(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms MemStorage) Restore(filepath string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	mr, err := fs.NewMetricReader(filepath)

	defer func(mr *fs.MetricReader) {
		closeErr := mr.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}(mr)

	for {
		m, err := mr.Read()
		if err == io.EOF {
			break
		}

		im, imErr := m.ToIMetric()
		if imErr != nil {
			return imErr
		}

		_, err = ms.Update(im)
		if err != nil {
			return err
		}
	}

	return
}

func (ms MemStorage) BackupData(path string) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	mw, err := fs.NewMetricWriter(path)
	if err != nil {
		return err
	}

	for _, m := range ms.data {
		metrics, imErr := metric.NewMetricsFromIMetric(m)
		if imErr != nil {
			return imErr
		}
		if writeErr := mw.Write(metrics); writeErr != nil {
			return writeErr
		}
	}

	if closeErr := mw.Close(); closeErr != nil {
		return closeErr
	}

	return nil
}

func newMemStorage() MemStorage {
	return MemStorage{
		data: make(map[string]metric.IMetric),
	}
}

func NewMemStorage() Storage {
	return newMemStorage()
}
