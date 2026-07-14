// Package gen generates Go model code from database schema introspection.
package gen

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var tableNameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateTableName(name string) error {
	if name == "" {
		return fmt.Errorf("table name is required")
	}
	if !tableNameRE.MatchString(name) {
		return fmt.Errorf("invalid table name %q", name)
	}
	return nil
}

// ValidateTableName reports whether a table name is safe for introspection.
func ValidateTableName(name string) error {
	return validateTableName(name)
}

var titleCaser = cases.Title(language.Und)

// Column describes one database column.
type Column struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
	Comment  string
}

// ForeignKey describes a foreign key column reference.
type ForeignKey struct {
	ColumnName           string
	ReferencedTableName  string
	ReferencedColumnName string
}

// GenerateOptions configures model code generation.
type GenerateOptions struct {
	OutputDir     string
	PackageName   string
	JSONTags      bool
	ValidatorTags bool
	WithRelations bool
}

// GenerateModels generates model files for the given tables.
func GenerateModels(ctx context.Context, db *database.DB, driver string, tables []string, opts GenerateOptions) error {
	if len(tables) == 0 {
		return nil
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for _, tableName := range tables {
		if err := generateModel(ctx, db, driver, tableName, opts); err != nil {
			return fmt.Errorf("generate model for table %s: %w", tableName, err)
		}
	}

	return nil
}

// GetTables returns user table names for the configured driver.
func GetTables(ctx context.Context, db *database.DB, driver string) ([]string, error) {
	var query string

	switch driver {
	case constant.DriverMySQL:
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'"
	case constant.DriverPostgres, constant.DriverPostgreSQL:
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'"
	case constant.DriverSQLite, constant.DriverSQLite3:
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// GetTableColumns returns column metadata for a table.
func GetTableColumns(ctx context.Context, db *database.DB, driver, tableName string) ([]Column, error) {
	if err := validateTableName(tableName); err != nil {
		return nil, err
	}

	switch driver {
	case constant.DriverMySQL:
		return getMySQLColumns(ctx, db, tableName)
	case constant.DriverPostgres, constant.DriverPostgreSQL:
		return getPostgresColumns(ctx, db, tableName)
	case constant.DriverSQLite, constant.DriverSQLite3:
		return getSQLiteColumns(ctx, db, tableName)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}

func generateModel(ctx context.Context, db *database.DB, driver, tableName string, opts GenerateOptions) error {
	columns, err := GetTableColumns(ctx, db, driver, tableName)
	if err != nil {
		return err
	}

	var foreignKeys []ForeignKey
	if opts.WithRelations {
		foreignKeys, err = getForeignKeys(ctx, db, driver, tableName)
		if err != nil {
			return fmt.Errorf("get foreign keys: %w", err)
		}
	}

	modelName := ToCamelCase(tableName)
	code := GenerateModelCode(opts, modelName, tableName, columns, foreignKeys)

	filename := filepath.Join(opts.OutputDir, strings.ToLower(tableName)+".go")
	if err := os.WriteFile(filename, []byte(code), 0o644); err != nil {
		return err
	}

	return nil
}

func getMySQLColumns(ctx context.Context, db *database.DB, tableName string) ([]Column, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default, column_comment
		FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ?
		ORDER BY ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStandardColumns(rows)
}

func getPostgresColumns(ctx context.Context, db *database.DB, tableName string) ([]Column, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default, '' as column_comment
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStandardColumns(rows)
}

func getSQLiteColumns(ctx context.Context, db *database.DB, tableName string) ([]Column, error) {
	query := fmt.Sprintf("PRAGMA table_info(%q)", tableName)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}

		columns = append(columns, Column{
			Name:     name,
			Type:     dataType,
			Nullable: notNull == 0,
			Default:  dfltValue.String,
		})
	}

	return columns, rows.Err()
}

func scanStandardColumns(rows *sql.Rows) ([]Column, error) {
	var columns []Column
	for rows.Next() {
		var columnName, dataType, isNullable sql.NullString
		var columnDefault, columnComment sql.NullString

		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &columnComment); err != nil {
			return nil, err
		}

		columns = append(columns, Column{
			Name:     columnName.String,
			Type:     dataType.String,
			Nullable: isNullable.String == "YES",
			Default:  columnDefault.String,
			Comment:  columnComment.String,
		})
	}

	return columns, rows.Err()
}

func getForeignKeys(ctx context.Context, db *database.DB, driver, tableName string) ([]ForeignKey, error) {
	switch driver {
	case constant.DriverMySQL:
		return getMySQLForeignKeys(ctx, db, tableName)
	case constant.DriverPostgres, constant.DriverPostgreSQL:
		return getPostgresForeignKeys(ctx, db, tableName)
	case constant.DriverSQLite, constant.DriverSQLite3:
		return getSQLiteForeignKeys(ctx, db, tableName)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}

func getMySQLForeignKeys(ctx context.Context, db *database.DB, tableName string) ([]ForeignKey, error) {
	query := `
		SELECT kcu.column_name, kcu.referenced_table_name, kcu.referenced_column_name
		FROM information_schema.key_column_usage kcu
		WHERE kcu.table_schema = DATABASE() AND kcu.table_name = ?
			AND kcu.referenced_table_name IS NOT NULL
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanForeignKeys(rows)
}

func getPostgresForeignKeys(ctx context.Context, db *database.DB, tableName string) ([]ForeignKey, error) {
	query := `
		SELECT kcu.column_name, ccu.table_name, ccu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu ON kcu.constraint_name = ccu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = 'public' AND tc.table_name = $1
			AND ccu.table_schema = 'public'
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanForeignKeys(rows)
}

func getSQLiteForeignKeys(ctx context.Context, db *database.DB, tableName string) ([]ForeignKey, error) {
	query := fmt.Sprintf("PRAGMA foreign_key_list(%q)", tableName)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []ForeignKey
	for rows.Next() {
		var id, seq int
		var table, from, to, onUpdate, onDelete sql.NullString
		var match sql.NullString

		if err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}

		if table.Valid {
			keys = append(keys, ForeignKey{
				ColumnName:           from.String,
				ReferencedTableName:  table.String,
				ReferencedColumnName: to.String,
			})
		}
	}

	return keys, rows.Err()
}

func scanForeignKeys(rows *sql.Rows) ([]ForeignKey, error) {
	var keys []ForeignKey
	for rows.Next() {
		var columnName, referencedTableName, referencedColumnName sql.NullString

		if err := rows.Scan(&columnName, &referencedTableName, &referencedColumnName); err != nil {
			return nil, err
		}

		if referencedColumnName.Valid {
			keys = append(keys, ForeignKey{
				ColumnName:           columnName.String,
				ReferencedTableName:  referencedTableName.String,
				ReferencedColumnName: referencedColumnName.String,
			})
		}
	}

	return keys, rows.Err()
}
