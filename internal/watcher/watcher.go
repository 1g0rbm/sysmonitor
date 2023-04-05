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
	config *config.AgentConfig
	logger zerolog.Logger
}

func NewWatcher(cfg *config.AgentConfig, logger zerolog.Logger) Watcher {
	return Watcher{
		poller: newPoller(cfg),
		sender: newSender(cfg),
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

	jobCh := make(chan *Job, w.config.RateLimit)
	errChan := make(chan error)

	w.poller.Run(jobCh, errChan, ctx)
	w.sender.Run(jobCh, errChan, ctx)

	for {
		select {
		case err := <-errChan:
			w.logger.Error().Msg(err.Error())
		case <-ctx.Done():
			return nil
		}
	}
}
