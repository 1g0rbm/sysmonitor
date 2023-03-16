package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/1g0rbm/sysmonitor/internal/metric"
)

type DBStorage struct {
	sql *sql.DB
}

func NewDBStorage(driverName string, dsn string) (Storage, CloseStorage, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = db.ExecContext(ctx, MetricsTable())
	if err != nil {
		return nil, nil, err
	}

	c := func() error { return db.Close() }

	return DBStorage{
		sql: db,
	}, c, nil
}

func (D DBStorage) Get(name string) (metric.IMetric, error) {
	var (
		id    string
		mType string
		delta *int64
		val   *float64
	)

	if err := D.sql.QueryRow(SelectMetric(), name).Scan(&id, &mType, &delta, &val); err != nil {
		return nil, err
	}

	switch mType {
	case metric.GaugeType:
		return metric.NewGaugeMetric(id, metric.Gauge(*val)), nil
	case metric.CounterType:
		return metric.NewCounterMetric(id, metric.Counter(*delta)), nil
	default:
		return nil, fmt.Errorf("invalid metric type: %s", mType)
	}
}

func (D DBStorage) All() (map[string]metric.IMetric, error) {
	var (
		id    string
		mType string
		delta *int64
		val   *float64
	)

	r, err := D.sql.Query(SelectMetric())
	if err != nil {
		return nil, err
	}

	defer func(r *sql.Rows) {
		err := r.Close()
		if err != nil {

		}
	}(r)

	ms := map[string]metric.IMetric{}
	for r.Next() {
		if err := r.Scan(&id, &mType, &delta, &val); err != nil {
			return nil, err
		}

		var m metric.IMetric
		switch mType {
		case metric.GaugeType:
			m = metric.NewGaugeMetric(id, metric.Gauge(*val))
		case metric.CounterType:
			m = metric.NewCounterMetric(id, metric.Counter(*delta))
		default:
			return nil, fmt.Errorf("invalid metric type: %s", mType)
		}

		ms[id] = m
	}

	return ms, nil
}

func (D DBStorage) Update(m metric.IMetric) (metric.IMetric, error) {
	switch m.Type() {
	case metric.GaugeType:
		val, _ := strconv.ParseFloat(m.ValueAsString(), 64)
		_, err := D.sql.Exec(CreateOrUpdateGauge(), m.Name(), m.Type(), val)
		if err != nil {
			return nil, err
		}

		return m, nil
	case metric.CounterType:
		delta, _ := strconv.ParseInt(m.ValueAsString(), 10, 64)

		var newDelta int
		err := D.sql.QueryRow(CreateOrUpdateCounter(), m.Name(), m.Type(), delta).Scan(&newDelta)
		if err != nil {
			return nil, err
		}
		return metric.NewMetric(m.Name(), m.Type(), fmt.Sprintf("%d", newDelta))
	default:
		return nil, fmt.Errorf("invalid metric type: %s", m.Type())
	}
}

func (D DBStorage) Ping(ctx context.Context) error {
	return D.sql.PingContext(ctx)
}
