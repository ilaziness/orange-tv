package admin

import (
	"context"
	"time"

	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// LogService queries login / system logs for admin.
type LogService interface {
	ListSystemLogs(ctx context.Context, req *admindto.SystemLogListRequest) ([]admindto.SystemLogItem, int, error)
	ListLoginLogs(ctx context.Context, req *admindto.LoginLogListRequest) ([]admindto.LoginLogItem, int, error)
}

type logService struct {
	repo repository.LogRepository
}

// NewLogService creates a LogService.
func NewLogService(repo repository.LogRepository) LogService {
	return &logService{repo: repo}
}

func (s *logService) ListSystemLogs(ctx context.Context, req *admindto.SystemLogListRequest) ([]admindto.SystemLogItem, int, error) {
	if req == nil {
		req = &admindto.SystemLogListRequest{}
	}
	start, end, err := parseTimeRange(req.Start, req.End)
	if err != nil {
		return nil, 0, errcode.WithMessage(errcode.ParamError, "时间范围无效")
	}
	items, total, err := s.repo.ListSystemLogs(ctx, repository.SystemLogFilter{
		Level:     req.Level,
		Module:    req.Module,
		AdminID:   req.AdminID,
		StartTime: start,
		EndTime:   end,
		Offset:    req.GetOffset(),
		Limit:     req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.SystemLogItem, 0, len(items))
	for i := range items {
		out = append(out, toSystemLogItem(&items[i]))
	}
	return out, total, nil
}

func (s *logService) ListLoginLogs(ctx context.Context, req *admindto.LoginLogListRequest) ([]admindto.LoginLogItem, int, error) {
	if req == nil {
		req = &admindto.LoginLogListRequest{}
	}
	start, end, err := parseTimeRange(req.Start, req.End)
	if err != nil {
		return nil, 0, errcode.WithMessage(errcode.ParamError, "时间范围无效")
	}
	items, total, err := s.repo.ListLoginLogs(ctx, repository.LoginLogFilter{
		UserType:  req.UserType,
		Username:  req.Username,
		Status:    req.Status,
		StartTime: start,
		EndTime:   end,
		Offset:    req.GetOffset(),
		Limit:     req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.LoginLogItem, 0, len(items))
	for i := range items {
		out = append(out, toLoginLogItem(&items[i]))
	}
	return out, total, nil
}

func toSystemLogItem(m *model.SystemLogs) admindto.SystemLogItem {
	content := ""
	if m.Content != nil {
		content = *m.Content
	}
	return admindto.SystemLogItem{
		ID:        m.ID,
		Level:     m.Level,
		Module:    m.Module,
		Action:    m.Action,
		AdminID:   m.AdminID,
		Content:   content,
		IPAddress: m.IPAddress,
		CreatedAt: formatTime(m.CreatedAt),
	}
}

func toLoginLogItem(m *model.LoginLogs) admindto.LoginLogItem {
	return admindto.LoginLogItem{
		ID:        m.ID,
		UserType:  m.UserType,
		UserID:    m.UserID,
		Username:  m.Username,
		IPAddress: m.IPAddress,
		UserAgent: m.UserAgent,
		Status:    m.Status,
		CreatedAt: formatTime(m.CreatedAt),
	}
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseTimeRange(start, end string) (*time.Time, *time.Time, error) {
	var sPtr, ePtr *time.Time
	if start != "" {
		t, err := parseFlexibleTime(start)
		if err != nil {
			return nil, nil, err
		}
		sPtr = &t
	}
	if end != "" {
		t, err := parseFlexibleTime(end)
		if err != nil {
			return nil, nil, err
		}
		ePtr = &t
	}
	return sPtr, ePtr, nil
}

func parseFlexibleTime(v string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	var last error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, v, time.Local)
		if err == nil {
			return t, nil
		}
		last = err
	}
	return time.Time{}, last
}
