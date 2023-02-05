package main

import (
	"log"
	"net/http"

	"github.com/1g0rbm/sysmonitor/internal/handlers"
)

func main() {
	handlers.RegisterUpdateHndler(http.DefaultServeMux)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
