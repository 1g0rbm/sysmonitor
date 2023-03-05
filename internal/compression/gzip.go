package compression

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func NewGzipResponseWriter(rw http.ResponseWriter) *GzipResponseWriter {
	rw.Header().Set("Content-Encoding", "gzip")
	gw := gzip.NewWriter(rw)
	defer gw.Close()
	return &GzipResponseWriter{
		ResponseWriter: rw,
		Writer:         gw,
	}
}

func (grw GzipResponseWriter) Write(b []byte) (int, error) {
	return grw.Writer.Write(b)
}
