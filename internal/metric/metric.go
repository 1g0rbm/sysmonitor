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
	Value() []byte
	ValueAsString() string
}

type GaugeMetric struct {
	name  string
	mType string
	value []byte
}

type CounterMetric struct {
	name  string
	mType string
	value []byte
}

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

func NewMetric(name string, mType string, value []byte) (IMetric, error) {
	switch mType {
	case GaugeType:
		return GaugeMetric{
			name:  name,
			value: value,
			mType: mType,
		}, nil
	case CounterType:
		return CounterMetric{
			name:  name,
			value: value,
			mType: mType,
		}, nil
	default:
		return nil, fmt.Errorf("invalid type %s", mType)
	}
}

func (gm GaugeMetric) Name() string {
	return gm.name
}

func (gm GaugeMetric) Type() string {
	return gm.mType
}

func (gm GaugeMetric) Value() []byte {
	return gm.value
}

func (gm GaugeMetric) ValueAsString() string {
	return string(gm.value)
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
	return cm.mType
}

func (cm CounterMetric) Value() []byte {
	return cm.value
}

func (cm CounterMetric) ValueAsString() string {
	return string(cm.value)
}

func (cm CounterMetric) NormalizeValue() (Counter, error) {
	val, err := strconv.ParseInt(cm.ValueAsString(), 10, 64)
	if err != nil {
		return 0, err
	}
	return Counter(val), nil
}
