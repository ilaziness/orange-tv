// Package migrate provides SQL migration discovery and execution using Bun.
package migrate

import (
	"fmt"
	"os"

	bunmigrate "github.com/uptrace/bun/migrate"
)

// LoadMigrations discovers SQL migration files in the given directory.
func LoadMigrations(dir string) (*bunmigrate.Migrations, error) {
	m := bunmigrate.NewMigrations(bunmigrate.WithMigrationsDirectory(dir))
	if err := m.Discover(os.DirFS(dir)); err != nil {
		return nil, fmt.Errorf("discover migrations: %w", err)
	}
	return m, nil
}
