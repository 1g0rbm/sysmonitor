package watcher

import (
	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkSender_SendMetrics(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := &sender{
		config: &config.AgentConfig{
			Address: server.URL,
		},
	}

	v := 10.0
	d := int64(1)
	batch := metric.MetricsBatch{
		Metrics: []metric.Metrics{
			{ID: "metric1", MType: "gauge", Value: &v},
			{ID: "metric2", MType: "counter", Delta: &d},
		},
	}
	url := server.URL + "/updates/"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.sendMetrics(url, batch)
		if err != nil {
			b.Errorf("Failed to send metrics: %v", err)
		}
	}
}

func TestSender_SendMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Unexpected request method: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Unexpected content type: %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := &sender{
		config: &config.AgentConfig{
			Address: server.URL,
		},
	}
	v := 10.0
	d := int64(1)
	batch := metric.MetricsBatch{
		Metrics: []metric.Metrics{
			{ID: "metric1", MType: "gauge", Value: &v},
			{ID: "metric2", MType: "counter", Delta: &d},
		},
	}

	err := s.sendMetrics(server.URL, batch)
	if err != nil {
		t.Errorf("Failed to send metrics: %v", err)
	}
}
