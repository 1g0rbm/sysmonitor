package metric

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

var (
	ErrInvalidValue = fmt.Errorf("invalid value")
)

func NewMetric(name string, mType string, value string) (IMetric, error) {
	switch mType {
	case GaugeType:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return GaugeMetric{}, ErrInvalidValue
		}
		return GaugeMetric{
			name:  name,
			value: Gauge(val),
		}, nil
	case CounterType:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return CounterMetric{}, ErrInvalidValue
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
	f := float64(gm.value) - math.Floor(float64(gm.value))
	str := fmt.Sprintf("%f", gm.value)
	if f > 0 {
		return strings.TrimRight(str, "0")
	}

	return str
}

func (gm GaugeMetric) NormalizeValue() (Gauge, error) {
	val, err := strconv.ParseFloat(gm.ValueAsString(), 64)
	if err != nil {
		return 0, err
	}
	return Gauge(val), nil
}

func (gm GaugeMetric) Update(ngm GaugeMetric) GaugeMetric {
	return ngm
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

func (cm CounterMetric) Update(ncm CounterMetric) (IMetric, error) {
	m, err := NewMetric(cm.Name(), cm.Type(), fmt.Sprintf("%d", cm.Value()+ncm.Value()))
	if err != nil {
		return nil, err
	}

	return m, nil
}
