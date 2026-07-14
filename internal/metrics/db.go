// Package metrics provides Prometheus metrics collection.
package metrics

// UpdateDBConnectionStats updates database connection statistics.
func (m *Metrics) UpdateDBConnectionStats(active, idle int) {
	if m.registry == nil {
		return
	}
	m.DBConnectionsActive.Set(float64(active))
	m.DBConnectionsIdle.Set(float64(idle))
}

// RecordDBQuery records a database query duration.
func (m *Metrics) RecordDBQuery(operation, table string, duration float64) {
	if m.registry == nil {
		return
	}
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}
