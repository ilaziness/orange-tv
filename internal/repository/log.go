package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
)

// AdminLoginLogFilter filters admin_login_logs queries.
type AdminLoginLogFilter struct {
	Username  string
	Status    *uint8
	StartTime *time.Time
	EndTime   *time.Time
	Offset    int
	Limit     int
}

// SystemLogFilter filters system_logs queries.
type SystemLogFilter struct {
	Level     *uint8
	Module    string
	AdminID   *uint64
	StartTime *time.Time
	EndTime   *time.Time
	Offset    int
	Limit     int
}

// LogRepository persists admin login and system logs.
type LogRepository interface {
	CreateAdminLoginLog(ctx context.Context, m *model.AdminLoginLogs) error
	ListAdminLoginLogs(ctx context.Context, f AdminLoginLogFilter) ([]model.AdminLoginLogs, int, error)
	CreateSystemLog(ctx context.Context, m *model.SystemLogs) error
	ListSystemLogs(ctx context.Context, f SystemLogFilter) ([]model.SystemLogs, int, error)
}

type logRepo struct {
	db *database.DB
}

// NewLogRepo creates a LogRepository.
func NewLogRepo(db *database.DB) LogRepository {
	return &logRepo{db: db}
}

func (r *logRepo) CreateAdminLoginLog(ctx context.Context, m *model.AdminLoginLogs) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create admin login log: %w", err)
	}
	return nil
}

func (r *logRepo) ListAdminLoginLogs(ctx context.Context, f AdminLoginLogFilter) ([]model.AdminLoginLogs, int, error) {
	var items []model.AdminLoginLogs
	q := r.db.NewSelect().Model(&items)
	if kw := strings.TrimSpace(f.Username); kw != "" {
		q = q.Where("username LIKE ?", "%"+kw+"%")
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", *f.EndTime)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count admin login logs: %w", err)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	err = q.Order("id DESC").Offset(f.Offset).Limit(limit).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin login logs: %w", err)
	}
	return items, total, nil
}

func (r *logRepo) CreateSystemLog(ctx context.Context, m *model.SystemLogs) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create system log: %w", err)
	}
	return nil
}

func (r *logRepo) ListSystemLogs(ctx context.Context, f SystemLogFilter) ([]model.SystemLogs, int, error) {
	var items []model.SystemLogs
	q := r.db.NewSelect().Model(&items)
	if f.Level != nil {
		q = q.Where("level = ?", *f.Level)
	}
	if mod := strings.TrimSpace(f.Module); mod != "" {
		q = q.Where("module = ?", mod)
	}
	if f.AdminID != nil && *f.AdminID > 0 {
		q = q.Where("admin_id = ?", *f.AdminID)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", *f.EndTime)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count system logs: %w", err)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	err = q.Order("id DESC").Offset(f.Offset).Limit(limit).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list system logs: %w", err)
	}
	return items, total, nil
}
