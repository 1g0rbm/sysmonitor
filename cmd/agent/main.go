package main

import (
	"log"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/watcher"
)

func main() {
	agentConfig := config.GetConfigAgent()
	w := watcher.NewWatcher(agentConfig)

	log.Fatal(w.Run())
}
