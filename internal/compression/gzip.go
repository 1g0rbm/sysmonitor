package compression

import (
	"io"
	"net/http"
)

type GzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func NewGzipResponseWriter(rw http.ResponseWriter, gw io.Writer) GzipResponseWriter {
	rw.Header().Set("Content-Encoding", "gzip")

	return GzipResponseWriter{
		ResponseWriter: rw,
		Writer:         gw,
	}
}

func (grw GzipResponseWriter) Write(b []byte) (int, error) {
	return grw.Writer.Write(b)
}
