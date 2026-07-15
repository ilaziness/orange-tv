package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
)

// LoginLogFilter filters login_logs queries.
type LoginLogFilter struct {
	UserType  *int8
	Username  string
	Status    *int8
	StartTime *time.Time
	EndTime   *time.Time
	Offset    int
	Limit     int
}

// SystemLogFilter filters system_logs queries.
type SystemLogFilter struct {
	Level   *int8
	Module  string
	AdminID *int64
	StartTime *time.Time
	EndTime   *time.Time
	Offset  int
	Limit   int
}

// LogRepository persists login and system logs.
type LogRepository interface {
	CreateLoginLog(ctx context.Context, m *model.LoginLogs) error
	ListLoginLogs(ctx context.Context, f LoginLogFilter) ([]model.LoginLogs, int, error)
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

func (r *logRepo) CreateLoginLog(ctx context.Context, m *model.LoginLogs) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create login log: %w", err)
	}
	return nil
}

func (r *logRepo) ListLoginLogs(ctx context.Context, f LoginLogFilter) ([]model.LoginLogs, int, error) {
	var items []model.LoginLogs
	q := r.db.NewSelect().Model(&items)
	if f.UserType != nil {
		q = q.Where("user_type = ?", *f.UserType)
	}
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
		return nil, 0, fmt.Errorf("count login logs: %w", err)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	err = q.Order("id DESC").Offset(f.Offset).Limit(limit).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list login logs: %w", err)
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
