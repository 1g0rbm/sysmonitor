package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/1g0rbm/sysmonitor/internal/metric"
)

func TestRun(t *testing.T) {
	m1, _ := metric.NewMetric("GCSys", "gauge", "2621440.000000")
	m2, _ := metric.NewMetric("PollCounter", "counter", "10")

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

			m1, err1 := s.Get("GCSys")
			assert.Nil(t, err1)
			assert.Equal(t, tt.data["GCSys"], m1)

			m2, err2 := s.Get("PollCounter")
			assert.Nil(t, err2)
			assert.Equal(t, tt.data["PollCounter"], m2)

			m3, err3 := s.Get("Undefined")
			assert.Errorf(t, err3, "metric not found by name 'Undefined'")
			assert.Empty(t, m3)

			ms := s.All()
			assert.Len(t, ms, 2)
			assert.Equal(t, tt.data, ms)

			_, gUpdErr := s.Update(tt.data["GCSys"])
			assert.Nil(t, gUpdErr)

			m4, err4 := s.Get("GCSys")
			assert.Nil(t, err4)
			assert.Equal(t, tt.data["GCSys"], m4)

			_, cUpdErr := s.Update(tt.data["PollCounter"])
			assert.Nil(t, cUpdErr)

			m5, err5 := s.Get("PollCounter")
			assert.Nil(t, err5)
			assert.Equal(t, "20", m5.ValueAsString())
		})
	}
}
