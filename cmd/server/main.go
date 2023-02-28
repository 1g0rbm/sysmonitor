package main

import (
	"github.com/1g0rbm/sysmonitor/internal/config"
	"log"

	"github.com/1g0rbm/sysmonitor/internal/application"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

func main() {
	cfg := config.GetConfigServer()
	s := storage.NewStorage()
	app := application.NewApp(s, cfg)

	log.Fatal(app.Run())
}
