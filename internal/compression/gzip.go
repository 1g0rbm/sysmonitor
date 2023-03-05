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

func NewGzipResponseWriter(w http.ResponseWriter) *GzipResponseWriter {
	w.Header().Set("Content-Encoding", "gzip")
	return &GzipResponseWriter{
		ResponseWriter: w,
		Writer:         gzip.NewWriter(w),
	}
}

func (grw GzipResponseWriter) Write(b []byte) (int, error) {
	return grw.Writer.Write(b)
}
