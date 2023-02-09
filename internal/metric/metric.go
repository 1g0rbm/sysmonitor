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

func Update() {
	gmc := &[]names.GaugeMetric{}
	cmc := &[]names.CounterMetric{}

	s := stats{PollCounter: 0}

	updMetricsTicker := time.NewTicker(time.Second * 2)
	sendMetricsticer := time.NewTicker(time.Second * 10)

	url := url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
	}

	for {
		select {
		case <-updMetricsTicker.C:
			runtime.ReadMemStats(&s.MemStats)
			s.PollCounter += 1

			gmc = &[]names.GaugeMetric{}
			cmc = &[]names.CounterMetric{}
			for _, name := range names.GauageMetrics {
				*gmc = append(*gmc, names.GaugeMetric{
					Name:  name,
					Value: getMemStatValue(s.MemStats, name),
				})
			}
			*cmc = append(*cmc, names.CounterMetric{
				Name:  names.PollCounter,
				Value: s.PollCounter,
			})
		case <-sendMetricsticer.C:
			fmt.Println(*gmc)
			for _, m := range *gmc {
				url.Path = fmt.Sprintf("/update/gauge/%v/%f", m.Name, m.Value)
				sc := sendMetrics(url.String())
				fmt.Printf("Response status code: %d\n", sc)
			}
			for _, m := range *cmc {
				url.Path = fmt.Sprintf("/update/counter/%v/%v", m.Name, m.Value)
				sc := sendMetrics(url.String())
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

func sendMetrics(url string) int {
	client := &http.Client{}

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	response.Body.Close()

	return response.StatusCode
}
