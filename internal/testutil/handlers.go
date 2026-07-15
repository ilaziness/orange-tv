// Package testutil provides shared test helpers.
package testutil

import (
	"context"

	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	adminhandler "github.com/ilaziness/orange-tv/internal/handler/http/admin"
	clienthandler "github.com/ilaziness/orange-tv/internal/handler/http/client"
	openhandler "github.com/ilaziness/orange-tv/internal/handler/http/open"
	"github.com/ilaziness/orange-tv/internal/model"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// BusinessHandlers holds no-op business handlers for route registration tests.
type BusinessHandlers struct {
	AuthService    adminsvc.AuthService
	AdminAuth      *adminhandler.AuthHandler
	AdminCategory  *adminhandler.CategoryHandler
	AdminVideo     *adminhandler.VideoHandler
	AdminMetadata  *adminhandler.MetadataHandler
	AdminPlay      *adminhandler.PlayHandler
	AdminLive      *adminhandler.LiveHandler
	AdminCollect   *adminhandler.CollectHandler
	AdminTheme     *adminhandler.ThemeHandler
	AdminSettings  *adminhandler.SettingsHandler
	AdminLog       *adminhandler.LogHandler
	ClientCategory *clienthandler.CategoryHandler
	ClientVideo    *clienthandler.VideoHandler
	ClientLive     *clienthandler.LiveHandler
	ClientTheme    *clienthandler.ThemeHandler
	ClientSite     *clienthandler.SiteHandler
	OpenResource   *openhandler.ResourceHandler
}

type authSvc struct{}

func (s authSvc) Login(ctx context.Context, req *admindto.LoginRequest, meta *adminsvc.LoginMeta) (*admindto.LoginResponse, error) {
	return &admindto.LoginResponse{}, nil
}
func (s authSvc) Profile(ctx context.Context, adminID int64) (*admindto.Profile, error) {
	return &admindto.Profile{ID: adminID}, nil
}
func (s authSvc) EnsureSuperAdmin(ctx context.Context, adminID int64) (*model.Admins, *model.UserGroups, error) {
	return &model.Admins{ID: adminID}, &model.UserGroups{Name: "super_admin"}, nil
}

type adminCategorySvc struct{}

func (s adminCategorySvc) ListTree(ctx context.Context, onlyEnabled bool) ([]shareddto.CategoryResponse, error) {
	return []shareddto.CategoryResponse{}, nil
}
func (s adminCategorySvc) Create(ctx context.Context, req *admindto.CreateCategoryRequest) (*shareddto.CategoryResponse, error) {
	return &shareddto.CategoryResponse{}, nil
}
func (s adminCategorySvc) Update(ctx context.Context, id int64, req *admindto.UpdateCategoryRequest) (*shareddto.CategoryResponse, error) {
	return &shareddto.CategoryResponse{ID: id}, nil
}
func (s adminCategorySvc) Delete(ctx context.Context, id int64) error { return nil }

type adminVideoSvc struct{}

func (s adminVideoSvc) List(ctx context.Context, req *admindto.VideoListRequest) ([]shareddto.VideoListItem, int, error) {
	return []shareddto.VideoListItem{}, 0, nil
}
func (s adminVideoSvc) Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{ID: id}, nil
}
func (s adminVideoSvc) Create(ctx context.Context, req *admindto.CreateVideoRequest) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{}, nil
}
func (s adminVideoSvc) Update(ctx context.Context, id int64, req *admindto.UpdateVideoRequest) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{ID: id}, nil
}
func (s adminVideoSvc) Delete(ctx context.Context, id int64) error { return nil }

type adminMetadataSvc struct{}

func (s adminMetadataSvc) ListDirectors(ctx context.Context, req *admindto.NameSearchRequest) ([]admindto.NamedResponse, int, error) {
	return nil, 0, nil
}
func (s adminMetadataSvc) CreateDirector(ctx context.Context, req *admindto.CreateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{}, nil
}
func (s adminMetadataSvc) UpdateDirector(ctx context.Context, id int64, req *admindto.UpdateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{ID: id}, nil
}
func (s adminMetadataSvc) DeleteDirector(ctx context.Context, id int64) error { return nil }
func (s adminMetadataSvc) ListActors(ctx context.Context, req *admindto.NameSearchRequest) ([]admindto.NamedResponse, int, error) {
	return nil, 0, nil
}
func (s adminMetadataSvc) CreateActor(ctx context.Context, req *admindto.CreateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{}, nil
}
func (s adminMetadataSvc) UpdateActor(ctx context.Context, id int64, req *admindto.UpdateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{ID: id}, nil
}
func (s adminMetadataSvc) DeleteActor(ctx context.Context, id int64) error { return nil }
func (s adminMetadataSvc) ListTags(ctx context.Context, req *admindto.NameSearchRequest) ([]admindto.NamedResponse, int, error) {
	return nil, 0, nil
}
func (s adminMetadataSvc) CreateTag(ctx context.Context, req *admindto.CreateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{}, nil
}
func (s adminMetadataSvc) UpdateTag(ctx context.Context, id int64, req *admindto.UpdateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{ID: id}, nil
}
func (s adminMetadataSvc) DeleteTag(ctx context.Context, id int64) error { return nil }

type adminPlaySvc struct{}

func (s adminPlaySvc) ListSources(ctx context.Context) ([]admindto.PlaySourceResponse, error) {
	return nil, nil
}
func (s adminPlaySvc) CreateSource(ctx context.Context, req *admindto.CreatePlaySourceRequest) (*admindto.PlaySourceResponse, error) {
	return &admindto.PlaySourceResponse{}, nil
}
func (s adminPlaySvc) UpdateSource(ctx context.Context, id int64, req *admindto.UpdatePlaySourceRequest) (*admindto.PlaySourceResponse, error) {
	return &admindto.PlaySourceResponse{ID: id}, nil
}
func (s adminPlaySvc) DeleteSource(ctx context.Context, id int64) error { return nil }
func (s adminPlaySvc) ListEpisodes(ctx context.Context, req *admindto.PlayEpisodeListRequest) ([]admindto.PlayEpisodeResponse, int, error) {
	return nil, 0, nil
}
func (s adminPlaySvc) CreateEpisode(ctx context.Context, req *admindto.CreatePlayEpisodeRequest) (*admindto.PlayEpisodeResponse, error) {
	return &admindto.PlayEpisodeResponse{}, nil
}
func (s adminPlaySvc) UpdateEpisode(ctx context.Context, id int64, req *admindto.UpdatePlayEpisodeRequest) (*admindto.PlayEpisodeResponse, error) {
	return &admindto.PlayEpisodeResponse{ID: id}, nil
}
func (s adminPlaySvc) DeleteEpisode(ctx context.Context, id int64) error { return nil }

type adminLiveSvc struct{}

func (s adminLiveSvc) List(ctx context.Context, req *admindto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error) {
	return nil, 0, nil
}
func (s adminLiveSvc) Create(ctx context.Context, req *admindto.CreateLiveRequest) (*shareddto.LiveChannelItem, error) {
	return &shareddto.LiveChannelItem{}, nil
}
func (s adminLiveSvc) Update(ctx context.Context, id int64, req *admindto.UpdateLiveRequest) (*shareddto.LiveChannelItem, error) {
	return &shareddto.LiveChannelItem{ID: id}, nil
}
func (s adminLiveSvc) Delete(ctx context.Context, id int64) error { return nil }

type adminCollectSvc struct{}

func (s adminCollectSvc) ListSources(ctx context.Context, req *admindto.CollectSourceListRequest) ([]shareddto.CollectSourceItem, int, error) {
	return nil, 0, nil
}
func (s adminCollectSvc) CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*shareddto.CollectSourceItem, error) {
	return &shareddto.CollectSourceItem{}, nil
}
func (s adminCollectSvc) UpdateSource(ctx context.Context, id int64, req *admindto.UpdateCollectSourceRequest) (*shareddto.CollectSourceItem, error) {
	return &shareddto.CollectSourceItem{ID: id}, nil
}
func (s adminCollectSvc) DeleteSource(ctx context.Context, id int64) error { return nil }
func (s adminCollectSvc) ListCategories(ctx context.Context, sourceID int64) ([]shareddto.CollectCategoryMapItem, error) {
	return nil, nil
}
func (s adminCollectSvc) SetCategories(ctx context.Context, sourceID int64, req *admindto.SetCollectCategoriesRequest) ([]shareddto.CollectCategoryMapItem, error) {
	return nil, nil
}
func (s adminCollectSvc) Start(ctx context.Context, sourceID int64) error { return nil }
func (s adminCollectSvc) Stop(ctx context.Context, sourceID int64) error  { return nil }
func (s adminCollectSvc) ListLogs(ctx context.Context, req *admindto.CollectLogListRequest) ([]shareddto.CollectLogItem, int, error) {
	return nil, 0, nil
}
func (s adminCollectSvc) ReloadScheduler(ctx context.Context) error { return nil }
func (s adminCollectSvc) StartScheduler(ctx context.Context) error  { return nil }
func (s adminCollectSvc) StopScheduler(ctx context.Context) error   { return nil }

type adminThemeSvc struct{}

func (s adminThemeSvc) List(ctx context.Context) ([]shareddto.ThemeListItem, error) { return nil, nil }
func (s adminThemeSvc) Update(ctx context.Context, id int64, req *admindto.UpdateThemeRequest) (*shareddto.ThemeListItem, error) {
	return &shareddto.ThemeListItem{ID: id}, nil
}
func (s adminThemeSvc) Activate(ctx context.Context, id int64) error { return nil }
func (s adminThemeSvc) EnsureDefaultTheme(ctx context.Context) error { return nil }

type adminSettingsSvc struct{}

func (s adminSettingsSvc) Get(ctx context.Context) (*admindto.SettingsResponse, error) {
	return &admindto.SettingsResponse{}, nil
}
func (s adminSettingsSvc) Update(ctx context.Context, req *admindto.UpdateSettingsRequest) (*admindto.SettingsResponse, error) {
	return &admindto.SettingsResponse{}, nil
}
func (s adminSettingsSvc) GetPublic(ctx context.Context) (*admindto.PublicSiteResponse, error) {
	return &admindto.PublicSiteResponse{Name: "Orange TV"}, nil
}
func (s adminSettingsSvc) ResourceConfig(ctx context.Context) (*adminsvc.ResourceConfig, error) {
	return &adminsvc.ResourceConfig{SiteMode: "video_site", APIOutputFormat: "default", EnableThirdPartyCollect: true}, nil
}
func (s adminSettingsSvc) InvalidateCache(ctx context.Context) {}

type adminLogSvc struct{}

func (s adminLogSvc) ListSystemLogs(ctx context.Context, req *admindto.SystemLogListRequest) ([]admindto.SystemLogItem, int, error) {
	return nil, 0, nil
}
func (s adminLogSvc) ListLoginLogs(ctx context.Context, req *admindto.LoginLogListRequest) ([]admindto.LoginLogItem, int, error) {
	return nil, 0, nil
}

type clientCategorySvc struct{}

func (s clientCategorySvc) ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	return []shareddto.CategoryResponse{}, nil
}

type clientVideoSvc struct{}

func (s clientVideoSvc) List(ctx context.Context, req *clientdto.VideoListRequest) ([]shareddto.VideoListItem, int, error) {
	return nil, 0, nil
}
func (s clientVideoSvc) Search(ctx context.Context, req *clientdto.SearchRequest) ([]shareddto.VideoListItem, int, error) {
	return nil, 0, nil
}
func (s clientVideoSvc) Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{ID: id}, nil
}
func (s clientVideoSvc) Related(ctx context.Context, id int64, limit int) ([]shareddto.VideoListItem, error) {
	return nil, nil
}

type clientLiveSvc struct{}

func (s clientLiveSvc) List(ctx context.Context, req *clientdto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error) {
	return nil, 0, nil
}

type clientThemeSvc struct{}

func (s clientThemeSvc) Current(ctx context.Context) (*shareddto.ThemeCurrentResponse, error) {
	return &shareddto.ThemeCurrentResponse{Name: "默认主题", Identifier: "default", Config: map[string]any{}}, nil
}

type openResourceSvc struct{}

func (s openResourceSvc) Authorize(ctx context.Context, providedKey string) (*adminsvc.ResourceConfig, error) {
	return &adminsvc.ResourceConfig{EnableThirdPartyCollect: true, APIOutputFormat: "default"}, nil
}
func (s openResourceSvc) ListVideos(ctx context.Context, page, pageSize int, format string) (any, error) {
	return map[string]any{"code": 200}, nil
}
func (s openResourceSvc) GetVideo(ctx context.Context, id int64, format string) (any, error) {
	return map[string]any{"code": 200}, nil
}
func (s openResourceSvc) ListCategories(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	return []shareddto.CategoryResponse{}, nil
}

// NewBusinessHandlers builds no-op business handlers for tests.
func NewBusinessHandlers() BusinessHandlers {
	auth := authSvc{}
	return BusinessHandlers{
		AuthService:    auth,
		AdminAuth:      adminhandler.NewAuthHandler(auth),
		AdminCategory:  adminhandler.NewCategoryHandler(adminCategorySvc{}),
		AdminVideo:     adminhandler.NewVideoHandler(adminVideoSvc{}),
		AdminMetadata:  adminhandler.NewMetadataHandler(adminMetadataSvc{}),
		AdminPlay:      adminhandler.NewPlayHandler(adminPlaySvc{}),
		AdminLive:      adminhandler.NewLiveHandler(adminLiveSvc{}),
		AdminCollect:   adminhandler.NewCollectHandler(adminCollectSvc{}),
		AdminTheme:     adminhandler.NewThemeHandler(adminThemeSvc{}),
		AdminSettings:  adminhandler.NewSettingsHandler(adminSettingsSvc{}, nil),
		AdminLog:       adminhandler.NewLogHandler(adminLogSvc{}),
		ClientCategory: clienthandler.NewCategoryHandler(clientCategorySvc{}),
		ClientVideo:    clienthandler.NewVideoHandler(clientVideoSvc{}),
		ClientLive:     clienthandler.NewLiveHandler(clientLiveSvc{}),
		ClientTheme:    clienthandler.NewThemeHandler(clientThemeSvc{}),
		ClientSite:     clienthandler.NewSiteHandler(adminSettingsSvc{}),
		OpenResource:   openhandler.NewResourceHandler(openResourceSvc{}),
	}
}
