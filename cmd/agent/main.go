package main

import (
	"log"
	"os"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/watcher"
)

func main() {
	w := watcher.NewWatcher()

	agentConfig := config.GetConfigAgent(os.Args[1:])

	log.Fatal(w.Run(agentConfig))
}
