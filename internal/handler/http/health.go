package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/health"
	"github.com/ilaziness/orange-tv/internal/response"
)

const readinessCheckTimeout = 5 * time.Second

type HealthHandler struct {
	appName    string
	appVersion string
	checkers   []health.Checker
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		appName:    cfg.App.Name,
		appVersion: cfg.App.Version,
		checkers:   make([]health.Checker, 0),
	}
}

func (h *HealthHandler) AddChecker(checker health.Checker) {
	h.checkers = append(h.checkers, checker)
}

// Health is a lightweight liveness probe; it does not check dependencies.
// @Summary 健康检查
// @Description 检查应用是否正常运行（不检查外部依赖）
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	h.Liveness(c)
}

// Readiness checks whether the application and its dependencies are ready to serve traffic.
// @Summary 就绪检查
// @Description 检查应用及其依赖是否就绪
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]any}
// @Failure 503 {object} response.Response{data=map[string]any}
// @Router /readiness [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessCheckTimeout)
	defer cancel()

	status := "ready"
	checks := make(map[string]bool)
	allReady := true

	for _, checker := range h.checkers {
		ready := checker.Check(ctx) == nil
		checks[checker.Name()] = ready
		if !ready {
			allReady = false
			status = "not_ready"
		}
	}

	payload := gin.H{
		"status": status,
		"checks": checks,
	}

	if allReady {
		response.Success(c, payload)
		return
	}

	response.ErrorWithData(c, errcode.WithMessage(errcode.ServiceUnavailable, status), payload)
}

// Liveness reports whether the process is alive; it does not check dependencies.
// @Summary 存活检查
// @Description 检查应用是否存活（不检查外部依赖）
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /liveness [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "alive",
	})
}

// Version returns application name and version.
// @Summary 获取版本信息
// @Description 获取应用名称和版本
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /version [get]
func (h *HealthHandler) Version(c *gin.Context) {
	response.Success(c, gin.H{
		"name":    h.appName,
		"version": h.appVersion,
	})
}
