package app

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/collect"
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
	liveRepo := repository.NewLiveRepo(a.db)
	collectRepo := repository.NewCollectRepo(a.db)
	themeRepo := repository.NewThemeRepo(a.db)

	authSvc := adminsvc.NewAuthService(adminRepo, a.jwtMgr, a.cfg)
	adminCategorySvc := adminsvc.NewCategoryService(categoryRepo)
	adminMetaSvc := adminsvc.NewMetadataService(metaRepo)
	adminPlaySvc := adminsvc.NewPlayService(playRepo, videoRepo)
	adminVideoSvc := adminsvc.NewVideoService(videoRepo, categoryRepo, metaRepo, playRepo)
	adminLiveSvc := adminsvc.NewLiveService(liveRepo)
	collectEngine := collect.NewEngine(collectRepo, videoRepo, categoryRepo, metaRepo, playRepo, a.log)
	adminCollectSvc := adminsvc.NewCollectService(collectRepo, playRepo, categoryRepo, collectEngine, a.log)
	adminThemeSvc := adminsvc.NewThemeService(themeRepo)

	clientCategorySvc := clientsvc.NewCategoryService(categoryRepo)
	clientVideoSvc := clientsvc.NewVideoService(videoRepo, metaRepo, playRepo)
	clientLiveSvc := clientsvc.NewLiveService(liveRepo)
	clientThemeSvc := clientsvc.NewThemeService(themeRepo, adminThemeSvc)

	handlers.AuthService = authSvc
	handlers.AdminAuth = adminhandler.NewAuthHandler(authSvc)
	handlers.AdminCategory = adminhandler.NewCategoryHandler(adminCategorySvc)
	handlers.AdminVideo = adminhandler.NewVideoHandler(adminVideoSvc)
	handlers.AdminMetadata = adminhandler.NewMetadataHandler(adminMetaSvc)
	handlers.AdminPlay = adminhandler.NewPlayHandler(adminPlaySvc)
	handlers.AdminLive = adminhandler.NewLiveHandler(adminLiveSvc)
	handlers.AdminCollect = adminhandler.NewCollectHandler(adminCollectSvc)
	handlers.AdminTheme = adminhandler.NewThemeHandler(adminThemeSvc)

	handlers.ClientCategory = clienthandler.NewCategoryHandler(clientCategorySvc)
	handlers.ClientVideo = clienthandler.NewVideoHandler(clientVideoSvc)
	handlers.ClientLive = clienthandler.NewLiveHandler(clientLiveSvc)
	handlers.ClientTheme = clienthandler.NewThemeHandler(clientThemeSvc)

	// collect scheduler lifecycle
	a.addHook(Hook{
		Name: "collect_scheduler",
		OnStart: func(ctx context.Context) error {
			return adminCollectSvc.StartScheduler(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return adminCollectSvc.StopScheduler(ctx)
		},
	})
	// ensure default theme on boot
	a.addHook(Hook{
		Name: "default_theme",
		OnStart: func(ctx context.Context) error {
			return adminThemeSvc.EnsureDefaultTheme(ctx)
		},
	})

	httpServer, err := server.NewHTTPServer(a.cfg, a.log, handlers, a.metrics, a.jwtMgr)
	if err != nil {
		return err
	}
	a.httpServer = httpServer
	return nil
}
