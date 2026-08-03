// Package audit provides fire-and-forget helpers for login / system operation logs.
package audit

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// Recorder writes audit logs without failing the main request path.
type Recorder struct {
	repo   repository.LogRepository
	logger *zap.Logger
}

// NewRecorder creates a Recorder. repo/logger may be nil (no-op).
func NewRecorder(repo repository.LogRepository, logger *zap.Logger) *Recorder {
	return &Recorder{repo: repo, logger: logger}
}

// AdminLogin records an admin login attempt.
func (r *Recorder) AdminLogin(ctx context.Context, userID uint32, username, ip, ua string, success bool) {
	if r == nil || r.repo == nil {
		return
	}
	status := constant.LoginStatusSuccess
	if !success {
		status = constant.LoginStatusFailed
	}
	m := &model.AdminLoginLogs{
		UserID:    userID,
		Username:  strings.TrimSpace(username),
		IP:        trimIP(ip),
		UserAgent: trimUA(ua),
		Status:    status,
	}
	if err := r.repo.CreateAdminLoginLog(ctx, m); err != nil && r.logger != nil {
		r.logger.Warn("record admin login log failed", zap.Error(err))
	}
}

// AdminAction records an admin write operation.
func (r *Recorder) AdminAction(ctx context.Context, adminID uint32, module, action, content, ip string) {
	if r == nil || r.repo == nil {
		return
	}
	m := &model.SystemLogs{
		Level:     constant.SystemLogLevelInfo,
		Module:    strings.TrimSpace(module),
		Action:    strings.TrimSpace(action),
		AdminID:   adminID,
		Content:   content,
		IPAddress: trimIP(ip),
	}
	if err := r.repo.CreateSystemLog(ctx, m); err != nil && r.logger != nil {
		r.logger.Warn("record system log failed", zap.Error(err))
	}
}

// Warning records a warning-level system log (e.g. resource API key failure).
func (r *Recorder) Warning(ctx context.Context, module, action, content, ip string) {
	if r == nil || r.repo == nil {
		return
	}
	m := &model.SystemLogs{
		Level:     constant.SystemLogLevelWarning,
		Module:    strings.TrimSpace(module),
		Action:    strings.TrimSpace(action),
		AdminID:   0,
		Content:   content,
		IPAddress: trimIP(ip),
	}
	if err := r.repo.CreateSystemLog(ctx, m); err != nil && r.logger != nil {
		r.logger.Warn("record system warning log failed", zap.Error(err))
	}
}

func trimIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if len(ip) > 45 {
		return ip[:45]
	}
	return ip
}

func trimUA(ua string) string {
	ua = strings.TrimSpace(ua)
	if len(ua) > 500 {
		return ua[:500]
	}
	return ua
}
