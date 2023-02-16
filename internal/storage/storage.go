package storage

import (
	"github.com/1g0rbm/sysmonitor/internal/metric"
	"strconv"
)

type Storage interface {
	Set(name string, val []byte) error
	Get(name string) ([]byte, bool)
	All() map[string][]byte
}

type TStorage interface {
	SType(sType string) (Storage, bool)
}

type GaugeMetricStorage map[string]metric.Gauge
type CounterMetricStorage map[string]metric.Counter

type MemStorage struct {
	gaugeMetrics   GaugeMetricStorage
	counterMetrics CounterMetricStorage
}

func (ms MemStorage) SType(sType string) (Storage, bool) {
	switch sType {
	case "gauge":
		return ms.gaugeMetrics, true
	case "counter":
		return ms.counterMetrics, true
	default:
		return nil, false
	}
}

func (gs GaugeMetricStorage) All() map[string][]byte {
	res := map[string][]byte{}
	for name, v := range gs {
		res[name] = []byte(strconv.FormatFloat(float64(v), 'f', -1, 64))
	}

	return res
}

func (cs CounterMetricStorage) All() map[string][]byte {
	res := map[string][]byte{}
	for name, v := range cs {
		res[name] = []byte(strconv.FormatInt(int64(v), 10))
	}

	return res
}

func (gs GaugeMetricStorage) Set(name string, val []byte) error {
	float, err := strconv.ParseFloat(string(val), 64)
	if err != nil {
		return err
	}

	gs[name] = metric.Gauge(float)

	return nil
}

func (cs CounterMetricStorage) Set(name string, val []byte) error {
	v, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return err
	}

	cs[name] += metric.Counter(v)

	return nil
}

func (gs GaugeMetricStorage) Get(name string) ([]byte, bool) {
	v, ok := gs[name]
	if !ok {
		return []byte(""), false
	}

	return []byte(strconv.FormatFloat(float64(v), 'f', -1, 64)), true
}

func (cs CounterMetricStorage) Get(name string) ([]byte, bool) {
	v, ok := cs[name]
	if !ok {
		return []byte(""), false
	}

	return []byte(strconv.FormatInt(int64(v), 10)), true
}

var storage TStorage

func NewStorage() TStorage {
	if storage != nil {
		return storage
	}

	return MemStorage{
		gaugeMetrics:   map[string]metric.Gauge{},
		counterMetrics: map[string]metric.Counter{},
	}
}
