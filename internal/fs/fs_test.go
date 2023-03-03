package fs

import (
	"github.com/1g0rbm/sysmonitor/internal/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"testing"
)

func TestMetricReaderWriter(t *testing.T) {
	cv := rand.Int63()
	m, err := metric.NewMetrics("PollCount", metric.CounterType, &cv, nil)
	require.Nil(t, err)

	tests := []struct {
		name   string
		path   string
		metric metric.IMetric
	}{
		{
			name:   "Write metric test",
			path:   "/tmp/metric_test.json",
			metric: m,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, wErr := NewMetricWriter(tt.path)
			require.Nil(t, wErr)
			defer mw.Close()

			wErr = mw.Write(tt.metric)
			require.Nil(t, wErr)

			assert.FileExists(t, tt.path)

			mr, rErr := NewMetricReader(tt.path)
			require.Nil(t, rErr)
			defer mr.Close()

			m, mrErr := mr.Read()
			require.Nil(t, mrErr)

			assert.Equal(t, tt.metric, m)

			fErr := os.Remove(tt.path)
			require.Nil(t, fErr)
		})
	}
}
