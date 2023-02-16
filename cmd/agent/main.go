package main

import (
	"github.com/1g0rbm/sysmonitor/internal/wathcer"
	"log"
)

const (
	updMetricsDuration  int = 2
	sendMetricsDuration int = 10
)

func main() {
	w := wathcer.NewWatcher()

	log.Fatal(w.Run(updMetricsDuration, sendMetricsDuration))
}
