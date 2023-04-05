package watcher

import (
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

func (p *poller) Run(jobCh chan<- *Job, errCh chan<- error) {
	ticker := time.NewTicker(p.config.PollInterval)
	go p.updateBasicMetrics(ticker)
	go p.updateAdditionalMetrics(ticker, errCh)
	go p.writeBatch(ticker, jobCh, errCh)
}

func (p *poller) writeBatch(ticker *time.Ticker, jobCh chan<- *Job, errCh chan<- error) {
	for {
		select {
		case <-ticker.C:
			batch, err := p.getBatch()
			if err != nil {
				errCh <- err
			}

			jobCh <- &Job{batch}
		}
	}
}

func (p *poller) updateBasicMetrics(ticker *time.Ticker) {
	var rms runtime.MemStats
	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&rms)
			p.cm.update()
			p.gm.update(rms)
		}
	}
}

func (p *poller) updateAdditionalMetrics(ticker *time.Ticker, errCh chan<- error) {
	for {
		select {
		case <-ticker.C:
			vm, err := mem.VirtualMemory()
			if err != nil {
				errCh <- err
			}
			p.gm.updateAdditional(vm)
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
