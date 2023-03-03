package fs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/1g0rbm/sysmonitor/internal/metric"
	"github.com/1g0rbm/sysmonitor/internal/storage"
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
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0777)
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

func DumpStorage(ms storage.MemStorage, filepath string) (err error) {
	mw, err := NewMetricWriter(filepath)
	if err != nil {
		return err
	}

	defer func(mw *MetricWriter) {
		closeErr := mw.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}(mw)

	for _, m := range ms {
		if err := mw.Write(m); err != nil {
			log.Println(fmt.Sprintf("save metric %s error: %s", m.Name(), err))
		}
	}

	closeErr := mw.Close()
	if closeErr != nil {
		return closeErr
	}

	return nil
}
