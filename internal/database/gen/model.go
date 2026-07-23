// Package gen generates Go model code from database schema introspection.
package gen

import (
	"context"
	"database/sql"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	Name          string
	Type          string
	Nullable      bool
	Default       string
	Comment       string
	PrimaryKey    bool
	AutoIncrement bool
	Unique        bool
	UniqueGroup   string
	Unsigned      bool
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
	Relations     []Relation
}

// GenerateModels generates model files for the given tables.
func GenerateModels(ctx context.Context, db *database.DB, driver string, tables []string, opts GenerateOptions) error {
	if len(tables) == 0 {
		return nil
	}

	if err := validateRelations(opts.Relations); err != nil {
		return err
	}

	physicalRelations, err := getPhysicalRelations(ctx, db, driver, tables, opts.WithRelations)
	if err != nil {
		return fmt.Errorf("get physical relations: %w", err)
	}
	relations, err := mergeRelations(physicalRelations, opts.Relations)
	if err != nil {
		return err
	}

	tables = expandTablesForRelations(tables, relations)

	if err := validateRelationColumns(ctx, db, driver, relations); err != nil {
		return err
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for _, tableName := range tables {
		if err := generateModel(ctx, db, driver, tableName, opts, relations); err != nil {
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
		// Skip Bun migration bookkeeping tables.
		if tableName == "bun_migrations" || tableName == "bun_migration_locks" {
			continue
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

func generateModel(ctx context.Context, db *database.DB, driver, tableName string, opts GenerateOptions, relations []Relation) error {
	columns, err := GetTableColumns(ctx, db, driver, tableName)
	if err != nil {
		return err
	}

	modelName := ToCamelCase(tableName)
	code := GenerateModelCode(opts, modelName, tableName, columns, relations)

	formatted, err := format.Source([]byte(code))
	if err != nil {
		return fmt.Errorf("format generated code: %w", err)
	}

	filename := filepath.Join(opts.OutputDir, strings.ToLower(tableName)+".go")
	if err := os.WriteFile(filename, formatted, 0o644); err != nil {
		return err
	}

	return nil
}

func getMySQLColumns(ctx context.Context, db *database.DB, tableName string) ([]Column, error) {
	query := `
		SELECT c.column_name, c.data_type, c.column_type, c.is_nullable, c.column_default, c.column_comment, c.extra,
			COALESCE(p.is_primary, 0) AS is_primary_key
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name, 1 AS is_primary
			FROM information_schema.key_column_usage kcu
			JOIN information_schema.table_constraints tc
				ON kcu.table_schema = tc.table_schema
				AND kcu.table_name = tc.table_name
				AND kcu.constraint_name = tc.constraint_name
			WHERE kcu.table_schema = DATABASE() AND kcu.table_name = ? AND tc.constraint_type = 'PRIMARY KEY'
		) p ON c.column_name = p.column_name
		WHERE c.table_schema = DATABASE() AND c.table_name = ?
		ORDER BY c.ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := scanMySQLColumns(rows)
	if err != nil {
		return nil, err
	}

	groups, err := getMySQLUniqueGroups(ctx, db, tableName)
	if err != nil {
		return nil, err
	}
	applyUniqueConstraints(columns, groups)

	return columns, nil
}

func getPostgresColumns(ctx context.Context, db *database.DB, tableName string) ([]Column, error) {
	query := `
		SELECT c.column_name, c.data_type, c.is_nullable, c.column_default, '' as column_comment, c.is_identity,
			COALESCE(p.is_primary, 0) AS is_primary_key
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name, 1 AS is_primary
			FROM information_schema.key_column_usage kcu
			JOIN information_schema.table_constraints tc
				ON kcu.table_schema = tc.table_schema
				AND kcu.table_name = tc.table_name
				AND kcu.constraint_name = tc.constraint_name
			WHERE kcu.table_schema = 'public' AND kcu.table_name = $1 AND tc.constraint_type = 'PRIMARY KEY'
		) p ON c.column_name = p.column_name
		WHERE c.table_schema = 'public' AND c.table_name = $1
		ORDER BY c.ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := scanPostgresColumns(rows)
	if err != nil {
		return nil, err
	}

	groups, err := getPostgresUniqueGroups(ctx, db, tableName)
	if err != nil {
		return nil, err
	}
	applyUniqueConstraints(columns, groups)

	return columns, nil
}

func getSQLiteColumns(ctx context.Context, db *database.DB, tableName string) ([]Column, error) {
	query := fmt.Sprintf("PRAGMA table_info(%q)", tableName)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rawColumn struct {
		cid       int
		name      string
		dataType  string
		notNull   int
		dfltValue sql.NullString
		pk        int
	}

	var rawCols []rawColumn
	for rows.Next() {
		var r rawColumn
		if err := rows.Scan(&r.cid, &r.name, &r.dataType, &r.notNull, &r.dfltValue, &r.pk); err != nil {
			return nil, err
		}
		rawCols = append(rawCols, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	pkCount := 0
	for _, r := range rawCols {
		if r.pk > 0 {
			pkCount++
		}
	}

	var columns []Column
	for _, r := range rawCols {
		primaryKey := r.pk > 0
		autoIncrement := primaryKey && pkCount == 1 && strings.EqualFold(r.dataType, "integer")
		columns = append(columns, Column{
			Name:          r.name,
			Type:          r.dataType,
			Nullable:      r.notNull == 0 && !primaryKey,
			Default:       r.dfltValue.String,
			PrimaryKey:    primaryKey,
			AutoIncrement: autoIncrement,
		})
	}

	groups, err := getSQLiteUniqueGroups(ctx, db, tableName)
	if err != nil {
		return nil, err
	}
	applyUniqueConstraints(columns, groups)

	return columns, nil
}

func scanMySQLColumns(rows *sql.Rows) ([]Column, error) {
	var columns []Column
	for rows.Next() {
		var columnName, dataType, columnType, isNullable sql.NullString
		var columnDefault, columnComment, extra sql.NullString
		var isPrimaryKey int

		if err := rows.Scan(&columnName, &dataType, &columnType, &isNullable, &columnDefault, &columnComment, &extra, &isPrimaryKey); err != nil {
			return nil, err
		}

		columns = append(columns, Column{
			Name:          columnName.String,
			Type:          dataType.String,
			Nullable:      isNullable.String == "YES",
			Default:       columnDefault.String,
			Comment:       columnComment.String,
			PrimaryKey:    isPrimaryKey == 1,
			AutoIncrement: strings.Contains(strings.ToLower(extra.String), "auto_increment"),
			Unsigned:      strings.Contains(strings.ToLower(columnType.String), "unsigned"),
		})
	}

	return columns, rows.Err()
}

func scanPostgresColumns(rows *sql.Rows) ([]Column, error) {
	var columns []Column
	for rows.Next() {
		var columnName, dataType, isNullable, columnDefault, columnComment, isIdentity sql.NullString
		var isPrimaryKey int

		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &columnComment, &isIdentity, &isPrimaryKey); err != nil {
			return nil, err
		}

		autoIncrement := isIdentity.Valid && isIdentity.String == "YES"
		if !autoIncrement {
			autoIncrement = strings.Contains(columnDefault.String, "nextval")
		}

		columns = append(columns, Column{
			Name:          columnName.String,
			Type:          dataType.String,
			Nullable:      isNullable.String == "YES",
			Default:       columnDefault.String,
			Comment:       columnComment.String,
			PrimaryKey:    isPrimaryKey == 1,
			AutoIncrement: autoIncrement,
		})
	}

	return columns, rows.Err()
}

func scanUniqueGroups(rows *sql.Rows) (map[string][]string, error) {
	groups := make(map[string][]string)
	for rows.Next() {
		var constraintName, columnName string
		if err := rows.Scan(&constraintName, &columnName); err != nil {
			return nil, err
		}
		groups[constraintName] = append(groups[constraintName], columnName)
	}
	return groups, rows.Err()
}

func applyUniqueConstraints(columns []Column, groups map[string][]string) {
	if len(groups) == 0 {
		return
	}

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	// Prefer single-column unique constraints, then multi-column groups.
	sort.Slice(keys, func(i, j int) bool {
		li, lj := len(groups[keys[i]]), len(groups[keys[j]])
		if li != lj {
			return li < lj
		}
		return keys[i] < keys[j]
	})

	for _, group := range keys {
		cols := groups[group]
		for _, colName := range cols {
			for i := range columns {
				if columns[i].Name != colName || columns[i].Unique {
					continue
				}
				// Skip single-column unique on primary keys (redundant);
				// allow PK columns in composite unique groups.
				if columns[i].PrimaryKey && len(cols) == 1 {
					continue
				}
				columns[i].Unique = true
				if len(cols) > 1 {
					columns[i].UniqueGroup = group
				}
			}
		}
	}
}

func getMySQLUniqueGroups(ctx context.Context, db *database.DB, tableName string) (map[string][]string, error) {
	query := `
		SELECT tc.constraint_name, kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.table_schema = kcu.table_schema
			AND tc.table_name = kcu.table_name
			AND tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = DATABASE() AND tc.table_name = ? AND tc.constraint_type = 'UNIQUE'
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUniqueGroups(rows)
}

func getPostgresUniqueGroups(ctx context.Context, db *database.DB, tableName string) (map[string][]string, error) {
	query := `
		SELECT tc.constraint_name, kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.table_schema = kcu.table_schema
			AND tc.table_name = kcu.table_name
			AND tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = 'public' AND tc.table_name = $1 AND tc.constraint_type = 'UNIQUE'
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUniqueGroups(rows)
}

func getSQLiteUniqueGroups(ctx context.Context, db *database.DB, tableName string) (map[string][]string, error) {
	query := fmt.Sprintf("PRAGMA index_list(%q)", tableName)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type indexInfo struct {
		name   string
		unique int
		origin string
	}

	var indexes []indexInfo
	for rows.Next() {
		var seq, partial int
		var name, origin string
		var unique int
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, err
		}
		if unique == 1 && origin != "pk" {
			indexes = append(indexes, indexInfo{name: name, unique: unique, origin: origin})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	groups := make(map[string][]string)
	for _, idx := range indexes {
		colQuery := fmt.Sprintf("SELECT seqno, name FROM pragma_index_info('%s') ORDER BY seqno", strings.ReplaceAll(idx.name, "'", "''"))
		colRows, err := db.QueryContext(ctx, colQuery)
		if err != nil {
			return nil, err
		}
		for colRows.Next() {
			var seqno int
			var colName string
			if err := colRows.Scan(&seqno, &colName); err != nil {
				colRows.Close()
				return nil, err
			}
			groups[idx.name] = append(groups[idx.name], colName)
		}
		if err := colRows.Err(); err != nil {
			colRows.Close()
			return nil, err
		}
		colRows.Close()
	}

	return groups, nil
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

func getPhysicalRelations(ctx context.Context, db *database.DB, driver string, tables []string, enabled bool) ([]Relation, error) {
	if !enabled {
		return nil, nil
	}

	var relations []Relation
	for _, tableName := range tables {
		foreignKeys, err := getForeignKeys(ctx, db, driver, tableName)
		if err != nil {
			return nil, fmt.Errorf("get foreign keys for table %s: %w", tableName, err)
		}
		for _, foreignKey := range foreignKeys {
			relations = append(relations, Relation{
				SourceTable:  tableName,
				SourceColumn: foreignKey.ColumnName,
				TargetTable:  foreignKey.ReferencedTableName,
				TargetColumn: foreignKey.ReferencedColumnName,
			})
		}
	}
	return relations, nil
}

func validateRelationColumns(ctx context.Context, db *database.DB, driver string, relations []Relation) error {
	columnsByTable := make(map[string]map[string]struct{})
	fieldsByTable := make(map[string]map[string]struct{})
	loadColumns := func(tableName string) (map[string]struct{}, map[string]struct{}, error) {
		if columns, ok := columnsByTable[tableName]; ok {
			return columns, fieldsByTable[tableName], nil
		}
		columns, err := GetTableColumns(ctx, db, driver, tableName)
		if err != nil {
			return nil, nil, fmt.Errorf("read relation table %s: %w", tableName, err)
		}
		columnNames := make(map[string]struct{}, len(columns))
		fieldNames := make(map[string]struct{}, len(columns)+1)
		fieldNames["BaseModel"] = struct{}{}
		for _, column := range columns {
			columnNames[column.Name] = struct{}{}
			fieldNames[ToCamelCase(column.Name)] = struct{}{}
		}
		columnsByTable[tableName] = columnNames
		fieldsByTable[tableName] = fieldNames
		return columnNames, fieldNames, nil
	}

	for _, relation := range relations {
		sourceColumns, sourceFields, err := loadColumns(relation.SourceTable)
		if err != nil {
			return err
		}
		if _, ok := sourceColumns[relation.SourceColumn]; !ok {
			return fmt.Errorf("relation source column not found: %s.%s", relation.SourceTable, relation.SourceColumn)
		}
		field, reverseField := relationFieldNames(relation)
		if _, ok := sourceFields[field]; ok {
			return fmt.Errorf("relation field conflict with column: %s.%s", relation.SourceTable, field)
		}

		targetColumns, targetFields, err := loadColumns(relation.TargetTable)
		if err != nil {
			return err
		}
		if _, ok := targetColumns[relation.TargetColumn]; !ok {
			return fmt.Errorf("relation target column not found: %s.%s", relation.TargetTable, relation.TargetColumn)
		}
		if _, ok := targetFields[reverseField]; ok {
			return fmt.Errorf("relation field conflict with column: %s.%s", relation.TargetTable, reverseField)
		}
	}
	return nil
}

// expandTablesForRelations ensures that both ends of every relation are
// generated so that referenced model types are always available.
func expandTablesForRelations(tables []string, relations []Relation) []string {
	seen := make(map[string]struct{}, len(tables))
	for _, tableName := range tables {
		seen[tableName] = struct{}{}
	}
	for _, relation := range relations {
		if _, ok := seen[relation.SourceTable]; !ok {
			tables = append(tables, relation.SourceTable)
			seen[relation.SourceTable] = struct{}{}
		}
		if _, ok := seen[relation.TargetTable]; !ok {
			tables = append(tables, relation.TargetTable)
			seen[relation.TargetTable] = struct{}{}
		}
	}
	return tables
}
