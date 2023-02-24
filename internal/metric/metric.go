package metric

import (
	"fmt"
	"strconv"
)

type Gauge float64
type Counter int64

type IMetric interface {
	Name() string
	Type() string
	ValueAsString() string
}

type GaugeMetric struct {
	name  string
	value Gauge
}

type CounterMetric struct {
	name  string
	value Counter
}

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

func NewMetric(name string, mType string, value string) (IMetric, error) {
	switch mType {
	case GaugeType:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return GaugeMetric{}, err
		}
		return GaugeMetric{
			name:  name,
			value: Gauge(val),
		}, nil
	case CounterType:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return CounterMetric{}, err
		}
		return CounterMetric{
			name:  name,
			value: Counter(val),
		}, nil
	default:
		return nil, fmt.Errorf("invalid type %s", mType)
	}
}

func (gm GaugeMetric) Name() string {
	return gm.name
}

func (gm GaugeMetric) Type() string {
	return GaugeType
}

func (gm GaugeMetric) Value() Gauge {
	return gm.value
}

func (gm GaugeMetric) ValueAsString() string {
	return fmt.Sprintf("%f", gm.value)
}

func (gm GaugeMetric) NormalizeValue() (Gauge, error) {
	val, err := strconv.ParseFloat(gm.ValueAsString(), 64)
	if err != nil {
		return 0, err
	}
	return Gauge(val), nil
}

func (cm CounterMetric) Name() string {
	return cm.name
}

func (cm CounterMetric) Type() string {
	return CounterType
}

func (cm CounterMetric) Value() Counter {
	return cm.value
}

func (cm CounterMetric) ValueAsString() string {
	return fmt.Sprintf("%d", cm.value)
}

func (cm CounterMetric) NormalizeValue() (Counter, error) {
	val, err := strconv.ParseInt(cm.ValueAsString(), 10, 64)
	if err != nil {
		return 0, err
	}
	return Counter(val), nil
}
