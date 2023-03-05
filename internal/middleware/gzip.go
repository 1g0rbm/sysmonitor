package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func Gzip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = gz
			defer func(gz *gzip.Reader) {
				if err := gz.Close(); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}(gz)
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)

		h.ServeHTTP(w, r)

		defer func(gz *gzip.Writer) {
			if err := gz.Close(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}(gz)
	})
}
