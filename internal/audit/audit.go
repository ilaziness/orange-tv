// Package audit provides fire-and-forget helpers for login / system operation logs.
package audit

import (
	"context"
	"strings"
	"time"

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

// Login records an admin/user login attempt.
func (r *Recorder) Login(ctx context.Context, userType uint8, userID int64, username, ip, ua string, success bool) {
	if r == nil || r.repo == nil {
		return
	}
	status := constant.LoginStatusSuccess
	if !success {
		status = constant.LoginStatusFailed
	}
	now := time.Now()
	m := &model.LoginLogs{
		UserType:  userType,
		UserID:    uint64(userID),
		Username:  strings.TrimSpace(username),
		IPAddress: trimIP(ip),
		UserAgent: trimUA(ua),
		Status:    status,
		CreatedAt: &now,
	}
	if err := r.repo.CreateLoginLog(ctx, m); err != nil && r.logger != nil {
		r.logger.Warn("record login log failed", zap.Error(err))
	}
}

// AdminAction records an admin write operation.
func (r *Recorder) AdminAction(ctx context.Context, adminID int64, module, action, content, ip string) {
	if r == nil || r.repo == nil {
		return
	}
	now := time.Now()
	c := content
	m := &model.SystemLogs{
		Level:     constant.SystemLogLevelInfo,
		Module:    strings.TrimSpace(module),
		Action:    strings.TrimSpace(action),
		AdminID:   uint64(adminID),
		Content:   &c,
		IPAddress: trimIP(ip),
		CreatedAt: &now,
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
	now := time.Now()
	c := content
	m := &model.SystemLogs{
		Level:     constant.SystemLogLevelWarning,
		Module:    strings.TrimSpace(module),
		Action:    strings.TrimSpace(action),
		AdminID:   0,
		Content:   &c,
		IPAddress: trimIP(ip),
		CreatedAt: &now,
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
