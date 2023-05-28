package watcher

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/1g0rbm/sysmonitor/internal/config"
)

func BenchmarkPoller_GetBatch(b *testing.B) {
	cfg := &config.AgentConfig{
		PollInterval: time.Second,
		Key:          "secret",
	}
	jobCh := make(chan *Job)
	errCh := make(chan error)
	p := newPoller(cfg, jobCh, errCh)

	ticker := time.NewTicker(10 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	b.ReportAllocs()
	b.ResetTimer()
	go func() {
		p.updateBasicMetrics(ctx, ticker)
		p.updateAdditionalMetrics(ctx, ticker)
		_, _ = p.getBatch()
	}()

	<-ctx.Done()
}

func TestPoller_GetBatch(t *testing.T) {
	cfg := &config.AgentConfig{
		PollInterval: time.Second,
		Key:          "secret",
	}
	jobCh := make(chan *Job)
	errCh := make(chan error)
	p := newPoller(cfg, jobCh, errCh)

	ticker := time.NewTicker(10 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	go func() {
		p.updateBasicMetrics(ctx, ticker)
		p.updateAdditionalMetrics(ctx, ticker)
	}()

	<-ctx.Done()
	batch, err := p.getBatch()

	assert.NoError(t, err)
	assert.Equal(t, len(batch.Metrics), 32)
}
