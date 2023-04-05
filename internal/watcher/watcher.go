package watcher

import (
	"errors"
	"fmt"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type Job struct {
	batch metric.MetricsBatch
}

type Watcher struct {
	poller poller
	sender sender
	config *config.AgentConfig
}

func NewWatcher(cfg *config.AgentConfig) Watcher {
	return Watcher{
		poller: newPoller(cfg),
		sender: newSender(cfg),
		config: cfg,
	}
}

func (w *Watcher) Run() error {
	if w.config.PollInterval >= w.config.ReportInterval {
		errMsg := fmt.Sprintf(
			"update duration (%d) should be less than send duration (%d)",
			w.config.PollInterval,
			w.config.ReportInterval)
		return errors.New(errMsg)
	}

	jobCh := make(chan *Job, w.config.RateLimit)
	errChan := make(chan error)

	w.poller.Run(jobCh, errChan)
	w.sender.Run(jobCh, errChan)

	for {
		select {
		case err := <-errChan:
			fmt.Printf("ERROR: %s\n", err)
		}
	}
}
