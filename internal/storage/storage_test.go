package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/1g0rbm/sysmonitor/internal/metric"
)

func TestRun(t *testing.T) {
	m1, _ := metric.NewMetric("GCSys", "gauge", []byte("2621440.000000"))
	m2, _ := metric.NewMetric("PollCounter", "counter", []byte("10"))

	tests := []struct {
		name string
		data map[string]metric.IMetric
	}{
		{
			name: "Success setting metrics",
			data: map[string]metric.IMetric{m1.Name(): m1, m2.Name(): m2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStorage()
			for _, m := range tt.data {
				s.Set(m)
			}

			m1, ok1 := s.Get("GCSys")
			assert.True(t, ok1)
			assert.Equal(t, tt.data["GCSys"], m1)

			m2, ok2 := s.Get("PollCounter")
			assert.True(t, ok2)
			assert.Equal(t, tt.data["PollCounter"], m2)

			m3, ok3 := s.Get("Undefined")
			assert.False(t, ok3)
			assert.Empty(t, m3)

			ms := s.All()
			assert.Len(t, ms, 2)
			assert.Equal(t, tt.data, ms)
		})
	}
}
