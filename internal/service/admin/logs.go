package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// LogService queries login / system logs for admin.
type LogService interface {
	ListSystemLogs(ctx context.Context, req *admindto.SystemLogListRequest) ([]admindto.SystemLogItem, int, error)
	ListLoginLogs(ctx context.Context, req *admindto.LoginLogListRequest) ([]admindto.LoginLogItem, int, error)
	ListAppLogs(ctx context.Context, req *admindto.AppLogListRequest) (*admindto.AppLogListResponse, error)
}

type logService struct {
	repo    repository.LogRepository
	log     *zap.Logger
	logFile string
}

// NewLogService creates a LogService.
func NewLogService(repo repository.LogRepository, log *zap.Logger, logFile string) LogService {
	if log == nil {
		log = zap.NewNop()
	}
	return &logService{repo: repo, log: log, logFile: logFile}
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
		s.log.Error("logs: list system logs failed", zap.Error(err))
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
		s.log.Error("logs: list login logs failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.LoginLogItem, 0, len(items))
	for i := range items {
		out = append(out, toLoginLogItem(&items[i]))
	}
	return out, total, nil
}

func (s *logService) ListAppLogs(ctx context.Context, req *admindto.AppLogListRequest) (*admindto.AppLogListResponse, error) {
	if req == nil {
		req = &admindto.AppLogListRequest{}
	}
	limit := req.GetLimit()

	if s.logFile == "" {
		return &admindto.AppLogListResponse{List: []admindto.AppLogItem{}, HasMore: false}, nil
	}

	info, err := os.Stat(s.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &admindto.AppLogListResponse{List: []admindto.AppLogItem{}, HasMore: false}, nil
		}
		s.log.Error("logs: stat app log file failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	fileSize := info.Size()
	if fileSize == 0 {
		return &admindto.AppLogListResponse{List: []admindto.AppLogItem{}, HasMore: false}, nil
	}

	// offset is how many bytes from the end we've already consumed.
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= fileSize {
		return &admindto.AppLogListResponse{List: []admindto.AppLogItem{}, HasMore: false}, nil
	}

	f, err := os.Open(s.logFile)
	if err != nil {
		s.log.Error("logs: open app log file failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	defer f.Close()

	const chunkSize = 4096
	var collected []admindto.AppLogItem
	var leftover []byte
	consumed := offset

	for {
		if err := ctx.Err(); err != nil {
			return nil, errcode.Wrap(errcode.InternalError, err)
		}

		// Read a chunk from position (fileSize - consumed - chunkSize).
		readStart := fileSize - consumed - chunkSize
		readLen := chunkSize
		if readStart < 0 {
			readLen = int(fileSize - consumed)
			readStart = 0
		}
		if readLen <= 0 {
			break
		}

		buf := make([]byte, readLen)
		n, err := f.ReadAt(buf, readStart)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			s.log.Error("logs: read app log chunk failed", zap.Error(err))
			return nil, errcode.Wrap(errcode.InternalError, err)
		}
		buf = buf[:n]

		// Append leftover from the previous (newer) chunk to the end.
		// leftover is the partial line at the beginning of the newer chunk;
		// its beginning is in the current (earlier) chunk.
		data := append(buf, leftover...)
		leftover = nil

		// Split into lines.
		lines := bytes.Split(data, []byte("\n"))

		// The FIRST element may be a partial line that continues in the
		// next (earlier) chunk — save it as leftover.
		// Exception: if we're at the beginning of the file, the first
		// element is a complete line.
		if readStart > 0 && len(lines) > 0 {
			if len(lines[0]) > 0 {
				leftover = lines[0]
			}
			lines = lines[1:]
		}

		// Process lines in reverse order (newest first within this chunk).
		for i := len(lines) - 1; i >= 0; i-- {
			line := bytes.TrimSpace(lines[i])
			if len(line) == 0 {
				continue
			}
			item, ok := parseAppLogLine(line)
			if !ok {
				continue
			}
			collected = append(collected, item)
			if len(collected) >= limit {
				totalConsumed := consumed + int64(n) - int64(len(leftover))
				hasMore := readStart > 0 || len(leftover) > 0
				return &admindto.AppLogListResponse{
					List:       collected,
					HasMore:    hasMore,
					NextOffset: totalConsumed,
				}, nil
			}
		}

		consumed += int64(n)
		if readStart == 0 {
			break
		}
	}

	// Reached the beginning of the file.
	hasMore := false
	nextOffset := consumed
	if len(leftover) > 0 {
		line := bytes.TrimSpace(leftover)
		if len(line) > 0 {
			if item, ok := parseAppLogLine(line); ok {
				collected = append(collected, item)
			}
		}
		nextOffset = fileSize
	}

	return &admindto.AppLogListResponse{
		List:       collected,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}, nil
}

// parseAppLogLine parses a single JSON log line into AppLogItem.
// Returns false if JSON parsing fails.
func parseAppLogLine(line []byte) (admindto.AppLogItem, bool) {
	var raw map[string]any
	if err := json.Unmarshal(line, &raw); err != nil {
		return admindto.AppLogItem{}, false
	}

	item := admindto.AppLogItem{
		Time:  formatLogTime(extractString(raw, "time")),
		Level: extractString(raw, "level"),
		Msg:   extractString(raw, "msg"),
	}

	// Remove known keys; remaining go into Fields.
	delete(raw, "time")
	delete(raw, "level")
	delete(raw, "msg")
	delete(raw, "caller")
	if len(raw) > 0 {
		item.Fields = raw
	}

	return item, true
}

func extractString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func formatLogTime(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return s
		}
	}
	return t.Local().Format(time.DateTime)
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
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
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
