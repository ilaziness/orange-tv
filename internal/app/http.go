package app

import (
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	adminhandler "github.com/ilaziness/orange-tv/internal/handler/http/admin"
	clienthandler "github.com/ilaziness/orange-tv/internal/handler/http/client"
	"github.com/ilaziness/orange-tv/internal/health"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/ilaziness/orange-tv/internal/server"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
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
	// 未配置 jwt.secret 时全局不挂 JWTAuth。
	// 配置 secret 后 jwtMgr 非 nil，admin 业务路由 RequireSuperAdmin 生效。
	handlers.RequireAdminAuth = a.jwtMgr != nil

	adminRepo := repository.NewAdminRepo(a.db)
	categoryRepo := repository.NewCategoryRepo(a.db)
	videoRepo := repository.NewVideoRepo(a.db)
	metaRepo := repository.NewMetadataRepo(a.db)
	playRepo := repository.NewPlayRepo(a.db)

	authSvc := adminsvc.NewAuthService(adminRepo, a.jwtMgr, a.cfg)
	adminCategorySvc := adminsvc.NewCategoryService(categoryRepo)
	adminMetaSvc := adminsvc.NewMetadataService(metaRepo)
	adminPlaySvc := adminsvc.NewPlayService(playRepo, videoRepo)
	adminVideoSvc := adminsvc.NewVideoService(videoRepo, categoryRepo, metaRepo, playRepo)

	clientCategorySvc := clientsvc.NewCategoryService(categoryRepo)
	clientVideoSvc := clientsvc.NewVideoService(videoRepo, metaRepo, playRepo)

	handlers.AuthService = authSvc
	handlers.AdminAuth = adminhandler.NewAuthHandler(authSvc)
	handlers.AdminCategory = adminhandler.NewCategoryHandler(adminCategorySvc)
	handlers.AdminVideo = adminhandler.NewVideoHandler(adminVideoSvc)
	handlers.AdminMetadata = adminhandler.NewMetadataHandler(adminMetaSvc)
	handlers.AdminPlay = adminhandler.NewPlayHandler(adminPlaySvc)

	handlers.ClientCategory = clienthandler.NewCategoryHandler(clientCategorySvc)
	handlers.ClientVideo = clienthandler.NewVideoHandler(clientVideoSvc)

	httpServer, err := server.NewHTTPServer(a.cfg, a.log, handlers, a.metrics, a.jwtMgr)
	if err != nil {
		return err
	}
	a.httpServer = httpServer
	return nil
}
