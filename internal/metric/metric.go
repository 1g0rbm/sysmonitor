package metric

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"

	"github.com/1g0rbm/sysmonitor/internal/metric/names"
)

type stats struct {
	MemStats    runtime.MemStats
	PollCounter names.Counter
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

var mc = &[]sentM{}

func Update() {
	s := stats{PollCounter: 0}

	updMetricsTicker := time.NewTicker(time.Second * 2)
	sendMetricsTicker := time.NewTicker(time.Second * 10)

	updUrl := url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
	}

	for {
		select {
		case <-updMetricsTicker.C:
			runtime.ReadMemStats(&s.MemStats)
			s.PollCounter += 1

			mc = &[]sentM{}
			for _, name := range names.GaugeMetrics {
				*mc = append(*mc, sentGM{*names.NewGaugeMetric(name, getMemStatValue(s.MemStats, name))})
			}
			*mc = append(*mc, sentCM{*names.NewCounterMetric(names.PollCounter, s.PollCounter)})
		case <-sendMetricsTicker.C:
			for _, m := range *mc {
				updUrl.Path = m.ToURLPath()
				sc, err := sendMetrics(updUrl.String())
				if err != nil {
					fmt.Println(err)
					return
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
