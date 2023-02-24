package main

import (
	"log"

	"github.com/1g0rbm/sysmonitor/internal/application"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

func main() {
	s := storage.NewStorage()
	app := application.NewApp(s)

	log.Fatal(app.Run())
}
