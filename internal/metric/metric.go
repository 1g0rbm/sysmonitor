package metric

import (
	"errors"
	"fmt"
	"strconv"
)

type Gauge float64
type Counter int64

type IMetric interface {
	Name() string
	Type() string
	Value() []byte
	ValueAsString() string
}

type Metric struct {
	name  string
	mType string
	value []byte
}

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

func NewMetric(name string, mType string, value []byte) (IMetric, error) {
	if mType != GaugeType && mType != CounterType {
		return nil, errors.New(fmt.Sprintf("invalid type %s", mType))
	}

	return Metric{
		name:  name,
		value: value,
		mType: mType,
	}, nil
}

func (m Metric) Name() string {
	return m.name
}

func (m Metric) Type() string {
	return m.mType
}

func (m Metric) Value() []byte {
	return m.value
}

func (m Metric) ValueAsString() string {
	return string(m.value)
}

func NormalizeGaugeMetricValue(m IMetric) (Gauge, error) {
	if m.Type() != GaugeType {
		return 0, errors.New(fmt.Sprintf("expect gauge type, but %s was passed", m.Type()))
	}

	val, err := strconv.ParseFloat(m.ValueAsString(), 64)
	if err != nil {
		return 0, err
	}
	return Gauge(val), nil
}

func NormalizeCounterMetricValue(m IMetric) (Counter, error) {
	if m.Type() != CounterType {
		return 0, errors.New(fmt.Sprintf("expect counter type, but %s was passed", m.Type()))
	}

	val, err := strconv.ParseInt(m.ValueAsString(), 10, 64)
	if err != nil {
		return 0, err
	}
	return Counter(val), nil
}
