package watcher

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type poller struct {
	config *config.AgentConfig
	cm     cMetrics
	gm     gMetrics
	jobCh  chan<- *Job
	errCh  chan<- error
}

func newPoller(config *config.AgentConfig, jobCh chan<- *Job, errCh chan<- error) poller {
	return poller{
		config: config,
		gm:     newGMetrics(),
		cm:     newCMetrics(),
		jobCh:  jobCh,
		errCh:  errCh,
	}
}

func (p *poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.config.PollInterval)
	go p.updateBasicMetrics(ctx, ticker)
	go p.updateAdditionalMetrics(ctx, ticker)
	go p.writeBatch(ctx, ticker)
}

func (p *poller) writeBatch(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			batch, err := p.getBatch()
			if err != nil {
				p.errCh <- err
			}

			p.jobCh <- &Job{batch}
		case <-ctx.Done():
			return
		}
	}
}

func (p *poller) updateBasicMetrics(ctx context.Context, ticker *time.Ticker) {
	var rms runtime.MemStats
	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&rms)
			p.cm.update()
			p.gm.update(rms)
		case <-ctx.Done():
			return
		}
	}
}

func (p *poller) updateAdditionalMetrics(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			vm, err := mem.VirtualMemory()
			if err != nil {
				p.errCh <- err
			}
			p.gm.updateAdditional(vm)
		case <-ctx.Done():
			return
		}
	}
}

func (p *poller) getBatch() (metric.MetricsBatch, error) {
	p.gm.mu.RLock()
	defer p.gm.mu.RUnlock()

	var mb metric.MetricsBatch

	for name, value := range p.gm.m {
		v := float64(value)
		m, _ := metric.NewMetrics(name, metric.GaugeType, nil, &v)
		if p.config.NeedSign() {
			sgnErr := m.Sign(p.config.Key)
			if sgnErr != nil {
				return metric.MetricsBatch{}, sgnErr
			}
		}
		mb.Metrics = append(mb.Metrics, m)
	}

	for name, value := range p.cm {
		v := int64(value)
		m, _ := metric.NewMetrics(name, metric.CounterType, &v, nil)
		if p.config.NeedSign() {
			sgnErr := m.Sign(p.config.Key)
			if sgnErr != nil {
				return metric.MetricsBatch{}, sgnErr
			}
		}
		mb.Metrics = append(mb.Metrics, m)
	}

	return mb, nil
}
