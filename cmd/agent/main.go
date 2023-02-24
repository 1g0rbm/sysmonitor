package main

import (
	"log"

	"github.com/1g0rbm/sysmonitor/internal/watcher"
)

const (
	updMetricsDuration  int = 2
	sendMetricsDuration int = 10
)

func main() {
	w := watcher.NewWatcher()

	log.Fatal(w.Run(updMetricsDuration, sendMetricsDuration))
}
