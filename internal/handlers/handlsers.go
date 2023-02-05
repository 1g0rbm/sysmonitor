package handlers

import (
	"fmt"
	"net/http"
)

func RegisterUpdateHndler(f *http.ServeMux) {
	f.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
	})
}
