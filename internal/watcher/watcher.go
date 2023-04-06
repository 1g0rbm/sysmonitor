package watcher

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type Job struct {
	batch metric.MetricsBatch
}

type Watcher struct {
	poller poller
	sender sender
	jobCh  chan *Job
	errCh  chan error
	config *config.AgentConfig
	logger zerolog.Logger
}

func NewWatcher(cfg *config.AgentConfig, logger zerolog.Logger) Watcher {
	jobCh := make(chan *Job, cfg.RateLimit)
	errCh := make(chan error)

	return Watcher{
		poller: newPoller(cfg, jobCh, errCh),
		sender: newSender(cfg, jobCh, errCh),
		jobCh:  jobCh,
		errCh:  errCh,
		config: cfg,
		logger: logger,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	if w.config.PollInterval >= w.config.ReportInterval {
		errMsg := fmt.Sprintf(
			"update duration (%d) should be less than send duration (%d)",
			w.config.PollInterval,
			w.config.ReportInterval)
		return errors.New(errMsg)
	}

	w.logger.Info().Msg("Agent started")

	w.poller.Run(ctx)
	w.sender.Run(ctx)

	for {
		select {
		case err := <-w.errCh:
			w.logger.Error().Msg(err.Error())
		case <-ctx.Done():
			return nil
		}
	}
}
