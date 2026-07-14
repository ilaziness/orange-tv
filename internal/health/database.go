package health

import (
	"context"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
)

const defaultDatabaseCheckTimeout = 5 * time.Second

// DatabaseChecker checks database connectivity via PingContext.
type DatabaseChecker struct {
	db      *database.DB
	timeout time.Duration
}

// NewDatabaseChecker creates a database readiness checker.
func NewDatabaseChecker(db *database.DB) Checker {
	return &DatabaseChecker{
		db:      db,
		timeout: defaultDatabaseCheckTimeout,
	}
}

func (c *DatabaseChecker) Name() string { return "database" }

func (c *DatabaseChecker) Check(ctx context.Context) error {
	timeout := c.timeout
	if timeout <= 0 {
		timeout = defaultDatabaseCheckTimeout
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.db.HealthCheck(checkCtx)
}
