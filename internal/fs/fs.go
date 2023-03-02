package fs

import (
	"encoding/json"
	"os"

	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type MetricWriter struct {
	file    *os.File
	encoder *json.Encoder
}

type MetricReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewMetricWriter(filepath string) (*MetricWriter, error) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	return &MetricWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (mw *MetricWriter) Write(m metric.IMetric) error {
	return mw.encoder.Encode(m)
}

func (mw *MetricWriter) Close() error {
	return mw.file.Close()
}

func NewMetricReader(filepath string) (*MetricReader, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &MetricReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (mr *MetricReader) Read() (metric.IMetric, error) {
	m := metric.Metrics{}
	err := mr.decoder.Decode(&m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (mr *MetricReader) Close() error {
	return mr.file.Close()
}
