package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

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

func (s DBStorage) Get(name string) (metric.IMetric, error) {
	var (
		id    string
		mType string
		delta *int64
		val   *float64
	)

	err := s.sql.QueryRow(SelectMetric(), name).Scan(&id, &mType, &delta, &val)
	if delta == nil && val == nil {
		ErrMetricNotFound = fmt.Errorf("metric not found by name '%s'", name)
		return nil, ErrMetricNotFound
	}
	if err != nil {
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

func (s DBStorage) All() (map[string]metric.IMetric, error) {
	var (
		id    string
		mType string
		delta *int64
		val   *float64
	)

	r, err := s.sql.Query(SelectMetrics())
	if err != nil {
		return nil, err
	}

	defer func(r *sql.Rows) {
		if rErr := r.Close(); rErr != nil {
			err = rErr
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

	if err := r.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

func (s DBStorage) Update(m metric.IMetric) (metric.IMetric, error) {
	switch m.Type() {
	case metric.GaugeType:
		val, _ := strconv.ParseFloat(m.ValueAsString(), 64)
		_, err := s.sql.Exec(CreateOrUpdateGauge(), m.Name(), m.Type(), val)
		if err != nil {
			return nil, err
		}

		return m, nil
	case metric.CounterType:
		delta, _ := strconv.ParseInt(m.ValueAsString(), 10, 64)

		var newDelta int
		err := s.sql.QueryRow(CreateOrUpdateCounter(), m.Name(), m.Type(), delta).Scan(&newDelta)
		if err != nil {
			return nil, err
		}
		return metric.NewMetric(m.Name(), m.Type(), fmt.Sprintf("%d", newDelta))
	default:
		return nil, fmt.Errorf("invalid metric type: %s", m.Type())
	}
}

func (s DBStorage) BatchUpdate(sm []metric.IMetric) (err error) {
	tx, err := s.sql.Begin()
	if err != nil {
		return
	}

	defer func(tx *sql.Tx) {
		if err != nil {
			err = tx.Rollback()
		}
	}(tx)

	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Second)
	defer cancel()

	gStmt, err := tx.PrepareContext(ctx, CreateOrUpdateGauge())
	if err != nil {
		return
	}
	defer func(gStmt *sql.Stmt) {
		err = gStmt.Close()
	}(gStmt)

	cStmt, err := tx.PrepareContext(ctx, CreateOrUpdateCounter())
	if err != nil {
		return
	}
	defer func(cStmt *sql.Stmt) {
		err = cStmt.Close()
	}(cStmt)

	for _, m := range sm {
		switch m.Type() {
		case metric.GaugeType:
			val, _ := strconv.ParseFloat(m.ValueAsString(), 64)
			_, err = gStmt.ExecContext(ctx, m.Name(), m.Type(), val)
			if err != nil {
				return
			}
		case metric.CounterType:
			delta, _ := strconv.ParseInt(m.ValueAsString(), 10, 64)
			_, err = cStmt.ExecContext(ctx, m.Name(), m.Type(), delta)
		default:
			return fmt.Errorf("invalid metric type: %s", m.Type())
		}
	}

	err = tx.Commit()

	return
}

func (s DBStorage) Ping(ctx context.Context) error {
	return s.sql.PingContext(ctx)
}
