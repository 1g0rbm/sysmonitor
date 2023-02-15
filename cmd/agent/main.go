package main

import (
	"github.com/1g0rbm/sysmonitor/internal/metric"
)

const (
	updMetricsDuration  int = 2
	sendMetricsDuration int = 10
)

func main() {
	metric.Update(updMetricsDuration, sendMetricsDuration)
}
