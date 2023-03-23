package storage

import "strings"

const metricsTable = `
CREATE TABLE IF NOT EXISTS metrics (
  id    VARCHAR(255),
  m_type VARCHAR(255),
  delta	BIGINT,
  val   DOUBLE PRECISION,
  PRIMARY KEY (id, m_type)
);
`

const createOrUpdateGauge = `
INSERT INTO metrics
	(id, m_type, val)
	VALUES ($1, $2, $3)
	ON CONFLICT (id,m_type) 
	DO UPDATE SET val=$3;
`

const createOrUpdateCounter = `
INSERT INTO metrics
	(id, m_type, delta)
	VALUES ($1, $2, $3)
	ON CONFLICT (id,m_type)
	DO UPDATE SET delta=(metrics.delta + ($3))
	RETURNING delta;
`

const selectMetric = `
SELECT id,m_type,delta,val
FROM metrics
WHERE id = $1
`

const selectMetrics = `
SELECT id,m_type,delta,val
FROM metrics
LIMIT $1 OFFSET $2
`

func MetricsTable() string {
	return strings.Trim(metricsTable, " ")
}

func CreateOrUpdateGauge() string {
	return strings.Trim(createOrUpdateGauge, " ")
}

func CreateOrUpdateCounter() string {
	return strings.Trim(createOrUpdateCounter, " ")
}

func SelectMetric() string {
	return strings.Trim(selectMetric, " ")
}

func SelectMetrics() string {
	return strings.Trim(selectMetrics, " ")
}
