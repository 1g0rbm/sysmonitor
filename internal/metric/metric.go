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
	Gauge() *Gauge
	Counter() *Counter
	Update(m IMetric) (IMetric, error)
}

type GaugeMetric struct {
	name  string
	value Gauge
}

type CounterMetric struct {
	name  string
	value Counter
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

var (
	ErrInvalidValue = fmt.Errorf("invalid value")
)

func NewMetrics(id string, mType string, delta *int64, value *float64) IMetric {
	return Metrics{
		ID:    id,
		MType: mType,
		Delta: delta,
		Value: value,
	}
}

func (m Metrics) Name() string {
	return m.ID
}

func (m Metrics) Type() string {
	return m.MType
}

func (m Metrics) ValueAsString() string {
	switch m.MType {
	case GaugeType:
		f := *m.Value - math.Floor(*m.Value)
		str := fmt.Sprintf("%f", *m.Value)
		if f > 0 {
			return strings.TrimRight(str, "0")
		}
		return str
	case CounterType:
		return fmt.Sprintf("%d", *m.Delta)
	default:
		return ""
	}
}

func (m Metrics) Gauge() *Gauge {
	if m.Value == nil {
		return nil
	}

	res := Gauge(*m.Value)

	return &res
}

func (m Metrics) Counter() *Counter {
	if m.Delta == nil {
		return nil
	}

	res := Counter(*m.Delta)

	return &res
}

func (m Metrics) Update(nm IMetric) (IMetric, error) {
	if m.Type() != nm.Type() {
		return nil, fmt.Errorf("invalid tytpe")
	}

	switch m.Type() {
	case GaugeType:
		return nm, nil
	case CounterType:
		upd := *m.Delta + int64(*nm.Counter())
		return NewMetrics(nm.Name(), nm.Type(), &upd, nil), nil
	default:
		return nil, fmt.Errorf("undefined tytpe")
	}
}

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

func (gm GaugeMetric) Update(ngm IMetric) (IMetric, error) {
	return ngm, nil
}

func (gm GaugeMetric) Gauge() *Gauge {
	return &gm.value
}

func (gm GaugeMetric) Counter() *Counter {
	return nil
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

func (cm CounterMetric) Update(ncm IMetric) (IMetric, error) {
	nv := cm.Value() + *ncm.Counter()
	snv := fmt.Sprintf("%d", nv)

	m, err := NewMetric(cm.Name(), cm.Type(), snv)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (cm CounterMetric) Gauge() *Gauge {
	return nil
}

func (cm CounterMetric) Counter() *Counter {
	return &cm.value
}
