package watcher

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

const (
	scheme         string = "http"
	clientTimeout         = 10 * time.Second
	requestTimeout        = 5 * time.Second
)

type sender struct {
	config *config.AgentConfig
	jobCh  <-chan *Job
	errCh  chan<- error
}

func newSender(config *config.AgentConfig, jobCh <-chan *Job, errCh chan<- error) sender {
	return sender{config, jobCh, errCh}
}

func (s *sender) Run(ctx context.Context) {
	updURL := url.URL{
		Scheme: scheme,
		Host:   s.config.Address,
	}
	updURL.Path = "/updates/"

	for i := 0; i < s.config.RateLimit; i++ {
		go func() {
			for {
				select {
				case job := <-s.jobCh:
					if err := s.sendMetrics(updURL.String(), job.batch); err != nil {
						s.errCh <- err
					} else {
						fmt.Printf("%d metrics was sent successfull\n", len(job.batch.Metrics))
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (s *sender) sendMetrics(url string, b metric.MetricsBatch) error {
	client := &http.Client{
		Timeout: clientTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	d, mErr := b.Encode()
	if mErr != nil {
		return mErr
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")

	response, rErr := client.Do(request.WithContext(ctx))
	if rErr != nil {
		return rErr
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response code %d", response.StatusCode)
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
