package main

import (
	"log"

	"github.com/1g0rbm/sysmonitor/internal/application"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

func main() {
	app := application.NewApp(storage.NewStorage())

	log.Fatal(app.Run())
}
