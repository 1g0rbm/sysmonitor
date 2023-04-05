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
}

func newSender(config *config.AgentConfig) sender {
	return sender{config}
}

func (s *sender) Run(jobCh <-chan *Job, errCh chan<- error, ctx context.Context) {
	updURL := url.URL{
		Scheme: scheme,
		Host:   s.config.Address,
	}
	updURL.Path = "/updates/"
	ticker := time.NewTicker(s.config.ReportInterval)

	for {
		select {
		case <-ticker.C:
			for i := 0; i < s.config.RateLimit; i++ {
				go func() {
					for job := range jobCh {
						if err := s.sendMetrics(updURL.String(), job.batch); err != nil {
							errCh <- err
						} else {
							fmt.Printf("%d metrics was sent successfull\n", len(job.batch.Metrics))
						}
					}
				}()
			}
		case <-ctx.Done():
			return
		}
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
