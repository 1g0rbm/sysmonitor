package metric

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	Hash  string   `json:"hash,omitempty"`
}

type MetricsBatch struct {
	Metrics []Metrics `json:""`
}

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

var (
	ErrInvalidValue = fmt.Errorf("invalid value")
)

func NewGaugeMetric(name string, value Gauge) GaugeMetric {
	return GaugeMetric{
		name:  name,
		value: value,
	}
}

func NewCounterMetric(name string, value Counter) CounterMetric {
	return CounterMetric{
		name:  name,
		value: value,
	}
}

func NewMetrics(id string, mType string, delta *int64, value *float64) (Metrics, error) {
	if mType == CounterType && delta == nil {
		return Metrics{}, fmt.Errorf("delata can not be nil for a counter type")
	} else if mType == GaugeType && value == nil {
		return Metrics{}, fmt.Errorf("value can not be nil for a gauge type")
	}

	return Metrics{
		ID:    id,
		MType: mType,
		Delta: delta,
		Value: value,
	}, nil
}

func NewMetricsFromIMetric(m IMetric) (Metrics, error) {
	switch m.Type() {
	case GaugeType:
		val, err := strconv.ParseFloat(m.ValueAsString(), 64)
		if err != nil {
			return Metrics{}, err
		}
		return NewMetrics(m.Name(), m.Type(), nil, &val)
	case CounterType:
		val, err := strconv.ParseInt(m.ValueAsString(), 10, 64)
		if err != nil {
			return Metrics{}, err
		}
		return NewMetrics(m.Name(), m.Type(), &val, nil)
	default:
		return Metrics{}, fmt.Errorf("invalid metric type")
	}
}

func (m *Metrics) Sign(key string) error {
	var s string
	switch m.MType {
	case GaugeType:
		s = fmt.Sprintf("%s:%s:%f", m.ID, m.MType, *m.Value)
	case CounterType:
		s = fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta)
	default:
		return fmt.Errorf("invalid metric type %s", m.MType)
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))

	m.Hash = hex.EncodeToString(h.Sum(nil))

	return nil
}

func (m *Metrics) CheckSign(key string) (bool, error) {
	var s string
	switch m.MType {
	case GaugeType:
		s = fmt.Sprintf("%s:%s:%f", m.ID, m.MType, *m.Value)
	case CounterType:
		s = fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta)
	default:
		return false, fmt.Errorf("invalid metric type %s", m.MType)
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))

	hash := hex.EncodeToString(h.Sum(nil))

	return m.Hash == hash, nil
}

func (m *Metrics) Decode(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return err
	}

	if m.MType != GaugeType && m.MType != CounterType {
		return fmt.Errorf("invalid metric type")
	}

	return nil
}

func (m *Metrics) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func (slm *MetricsBatch) Decode(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, &slm.Metrics); err != nil {
		return err
	}

	return nil
}

func (m *Metrics) ToIMetric() (IMetric, error) {
	var value string
	switch m.MType {
	case GaugeType:
		value = fmt.Sprintf("%v", *m.Value)
	case CounterType:
		value = fmt.Sprintf("%d", *m.Delta)
	default:
		return nil, fmt.Errorf("undefined metric type")
	}

	return NewMetric(m.ID, m.MType, value)
}

func NewMetric(name string, mType string, value string) (IMetric, error) {
	switch mType {
	case GaugeType:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, ErrInvalidValue
		}
		return GaugeMetric{
			name:  name,
			value: Gauge(val),
		}, nil
	case CounterType:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, ErrInvalidValue
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
	fl := float64(gm.value)
	str := strconv.FormatFloat(fl, 'f', -1, 64)
	if fl-math.Floor(fl) > 0 {
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
	newVal, newValErr := strconv.ParseInt(ncm.ValueAsString(), 10, 64)
	if newValErr != nil {
		return nil, newValErr
	}

	nv := cm.Value() + Counter(newVal)
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
