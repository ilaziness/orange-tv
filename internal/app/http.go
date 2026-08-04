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
	"github.com/ilaziness/orange-tv/internal/logger"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/ilaziness/orange-tv/internal/scheduler"
	"github.com/ilaziness/orange-tv/internal/server"
	"github.com/ilaziness/orange-tv/internal/service"
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
	commentRepo := repository.NewCommentRepo(a.db)

	sharedSettingsSvc := service.NewSettingsService(settingsRepo, a.cache, a.log)

	recorder := audit.NewRecorder(logRepo, a.log)

	authSvc := adminsvc.NewAuthService(adminRepo, a.jwtMgr, a.cfg, recorder, a.log)
	adminCategorySvc := adminsvc.NewCategoryService(categoryRepo, a.cache, a.log)
	adminMetaSvc := adminsvc.NewMetadataService(metaRepo, a.log)
	adminPlaySvc := adminsvc.NewPlayService(playRepo, videoRepo, a.log)
	adminVideoSvc := adminsvc.NewVideoService(videoRepo, categoryRepo, metaRepo, playRepo, a.cache, a.log)
	adminLiveSvc := adminsvc.NewLiveService(liveRepo, a.cache, a.log)
	adminCommentSvc := adminsvc.NewCommentService(commentRepo, a.log)
	collectEngine := collect.NewEngine(collectRepo, videoRepo, categoryRepo, metaRepo, playRepo, a.log)
	adminCollectSvc := adminsvc.NewCollectService(collectRepo, playRepo, categoryRepo, collectEngine, a.log, a.cache)
	adminSettingsSvc := adminsvc.NewSettingsService(sharedSettingsSvc, a.log)
	clientSettingsSvc := clientsvc.NewClientSettingsService(sharedSettingsSvc)
	adminLogSvc := adminsvc.NewLogService(logRepo, a.log, a.cfg.Log.Filename)
	adminMgmtSvc := adminsvc.NewManagementService(adminRepo, videoRepo, userFeatureRepo, recorder, a.log)
	adminDataSvc := adminsvc.NewDataService(a.db, a.cfg, logRepo, a.log)

	clientCategorySvc := clientsvc.NewCategoryService(categoryRepo, a.cache, a.log)
	clientVideoSvc := clientsvc.NewVideoService(videoRepo, metaRepo, playRepo, a.cache, a.log)
	clientLiveSvc := clientsvc.NewLiveService(liveRepo, a.cache, a.log)
	clientLiveProxySvc := clientsvc.NewLiveProxyService(clientLiveSvc, a.log)
	clientUserSvc := clientsvc.NewUserService(adminRepo, userFeatureRepo, videoRepo, categoryRepo, a.jwtMgr, a.cfg.JWT.AccessTokenTTL, sharedSettingsSvc, a.log)
	clientBannerSvc := clientsvc.NewBannerService(userFeatureRepo, a.log)

	openResourceSvc := opensvc.NewResourceService(settingsRepo, videoRepo, metaRepo, playRepo, categoryRepo, a.cache, a.log)

	handlers.AuthService = authSvc
	handlers.AdminAuth = adminhandler.NewAuthHandler(authSvc)
	handlers.AdminCategory = adminhandler.NewCategoryHandler(adminCategorySvc)
	handlers.AdminVideo = adminhandler.NewVideoHandler(adminVideoSvc)
	handlers.AdminMetadata = adminhandler.NewMetadataHandler(adminMetaSvc)
	handlers.AdminPlay = adminhandler.NewPlayHandler(adminPlaySvc)
	handlers.AdminLive = adminhandler.NewLiveHandler(adminLiveSvc)
	handlers.AdminComment = adminhandler.NewCommentHandler(adminCommentSvc, recorder)
	handlers.AdminCollect = adminhandler.NewCollectHandler(adminCollectSvc)
	handlers.AdminSettings = adminhandler.NewSettingsHandler(adminSettingsSvc, recorder)
	handlers.AdminLog = adminhandler.NewLogHandler(adminLogSvc)
	handlers.AdminMgmt = adminhandler.NewManagementHandler(adminMgmtSvc)
	handlers.AdminData = adminhandler.NewDataHandler(adminDataSvc)

	handlers.ClientCategory = clienthandler.NewCategoryHandler(clientCategorySvc)
	handlers.ClientVideo = clienthandler.NewVideoHandler(clientVideoSvc)
	handlers.ClientLive = clienthandler.NewLiveHandler(clientLiveSvc, clientLiveProxySvc)
	handlers.ClientSettings = clienthandler.NewSettingsHandler(clientSettingsSvc)
	handlers.ClientUser = clienthandler.NewUserHandler(clientUserSvc)
	handlers.ClientBanner = clienthandler.NewBannerHandler(clientBannerSvc)
	handlers.OpenResource = openhandler.NewResourceHandler(openResourceSvc)

	// ── scheduler ──────────────────────────────────────────────────────────
	// 创建 cron 管理器、注册所有调度任务，并绑定生命周期 hook。
	// Manager 统一启动/停止所有注册的 Job，app 层无需关心任务内部实现。
	cronMgr := scheduler.NewManager()
	scheduler.Setup(cronMgr, scheduler.Deps{
		CollectRepo:   collectRepo,
		CollectRunner: adminCollectSvc,
	})

	a.addHook(Hook{
		Name: "cron_manager",
		OnStart: func(ctx context.Context) error {
			return cronMgr.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return cronMgr.Stop(ctx)
		},
	})

	// Access logs (gin request logs) go to stdout only so they don't clutter
	// the application log file. Business/recovery logs still use a.log.
	accessLogger := logger.NewStdoutLogger(a.cfg.Log.Level)
	httpServer, err := server.NewHTTPServer(a.cfg, a.log, accessLogger, handlers, a.metrics, a.jwtMgr)
	if err != nil {
		return err
	}
	a.httpServer = httpServer
	return nil
}
