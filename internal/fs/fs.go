package fs

import (
	"encoding/json"
	"io"
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

func (mw *MetricWriter) Write(m metric.Metrics) error {
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

func (mr *MetricReader) Read() (metric.Metrics, error) {
	m := metric.Metrics{}
	err := mr.decoder.Decode(&m)
	if err != nil {
		return metric.Metrics{}, err
	}

	return m, nil
}

func (mr *MetricReader) Close() error {
	return mr.file.Close()
}

func DumpStorage(ms storage.MemStorage, filepath string) error {
	mw, err := NewMetricWriter(filepath)
	if err != nil {
		return err
	}

	for _, m := range ms {
		metrics, imErr := metric.NewMetricsFromIMetric(m)
		if imErr != nil {
			log.Printf("convert metric %s error: %s", m.Name(), err)
		}
		if err := mw.Write(metrics); err != nil {
			log.Printf("save metric %s error: %s", m.Name(), err)
		}
	}

	closeErr := mw.Close()
	if closeErr != nil {
		return closeErr
	}

	return nil
}

func RestoreStorage(ms storage.Storage, filepath string) (err error) {
	mr, err := NewMetricReader(filepath)

	defer func(mr *MetricReader) {
		closeErr := mr.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}(mr)

	for {
		m, err := mr.Read()
		if err == io.EOF {
			break
		}

		im, imErr := m.ToIMetric()
		if imErr != nil {
			return imErr
		}

		ms.Set(im)
	}

	return
}
