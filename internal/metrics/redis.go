// Package metrics provides Prometheus metrics collection.
package metrics

// RecordRedisCacheHit records a Redis cache hit.
func (m *Metrics) RecordRedisCacheHit(keyPrefix string) {
	if m.registry == nil {
		return
	}
	m.RedisCacheHits.WithLabelValues(keyPrefix).Inc()
}

// RecordRedisCacheMiss records a Redis cache miss.
func (m *Metrics) RecordRedisCacheMiss(keyPrefix string) {
	if m.registry == nil {
		return
	}
	m.RedisCacheMisses.WithLabelValues(keyPrefix).Inc()
}

// RecordRedisOperation records a Redis operation duration.
func (m *Metrics) RecordRedisOperation(operation string, duration float64) {
	if m.registry == nil {
		return
	}
	m.RedisOperationDuration.WithLabelValues(operation).Observe(duration)
}
