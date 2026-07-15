package app

import (
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/health"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/ilaziness/orange-tv/internal/server"
)

func (a *App) wireHTTP() error {
	if !a.cfg.HTTP.Enabled {
		return nil
	}

	healthHandler := httphandler.NewHealthHandler(a.cfg)
	healthHandler.AddChecker(health.NewDatabaseChecker(a.db))

	handlers, err := router.NewHandlers(healthHandler)
	if err != nil {
		return err
	}
	handlers.InternalServiceKey = a.cfg.HTTP.InternalServiceKey
	// 未配置 jwt.secret 时全局不挂 JWTAuth；骨架阶段允许管理端 stub 无 token 联调。
	// 配置 secret 后 jwtMgr 非 nil，admin 业务路由 RequireAuth 生效。
	handlers.RequireAdminAuth = a.jwtMgr != nil

	httpServer, err := server.NewHTTPServer(a.cfg, a.log, handlers, a.metrics, a.jwtMgr)
	if err != nil {
		return err
	}
	a.httpServer = httpServer
	return nil
}
