package middleware

import (
	"compress/gzip"
	"github.com/1g0rbm/sysmonitor/internal/compression"
	"net/http"
	"strings"
)

func Gzip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gw := gzip.NewWriter(w)
		defer func(gw *gzip.Writer) {
			if err := gw.Close(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}(gw)

		grw := compression.NewGzipResponseWriter(w, gw)

		h.ServeHTTP(grw, r)

	})
}
