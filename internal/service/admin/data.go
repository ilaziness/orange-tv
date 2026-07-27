package admin

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

// DataService handles database backup and controlled batch data updates.
type DataService interface {
	Backup(ctx context.Context, w io.Writer, useNative bool) error
	BatchUpdatePreview(ctx context.Context, req *dto.BatchUpdatePreviewRequest) (int64, error)
	BatchUpdateExecute(ctx context.Context, req *dto.BatchUpdateExecuteRequest, adminID int64, ip string) (int64, error)
}

type dataService struct {
	db   *database.DB
	cfg  *config.Config
	repo repository.LogRepository
	log  *zap.Logger
}

// NewDataService creates a DataService.
func NewDataService(db *database.DB, cfg *config.Config, repo repository.LogRepository, log *zap.Logger) DataService {
	if log == nil {
		log = zap.NewNop()
	}
	return &dataService{db: db, cfg: cfg, repo: repo, log: log}
}

func (s *dataService) backupWithNativeTool(ctx context.Context, w io.Writer) error {
	driver := strings.ToLower(s.cfg.Database.Driver)
	switch driver {
	case constant.DriverMySQL:
		return s.backupWithMysqldump(ctx, w)
	case constant.DriverPostgres, constant.DriverPostgreSQL:
		return s.backupWithPgDump(ctx, w)
	default:
		return errcode.WithMessage(errcode.ServiceUnavailable, "当前数据库类型不支持原生工具导出")
	}
}

func (s *dataService) backupWithMysqldump(ctx context.Context, w io.Writer) error {
	cfg := s.cfg.Database
	// Use a temporary option file to avoid exposing the password on the command line.
	optsFile := filepath.Join(os.TempDir(), fmt.Sprintf("orange-tv-mysqldump-%d.cnf", time.Now().UnixNano()))
	content := fmt.Sprintf("[mysqldump]\nhost=%s\nport=%d\nuser=%s\npassword=%s\n", cfg.Host, cfg.Port, cfg.User, cfg.Password)
	if err := os.WriteFile(optsFile, []byte(content), 0o600); err != nil {
		return errcode.Wrap(errcode.ServiceUnavailable, fmt.Errorf("write mysqldump option file: %w", err))
	}
	defer func() {
		if err := os.Remove(optsFile); err != nil {
			s.log.Warn("data: failed to remove mysqldump option file", zap.String("file", optsFile), zap.Error(err))
		}
	}()

	// #nosec G204 -- command arguments come from server-side config, not user input.
	cmd := exec.CommandContext(ctx, "mysqldump",
		"--defaults-extra-file="+optsFile,
		"--single-transaction",
		"--routines",
		"--triggers",
		"--events",
		"--hex-blob",
		cfg.Database,
	)
	return s.runNativeCommand(ctx, cmd, w, "mysqldump")
}

func (s *dataService) backupWithPgDump(ctx context.Context, w io.Writer) error {
	cfg := s.cfg.Database
	// Keep the password out of the command line by passing it via PGPASSWORD.
	dsn := fmt.Sprintf("postgresql://%s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.Database)
	// #nosec G204 -- DSN is built from server-side config, not user input.
	cmd := exec.CommandContext(ctx, "pg_dump",
		"--dbname", dsn,
		"--clean",
		"--if-exists",
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
	return s.runNativeCommand(ctx, cmd, w, "pg_dump")
}

func (s *dataService) runNativeCommand(ctx context.Context, cmd *exec.Cmd, w io.Writer, tool string) error {
	cmd.Stdout = w
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errcode.Wrap(errcode.ServiceUnavailable, fmt.Errorf("setup %s stderr: %w", tool, err))
	}
	if err := cmd.Start(); err != nil {
		return errcode.WithMessage(errcode.ServiceUnavailable, fmt.Sprintf("启动 %s 失败，请确认已安装并加入 PATH", tool))
	}
	stderrBytes, _ := io.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return errcode.WithMessage(errcode.ServiceUnavailable, "导出操作已取消")
		}
		msg := strings.TrimSpace(string(stderrBytes))
		if msg == "" {
			msg = err.Error()
		}
		return errcode.WithMessage(errcode.ServiceUnavailable, fmt.Sprintf("%s 导出失败: %s", tool, msg))
	}
	return nil
}

// allowedBatchUpdateTargets defines the only tables/fields that can be batch updated.
var allowedBatchUpdateTargets = map[string]struct {
	Table string
	Field string
	Label string
}{
	dto.TargetVideoCover: {Table: "videos", Field: "cover_image", Label: "影视封面"},
	dto.TargetEpisodeURL: {Table: "play_episodes", Field: "play_url", Label: "播放链接"},
}

// Backup streams a SQL dump of the configured database to w.
// When useNative is true, it delegates to mysqldump (MySQL) or pg_dump (PostgreSQL)
// and produces a full dump including schema and data.
// Otherwise it generates INSERT statements directly via SQL queries
// without relying on external command-line tools.
func (s *dataService) Backup(ctx context.Context, w io.Writer, useNative bool) error {
	if s.cfg == nil || s.cfg.Database.Driver == "" {
		return errcode.WithMessage(errcode.ServiceUnavailable, "数据库配置不可用")
	}

	if useNative {
		return s.backupWithNativeTool(ctx, w)
	}

	driver := strings.ToLower(s.cfg.Database.Driver)
	var dumper dumper
	switch driver {
	case constant.DriverMySQL:
		dumper = &mysqlDumper{db: s.db, cfg: s.cfg.Database}
	case constant.DriverPostgres, constant.DriverPostgreSQL:
		dumper = &pgDumper{db: s.db, cfg: s.cfg.Database}
	default:
		return errcode.WithMessage(errcode.ServiceUnavailable, "当前数据库类型不支持导出")
	}

	if err := dumper.dumpHeader(ctx, w); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}

	tables, err := dumper.listTables(ctx)
	if err != nil {
		s.log.Error("data: list tables for backup failed", zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}

	for _, table := range tables {
		if err := dumper.dumpTable(ctx, w, table); err != nil {
			s.log.Error("data: dump table failed", zap.String("table", table), zap.Error(err))
			return errcode.Wrap(errcode.DatabaseError, fmt.Errorf("dump table %s: %w", table, err))
		}
	}
	if err := dumper.dumpFooter(ctx, w); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// BatchUpdatePreview returns the number of rows that would be affected.
func (s *dataService) BatchUpdatePreview(ctx context.Context, req *dto.BatchUpdatePreviewRequest) (int64, error) {
	target, err := s.resolveTarget(req.Target)
	if err != nil {
		return 0, err
	}
	oldValue := strings.TrimSpace(req.OldValue)
	if oldValue == "" {
		return 0, errcode.WithMessage(errcode.ParamError, "查找字符串不能为空")
	}

	count, err := s.db.NewSelect().
		Table(target.Table).
		Where("? LIKE ?", bun.Ident(target.Field), "%"+oldValue+"%").
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		s.log.Error("data: batch update preview failed", zap.String("table", target.Table), zap.String("field", target.Field), zap.Error(err))
		return 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return int64(count), nil
}

// BatchUpdateExecute performs the actual string replacement on the allowed field.
func (s *dataService) BatchUpdateExecute(ctx context.Context, req *dto.BatchUpdateExecuteRequest, adminID int64, ip string) (int64, error) {
	target, err := s.resolveTarget(req.Target)
	if err != nil {
		return 0, err
	}
	oldValue := strings.TrimSpace(req.OldValue)
	newValue := strings.TrimSpace(req.NewValue)
	if oldValue == "" {
		return 0, errcode.WithMessage(errcode.ParamError, "查找字符串不能为空")
	}
	if newValue == "" {
		return 0, errcode.WithMessage(errcode.ParamError, "替换字符串不能为空")
	}
	if oldValue == newValue {
		return 0, errcode.WithMessage(errcode.ParamError, "查找字符串与替换字符串不能相同")
	}

	res, err := s.db.NewUpdate().
		Table(target.Table).
		Set("? = REPLACE(?, ?, ?), updated_at = NOW()",
			bun.Ident(target.Field), bun.Ident(target.Field), oldValue, newValue).
		Where("? LIKE ?", bun.Ident(target.Field), "%"+oldValue+"%").
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		s.log.Error("data: batch update execute failed", zap.String("table", target.Table), zap.String("field", target.Field), zap.Error(err))
		return 0, errcode.Wrap(errcode.DatabaseError, err)
	}

	affected, _ := res.RowsAffected()
	if err := s.recordSystemLog(ctx, adminID, ip, target.Label, oldValue, newValue, affected); err != nil {
		s.log.Error("data: record batch update system log failed", zap.Error(err))
	}
	return affected, nil
}

func (s *dataService) resolveTarget(target string) (struct {
	Table string
	Field string
	Label string
}, error) {
	if t, ok := allowedBatchUpdateTargets[target]; ok {
		return t, nil
	}
	return struct {
		Table string
		Field string
		Label string
	}{}, errcode.WithMessage(errcode.ParamError, "不支持的目标字段")
}

func (s *dataService) recordSystemLog(ctx context.Context, adminID int64, ip, target, oldValue, newValue string, affected int64) error {
	if s.repo == nil {
		return nil
	}
	content, err := json.Marshal(map[string]any{
		"target":        target,
		"old_value":     oldValue,
		"new_value":     newValue,
		"affected_rows": affected,
		"ip_address":    ip,
	})
	if err != nil {
		return fmt.Errorf("marshal system log content: %w", err)
	}
	contentStr := string(content)
	now := time.Now()
	if adminID < 0 {
		adminID = 0
	}
	return s.repo.CreateSystemLog(ctx, &model.SystemLogs{
		Level:     1,
		Module:    "data_admin",
		Action:    "batch_update",
		AdminID:   uint64(adminID),
		Content:   &contentStr,
		IPAddress: ip,
		CreatedAt: &now,
	})
}

// BackupFilename returns a suggested filename for the backup download.
func BackupFilename() string {
	return fmt.Sprintf("orange-tv-backup-%s.sql", time.Now().Format("20060102-150405"))
}

// dumper defines database-specific dump logic.
type dumper interface {
	listTables(ctx context.Context) ([]string, error)
	dumpHeader(ctx context.Context, w io.Writer) error
	dumpTable(ctx context.Context, w io.Writer, table string) error
	dumpFooter(ctx context.Context, w io.Writer) error
}

type columnInfo struct {
	name     string
	isBinary bool
}

type mysqlDumper struct {
	db  *database.DB
	cfg config.DatabaseConfig
}

type pgDumper struct {
	db  *database.DB
	cfg config.DatabaseConfig
}

func (d *mysqlDumper) listTables(ctx context.Context) ([]string, error) {
	var tables []string
	err := d.db.NewSelect().
		ColumnExpr("table_name").
		Table("information_schema.tables").
		Where("table_schema = ? AND table_type = 'BASE TABLE'", d.cfg.Database).
		Order("table_name").
		Scan(ctx, &tables)
	return tables, err
}

func (d *mysqlDumper) dumpHeader(ctx context.Context, w io.Writer) error {
	_, err := fmt.Fprintf(w, "-- Orange TV database backup\n-- Driver: mysql\n-- Database: %s\n-- Generated at: %s\n\nSET NAMES utf8mb4;\nSET FOREIGN_KEY_CHECKS = 0;\n\n", d.cfg.Database, time.Now().Format(time.RFC3339))
	return err
}

func (d *mysqlDumper) dumpFooter(ctx context.Context, w io.Writer) error {
	_, err := fmt.Fprintln(w, "\nSET FOREIGN_KEY_CHECKS = 1;")
	return err
}

func (d *mysqlDumper) dumpTable(ctx context.Context, w io.Writer, table string) error {
	cols, err := d.listColumns(ctx, table)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return nil
	}

	quotedTable := utils.QuoteMySQL(table)
	colNames := make([]string, len(cols))
	for i, c := range cols {
		colNames[i] = utils.QuoteMySQL(c.name)
	}

	batchSize := 500
	offset := 0
	for {
		rows, err := d.db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM %s LIMIT %d OFFSET %d", strings.Join(colNames, ", "), quotedTable, batchSize, offset))
		if err != nil {
			return fmt.Errorf("select data: %w", err)
		}
		count, err := d.writeInsertRows(w, quotedTable, colNames, cols, rows)
		_ = rows.Close()
		if err != nil {
			return err
		}
		if count == 0 {
			break
		}
		offset += batchSize
	}
	return nil
}

func (d *mysqlDumper) listColumns(ctx context.Context, table string) ([]columnInfo, error) {
	var raw []struct {
		Name     string `bun:"COLUMN_NAME"`
		DataType string `bun:"DATA_TYPE"`
	}
	err := d.db.NewSelect().
		ColumnExpr("COLUMN_NAME, DATA_TYPE").
		Table("information_schema.columns").
		Where("table_schema = ? AND table_name = ?", d.cfg.Database, table).
		Order("ORDINAL_POSITION").
		Scan(ctx, &raw)
	if err != nil {
		return nil, err
	}
	cols := make([]columnInfo, len(raw))
	for i, r := range raw {
		cols[i] = columnInfo{
			name:     r.Name,
			isBinary: utils.IsMySQLBinaryType(r.DataType),
		}
	}
	return cols, nil
}

func (d *mysqlDumper) writeInsertRows(w io.Writer, table string, colNames []string, cols []columnInfo, rows *sql.Rows) (int, error) {
	values := make([]any, len(cols))
	valuePtrs := make([]any, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	count := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return 0, fmt.Errorf("scan row: %w", err)
		}
		if count == 0 {
			if _, err := fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES\n", table, strings.Join(colNames, ", ")); err != nil {
				return 0, err
			}
		} else {
			if _, err := fmt.Fprint(w, ",\n"); err != nil {
				return 0, err
			}
		}
		rowVals := make([]string, len(cols))
		for i, v := range values {
			rowVals[i] = formatMySQLValue(v, cols[i])
		}
		if _, err := fmt.Fprintf(w, "(%s)", strings.Join(rowVals, ", ")); err != nil {
			return 0, err
		}
		count++
	}
	if count > 0 {
		if _, err := fmt.Fprintln(w, ";"); err != nil {
			return 0, err
		}
	}
	return count, rows.Err()
}

func formatMySQLValue(v any, col columnInfo) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		if col.isBinary {
			return "0x" + hex.EncodeToString(val)
		}
		return "'" + utils.EscapeMySQLString(string(val)) + "'"
	case string:
		return "'" + utils.EscapeMySQLString(val) + "'"
	case bool:
		if val {
			return "1"
		}
		return "0"
	case time.Time:
		return "'" + val.Format("2006-01-02 15:04:05.000000") + "'"
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int:
		return strconv.FormatInt(int64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'g', -1, 32)
	default:
		return "'" + utils.EscapeMySQLString(fmt.Sprint(val)) + "'"
	}
}

func (d *pgDumper) listTables(ctx context.Context) ([]string, error) {
	var tables []string
	err := d.db.NewSelect().
		ColumnExpr("table_name").
		Table("information_schema.tables").
		Where("table_schema = 'public' AND table_type = 'BASE TABLE'").
		Order("table_name").
		Scan(ctx, &tables)
	return tables, err
}

func (d *pgDumper) dumpHeader(ctx context.Context, w io.Writer) error {
	_, err := fmt.Fprintf(w, "-- Orange TV database backup\n-- Driver: postgres\n-- Database: %s\n-- Generated at: %s\n\n", d.cfg.Database, time.Now().Format(time.RFC3339))
	return err
}

func (d *pgDumper) dumpFooter(ctx context.Context, w io.Writer) error {
	return nil
}

func (d *pgDumper) dumpTable(ctx context.Context, w io.Writer, table string) error {
	cols, err := d.listColumns(ctx, table)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return nil
	}

	quotedTable := utils.QuotePG(table)
	colNames := make([]string, len(cols))
	for i, c := range cols {
		colNames[i] = utils.QuotePG(c.name)
	}

	batchSize := 500
	offset := 0
	for {
		rows, err := d.db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM %s LIMIT %d OFFSET %d", strings.Join(colNames, ", "), quotedTable, batchSize, offset))
		if err != nil {
			return fmt.Errorf("select data: %w", err)
		}
		count, err := d.writeInsertRows(w, quotedTable, colNames, cols, rows)
		_ = rows.Close()
		if err != nil {
			return err
		}
		if count == 0 {
			break
		}
		offset += batchSize
	}
	return nil
}

func (d *pgDumper) listColumns(ctx context.Context, table string) ([]columnInfo, error) {
	var raw []struct {
		Name     string `bun:"COLUMN_NAME"`
		BaseType string `bun:"BASE_TYPE"`
	}
	err := d.db.NewSelect().
		ColumnExpr("a.attname AS COLUMN_NAME, a.atttypid::regtype AS BASE_TYPE").
		TableExpr("pg_catalog.pg_attribute AS a").
		Join("JOIN pg_catalog.pg_class AS c ON c.oid = a.attrelid").
		Join("JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace").
		Where("c.relname = ? AND n.nspname = 'public' AND a.attnum > 0 AND NOT a.attisdropped", table).
		OrderExpr("a.attnum").
		Scan(ctx, &raw)
	if err != nil {
		return nil, err
	}
	cols := make([]columnInfo, len(raw))
	for i, r := range raw {
		cols[i] = columnInfo{
			name:     r.Name,
			isBinary: strings.EqualFold(r.BaseType, "bytea"),
		}
	}
	return cols, nil
}

func (d *pgDumper) writeInsertRows(w io.Writer, table string, colNames []string, cols []columnInfo, rows *sql.Rows) (int, error) {
	values := make([]any, len(cols))
	valuePtrs := make([]any, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	count := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return 0, fmt.Errorf("scan row: %w", err)
		}
		if count == 0 {
			if _, err := fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES\n", table, strings.Join(colNames, ", ")); err != nil {
				return 0, err
			}
		} else {
			if _, err := fmt.Fprint(w, ",\n"); err != nil {
				return 0, err
			}
		}
		rowVals := make([]string, len(cols))
		for i, v := range values {
			rowVals[i] = formatPGValue(v, cols[i])
		}
		if _, err := fmt.Fprintf(w, "(%s)", strings.Join(rowVals, ", ")); err != nil {
			return 0, err
		}
		count++
	}
	if count > 0 {
		if _, err := fmt.Fprintln(w, ";"); err != nil {
			return 0, err
		}
	}
	return count, rows.Err()
}

func formatPGValue(v any, col columnInfo) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		if col.isBinary {
			return "'\\x" + hex.EncodeToString(val) + "'::bytea"
		}
		return "'" + utils.EscapePGString(string(val)) + "'"
	case string:
		return "'" + utils.EscapePGString(val) + "'"
	case bool:
		return strconv.FormatBool(val)
	case time.Time:
		return "'" + val.UTC().Format("2006-01-02 15:04:05.999999+00") + "'"
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int:
		return strconv.FormatInt(int64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'g', -1, 32)
	default:
		return "'" + utils.EscapePGString(fmt.Sprint(val)) + "'"
	}
}
