package main

import (
	"log"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/watcher"
)

func main() {
	w := watcher.NewWatcher()

	agentConfig := config.GetConfigAgent()

	log.Fatal(w.Run(agentConfig))
}
