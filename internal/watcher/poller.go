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
}

func newPoller(config *config.AgentConfig) poller {
	return poller{
		config: config,
		gm:     newGMetrics(),
		cm:     newCMetrics(),
	}
}

func (p *poller) Run(jobCh chan<- *Job, errCh chan<- error, ctx context.Context) {
	ticker := time.NewTicker(p.config.PollInterval)
	go p.updateBasicMetrics(ticker, ctx)
	go p.updateAdditionalMetrics(ticker, errCh, ctx)
	go p.writeBatch(ticker, jobCh, errCh, ctx)
}

func (p *poller) writeBatch(ticker *time.Ticker, jobCh chan<- *Job, errCh chan<- error, ctx context.Context) {
	for {
		select {
		case <-ticker.C:
			batch, err := p.getBatch()
			if err != nil {
				errCh <- err
			}

			jobCh <- &Job{batch}
		case <-ctx.Done():
			return
		}
	}
}

func (p *poller) updateBasicMetrics(ticker *time.Ticker, ctx context.Context) {
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

func (p *poller) updateAdditionalMetrics(ticker *time.Ticker, errCh chan<- error, ctx context.Context) {
	for {
		select {
		case <-ticker.C:
			vm, err := mem.VirtualMemory()
			if err != nil {
				errCh <- err
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
