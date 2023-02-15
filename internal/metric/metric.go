package metric

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"

	"github.com/1g0rbm/sysmonitor/internal/metric/names"
)

const (
	scheme string = "http"
	host   string = "localhost:8080"
)

type stats struct {
	MemStats    runtime.MemStats
	PollCounter names.Counter
	SentM       map[string]sentM
}

type sentM interface {
	ToURLPath() string
}

type sentGM struct {
	names.GaugeMetric
}

func (gm sentGM) ToURLPath() string {
	return fmt.Sprintf("/update/gauge/%v/%f", gm.Name(), gm.Value())
}

type sentCM struct {
	names.CounterMetric
}

func (cm sentCM) ToURLPath() string {
	return fmt.Sprintf("/update/gauge/%v/%v", cm.Name(), cm.Value())
}

func Update(updMetricsDuration int, sendMetricsDuration int) error {
	if updMetricsDuration >= sendMetricsDuration {
		errMsg := fmt.Sprintf(
			"update duration (%d) should be less than send duration (%d)",
			updMetricsDuration,
			sendMetricsDuration)
		return errors.New(errMsg)
	}
	s := stats{
		PollCounter: 0,
		SentM:       map[string]sentM{},
	}

	updMetricsTicker := time.NewTicker(time.Second * time.Duration(updMetricsDuration))
	sendMetricsTicker := time.NewTicker(time.Second * time.Duration(sendMetricsDuration))

	updURL := url.URL{
		Scheme: scheme,
		Host:   host,
	}

	for {
		select {
		case <-updMetricsTicker.C:
			runtime.ReadMemStats(&s.MemStats)
			s.PollCounter += 1

			for _, name := range names.GaugeMetrics {
				s.SentM[name] = sentGM{*names.NewGaugeMetric(name, getMemStatValue(s.MemStats, name))}
			}
			s.SentM[names.PollCounter] = sentCM{*names.NewCounterMetric(names.PollCounter, s.PollCounter)}
		case <-sendMetricsTicker.C:
			for _, m := range s.SentM {
				updURL.Path = m.ToURLPath()
				sc, err := sendMetrics(updURL.String())
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Response status code: %d\n", sc)
			}
		}
	}
}

func getMemStatValue(m runtime.MemStats, name string) names.Gauge {
	val := reflect.ValueOf(m)
	f := val.FieldByName(name)

	if !val.IsValid() && name != names.RandomValue {
		panic("invalid metric name")
	}

	if name == names.RandomValue {
		return names.Gauge(rand.Float64())
	}

	switch f.Kind() {
	case reflect.Uint64:
		return names.Gauge(f.Uint())
	case reflect.Uint32:
		return names.Gauge(f.Uint())
	case reflect.Float64:
		return names.Gauge(f.Float())
	default:
		fmt.Println(name, f.Kind())
		panic("unknown metric type")
	}
}

func sendMetrics(url string) (int, error) {
	client := &http.Client{}

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, err
	}
	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	err = response.Body.Close()
	if err != nil {
		return 0, err
	}

	return response.StatusCode, nil
}
