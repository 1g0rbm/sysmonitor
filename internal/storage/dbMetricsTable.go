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

func MetricsTable() string {
	return strings.Trim(metricsTable, " ")
}
