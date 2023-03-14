package storage

import (
	"context"
	"database/sql"
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

	c := func() error { return db.Close() }

	return DBStorage{
		sql: db,
	}, c, nil
}

func (D DBStorage) Set(m metric.IMetric) {
	//TODO implement me
}

func (D DBStorage) Get(name string) (metric.IMetric, error) {
	//TODO implement me
	return nil, nil
}

func (D DBStorage) All() map[string]metric.IMetric {
	//TODO implement me
	return map[string]metric.IMetric{}
}

func (D DBStorage) Update(m metric.IMetric) (metric.IMetric, error) {
	//TODO implement me
	return nil, nil
}

func (D DBStorage) GetCounter(name string) (metric.CounterMetric, error) {
	//TODO implement me
	return metric.CounterMetric{}, nil
}

func (D DBStorage) GetGauge(name string) (metric.GaugeMetric, error) {
	//TODO implement me
	return metric.GaugeMetric{}, nil
}

func (D DBStorage) Ping(ctx context.Context) error {
	return D.sql.PingContext(ctx)
}
