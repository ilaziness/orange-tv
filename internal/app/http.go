package app

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/audit"
	"github.com/ilaziness/orange-tv/internal/collect"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	adminhandler "github.com/ilaziness/orange-tv/internal/handler/http/admin"
	clienthandler "github.com/ilaziness/orange-tv/internal/handler/http/client"
	openhandler "github.com/ilaziness/orange-tv/internal/handler/http/open"
	"github.com/ilaziness/orange-tv/internal/health"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/ilaziness/orange-tv/internal/server"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
	opensvc "github.com/ilaziness/orange-tv/internal/service/open"
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
	settingsRepo := repository.NewSettingsRepo(a.db)
	logRepo := repository.NewLogRepo(a.db)
	userFeatureRepo := repository.NewUserFeatureRepo(a.db)

	recorder := audit.NewRecorder(logRepo, a.log)

	authSvc := adminsvc.NewAuthService(adminRepo, a.jwtMgr, a.cfg, recorder, a.log)
	adminCategorySvc := adminsvc.NewCategoryService(categoryRepo, a.cache, a.log)
	adminMetaSvc := adminsvc.NewMetadataService(metaRepo, a.log)
	adminPlaySvc := adminsvc.NewPlayService(playRepo, videoRepo, a.log)
	adminVideoSvc := adminsvc.NewVideoService(videoRepo, categoryRepo, metaRepo, playRepo, a.cache, a.log)
	adminLiveSvc := adminsvc.NewLiveService(liveRepo, a.log)
	collectEngine := collect.NewEngine(collectRepo, videoRepo, categoryRepo, metaRepo, playRepo, a.log)
	adminCollectSvc := adminsvc.NewCollectService(collectRepo, playRepo, categoryRepo, collectEngine, a.log, a.cache)
	adminSettingsSvc := adminsvc.NewSettingsService(settingsRepo, a.cache, a.log)
	adminLogSvc := adminsvc.NewLogService(logRepo, a.log)
	adminMgmtSvc := adminsvc.NewManagementService(adminRepo, videoRepo, userFeatureRepo, recorder, a.log)

	clientCategorySvc := clientsvc.NewCategoryService(categoryRepo, a.cache, a.log)
	clientVideoSvc := clientsvc.NewVideoService(videoRepo, metaRepo, playRepo, a.cache, a.log)
	clientLiveSvc := clientsvc.NewLiveService(liveRepo, a.log)
	clientUserSvc := clientsvc.NewUserService(adminRepo, userFeatureRepo, videoRepo, a.jwtMgr, a.cfg.JWT.AccessTokenTTL, a.log)
	clientBannerSvc := clientsvc.NewBannerService(userFeatureRepo, a.log)

	openResourceSvc := opensvc.NewResourceService(adminSettingsSvc, videoRepo, metaRepo, playRepo, categoryRepo, a.cache, a.log)

	handlers.AuthService = authSvc
	handlers.AdminAuth = adminhandler.NewAuthHandler(authSvc)
	handlers.AdminCategory = adminhandler.NewCategoryHandler(adminCategorySvc)
	handlers.AdminVideo = adminhandler.NewVideoHandler(adminVideoSvc)
	handlers.AdminMetadata = adminhandler.NewMetadataHandler(adminMetaSvc)
	handlers.AdminPlay = adminhandler.NewPlayHandler(adminPlaySvc)
	handlers.AdminLive = adminhandler.NewLiveHandler(adminLiveSvc)
	handlers.AdminCollect = adminhandler.NewCollectHandler(adminCollectSvc)
	handlers.AdminSettings = adminhandler.NewSettingsHandler(adminSettingsSvc, recorder)
	handlers.AdminLog = adminhandler.NewLogHandler(adminLogSvc)
	handlers.AdminMgmt = adminhandler.NewManagementHandler(adminMgmtSvc)

	handlers.ClientCategory = clienthandler.NewCategoryHandler(clientCategorySvc)
	handlers.ClientVideo = clienthandler.NewVideoHandler(clientVideoSvc)
	handlers.ClientLive = clienthandler.NewLiveHandler(clientLiveSvc)
	handlers.ClientSite = clienthandler.NewSiteHandler(adminSettingsSvc)
	handlers.ClientUser = clienthandler.NewUserHandler(clientUserSvc)
	handlers.ClientBanner = clienthandler.NewBannerHandler(clientBannerSvc)
	handlers.OpenResource = openhandler.NewResourceHandler(openResourceSvc)

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
	httpServer, err := server.NewHTTPServer(a.cfg, a.log, handlers, a.metrics, a.jwtMgr)
	if err != nil {
		return err
	}
	a.httpServer = httpServer
	return nil
}
