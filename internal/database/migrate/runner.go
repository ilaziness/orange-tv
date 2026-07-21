package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
	bunmigrate "github.com/uptrace/bun/migrate"
)

var migrationNameRE = regexp.MustCompile(`^[0-9a-z_-]+$`)

// UpResult summarizes a migration up operation.
type UpResult struct {
	AppliedCount int
	PendingCount int
	PendingNames []string
	Group        string
}

// DownResult summarizes a migration down operation.
type DownResult struct {
	RollbackCount int
	RollbackNames []string
	Group         string
}

// StatusLine describes one migration's apply state.
type StatusLine struct {
	Name       string
	Applied    bool
	GroupID    int64
	MigratedAt time.Time
}

func newMigrator(ctx context.Context, db *database.DB, dir string) (*bunmigrate.Migrator, error) {
	migrations, err := LoadMigrations(dir)
	if err != nil {
		return nil, err
	}

	migrator := bunmigrate.NewMigrator(db.DB, migrations)
	if err := migrator.Init(ctx); err != nil {
		return nil, fmt.Errorf("initialize migrations: %w", err)
	}

	return migrator, nil
}

// Up runs pending migrations. When dryRun is true, migrations are not executed.
func Up(ctx context.Context, db *database.DB, dir string, dryRun bool) (*UpResult, error) {
	migrator, err := newMigrator(ctx, db, dir)
	if err != nil {
		return nil, err
	}

	list, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("get migrations: %w", err)
	}

	pending := pendingMigrations(list)
	if len(pending) == 0 {
		return &UpResult{}, nil
	}

	result := &UpResult{
		PendingCount: len(pending),
		PendingNames: migrationNames(pending),
	}
	if dryRun {
		return result, nil
	}

	if lockErr := migrator.Lock(ctx); lockErr != nil {
		return nil, fmt.Errorf("lock migrations: %w", lockErr)
	}
	defer func() {
		_ = migrator.Unlock(ctx)
	}()

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	if group != nil {
		result.AppliedCount = len(group.Migrations)
		result.Group = group.String()
	}

	return result, nil
}

// Down rolls back the last migration group. When dryRun is true, nothing is executed.
func Down(ctx context.Context, db *database.DB, dir string, dryRun bool) (*DownResult, error) {
	migrator, err := newMigrator(ctx, db, dir)
	if err != nil {
		return nil, err
	}

	list, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("get migrations: %w", err)
	}

	lastGroup := lastMigrationGroup(list)
	if len(lastGroup) == 0 {
		return &DownResult{}, nil
	}

	result := &DownResult{
		RollbackCount: len(lastGroup),
		RollbackNames: migrationNames(lastGroup),
	}
	if dryRun {
		return result, nil
	}

	if lockErr := migrator.Lock(ctx); lockErr != nil {
		return nil, fmt.Errorf("lock migrations: %w", lockErr)
	}
	defer func() {
		_ = migrator.Unlock(ctx)
	}()

	group, err := migrator.Rollback(ctx)
	if err != nil {
		return nil, fmt.Errorf("rollback migrations: %w", err)
	}

	if group != nil {
		result.RollbackCount = len(group.Migrations)
		result.RollbackNames = migrationNames(group.Migrations)
		result.Group = group.String()
	}

	return result, nil
}

// Status returns migration apply state for all discovered migrations.
func Status(ctx context.Context, db *database.DB, dir string) ([]StatusLine, error) {
	migrator, err := newMigrator(ctx, db, dir)
	if err != nil {
		return nil, err
	}

	list, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("get migration status: %w", err)
	}

	lines := make([]StatusLine, 0, len(list))
	for _, m := range list {
		line := StatusLine{Name: m.Name}
		if m.GroupID != 0 {
			line.Applied = true
			line.GroupID = m.GroupID
			line.MigratedAt = m.MigratedAt
		}
		lines = append(lines, line)
	}

	return lines, nil
}

// Create writes empty up and down SQL migration files.
func Create(dir, name string) (upFile, downFile string, err error) {
	if err := validateMigrationName(name); err != nil {
		return "", "", err
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, name)

	upFile = filepath.Join(dir, baseName+".up.sql")
	downFile = filepath.Join(dir, baseName+".down.sql")

	if err := ensurePathWithinDir(dir, upFile); err != nil {
		return "", "", err
	}
	if err := ensurePathWithinDir(dir, downFile); err != nil {
		return "", "", err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create migrations directory: %w", err)
	}

	if err := os.WriteFile(upFile, []byte("-- Up migration\n"), 0o644); err != nil {
		return "", "", fmt.Errorf("create up migration file: %w", err)
	}

	if err := os.WriteFile(downFile, []byte("-- Down migration\n"), 0o644); err != nil {
		return "", "", fmt.Errorf("create down migration file: %w", err)
	}

	return upFile, downFile, nil
}

func validateMigrationName(name string) error {
	if name == "" {
		return fmt.Errorf("migration name is required")
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("migration name must not contain path separators")
	}
	if !migrationNameRE.MatchString(name) {
		return fmt.Errorf("migration name must match [0-9a-z_-]+")
	}
	return nil
}

func ensurePathWithinDir(dir, filePath string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve migrations directory: %w", err)
	}

	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolve migration file path: %w", err)
	}

	rel, err := filepath.Rel(absDir, absFile)
	if err != nil {
		return fmt.Errorf("migration file must be inside migrations directory: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("migration file must be inside migrations directory")
	}

	return nil
}

func pendingMigrations(list bunmigrate.MigrationSlice) bunmigrate.MigrationSlice {
	var pending bunmigrate.MigrationSlice
	for _, m := range list {
		if m.GroupID == 0 {
			pending = append(pending, m)
		}
	}
	return pending
}

func lastMigrationGroup(list bunmigrate.MigrationSlice) bunmigrate.MigrationSlice {
	var maxGroupID int64
	for _, m := range list {
		if m.GroupID > maxGroupID {
			maxGroupID = m.GroupID
		}
	}
	if maxGroupID == 0 {
		return nil
	}

	var group bunmigrate.MigrationSlice
	for _, m := range list {
		if m.GroupID == maxGroupID {
			group = append(group, m)
		}
	}
	return group
}

func migrationNames(list bunmigrate.MigrationSlice) []string {
	names := make([]string, len(list))
	for i, m := range list {
		names[i] = m.Name
	}
	return names
}
