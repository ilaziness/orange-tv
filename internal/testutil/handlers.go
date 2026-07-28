// Package testutil provides shared test helpers.
package testutil

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
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
	AdminSettings  *adminhandler.SettingsHandler
	AdminLog       *adminhandler.LogHandler
	AdminMgmt      *adminhandler.ManagementHandler
	AdminData      *adminhandler.DataHandler
	ClientCategory *clienthandler.CategoryHandler
	ClientVideo    *clienthandler.VideoHandler
	ClientLive     *clienthandler.LiveHandler
	ClientSite     *clienthandler.SiteHandler
	ClientUser     *clienthandler.UserHandler
	ClientBanner   *clienthandler.BannerHandler
	OpenResource   *openhandler.ResourceHandler
}

type authSvc struct{}

func (s authSvc) Login(ctx context.Context, req *admindto.LoginRequest, meta *adminsvc.LoginMeta) (*admindto.LoginResponse, error) {
	return &admindto.LoginResponse{}, nil
}
func (s authSvc) Profile(ctx context.Context, adminID int64) (*admindto.Profile, error) {
	return &admindto.Profile{ID: uint64(adminID)}, nil
}
func (s authSvc) EnsureSuperAdmin(ctx context.Context, adminID int64) (*model.Admins, *model.UserGroups, error) {
	return &model.Admins{ID: uint64(adminID)}, &model.UserGroups{Name: "super_admin"}, nil
}

type adminCategorySvc struct{}

func (s adminCategorySvc) ListTree(ctx context.Context, onlyEnabled bool) ([]shareddto.CategoryResponse, error) {
	return []shareddto.CategoryResponse{}, nil
}
func (s adminCategorySvc) Create(ctx context.Context, req *admindto.CreateCategoryRequest) (*shareddto.CategoryResponse, error) {
	return &shareddto.CategoryResponse{}, nil
}
func (s adminCategorySvc) Update(ctx context.Context, id int64, req *admindto.UpdateCategoryRequest) (*shareddto.CategoryResponse, error) {
	return &shareddto.CategoryResponse{ID: uint64(id)}, nil
}
func (s adminCategorySvc) Delete(ctx context.Context, id int64) error { return nil }

type adminVideoSvc struct{}

func (s adminVideoSvc) List(ctx context.Context, req *admindto.VideoListRequest) ([]shareddto.VideoListItem, int, error) {
	return []shareddto.VideoListItem{}, 0, nil
}
func (s adminVideoSvc) Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{ID: uint64(id)}, nil
}
func (s adminVideoSvc) Create(ctx context.Context, req *admindto.CreateVideoRequest) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{}, nil
}
func (s adminVideoSvc) Update(ctx context.Context, id int64, req *admindto.UpdateVideoRequest) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{ID: uint64(id)}, nil
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
	return &admindto.NamedResponse{ID: uint64(id)}, nil
}
func (s adminMetadataSvc) DeleteDirector(ctx context.Context, id int64) error { return nil }
func (s adminMetadataSvc) ListActors(ctx context.Context, req *admindto.NameSearchRequest) ([]admindto.NamedResponse, int, error) {
	return nil, 0, nil
}
func (s adminMetadataSvc) CreateActor(ctx context.Context, req *admindto.CreateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{}, nil
}
func (s adminMetadataSvc) UpdateActor(ctx context.Context, id int64, req *admindto.UpdateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{ID: uint64(id)}, nil
}
func (s adminMetadataSvc) DeleteActor(ctx context.Context, id int64) error { return nil }
func (s adminMetadataSvc) ListTags(ctx context.Context, req *admindto.NameSearchRequest) ([]admindto.NamedResponse, int, error) {
	return nil, 0, nil
}
func (s adminMetadataSvc) CreateTag(ctx context.Context, req *admindto.CreateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{}, nil
}
func (s adminMetadataSvc) UpdateTag(ctx context.Context, id int64, req *admindto.UpdateNamedRequest) (*admindto.NamedResponse, error) {
	return &admindto.NamedResponse{ID: uint64(id)}, nil
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
	return &admindto.PlaySourceResponse{ID: uint64(id)}, nil
}
func (s adminPlaySvc) DeleteSource(ctx context.Context, id int64) error { return nil }
func (s adminPlaySvc) ListEpisodes(ctx context.Context, req *admindto.PlayEpisodeListRequest) ([]admindto.PlayEpisodeResponse, int, error) {
	return nil, 0, nil
}
func (s adminPlaySvc) CreateEpisode(ctx context.Context, req *admindto.CreatePlayEpisodeRequest) (*admindto.PlayEpisodeResponse, error) {
	return &admindto.PlayEpisodeResponse{}, nil
}
func (s adminPlaySvc) UpdateEpisode(ctx context.Context, id int64, req *admindto.UpdatePlayEpisodeRequest) (*admindto.PlayEpisodeResponse, error) {
	return &admindto.PlayEpisodeResponse{ID: uint64(id)}, nil
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
	return &shareddto.LiveChannelItem{ID: uint64(id)}, nil
}
func (s adminLiveSvc) Delete(ctx context.Context, id int64) error { return nil }
func (s adminLiveSvc) SyncFromSource(ctx context.Context) (*shareddto.LiveSyncResult, error) {
	return &shareddto.LiveSyncResult{}, nil
}

type adminCollectSvc struct{}

func (s adminCollectSvc) ListSources(ctx context.Context, req *admindto.CollectSourceListRequest) ([]shareddto.CollectSourceItem, int, error) {
	return nil, 0, nil
}
func (s adminCollectSvc) CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*shareddto.CollectSourceItem, error) {
	return &shareddto.CollectSourceItem{}, nil
}
func (s adminCollectSvc) UpdateSource(ctx context.Context, id int64, req *admindto.UpdateCollectSourceRequest) (*shareddto.CollectSourceItem, error) {
	return &shareddto.CollectSourceItem{ID: uint64(id)}, nil
}
func (s adminCollectSvc) DeleteSource(ctx context.Context, id int64) error { return nil }
func (s adminCollectSvc) ListCategories(ctx context.Context, sourceID int64) ([]shareddto.CollectCategoryMapItem, error) {
	return nil, nil
}
func (s adminCollectSvc) SetCategories(ctx context.Context, sourceID int64, req *admindto.SetCollectCategoriesRequest) ([]shareddto.CollectCategoryMapItem, error) {
	return nil, nil
}
func (s adminCollectSvc) ListLogs(ctx context.Context, req *admindto.CollectLogListRequest) ([]shareddto.CollectLogItem, int, error) {
	return nil, 0, nil
}
func (s adminCollectSvc) FetchRemoteCategories(ctx context.Context, sourceID int64) (*admindto.RemoteCategoryResponse, error) {
	return &admindto.RemoteCategoryResponse{}, nil
}
func (s adminCollectSvc) EnableSchedule(ctx context.Context, sourceID int64) error  { return nil }
func (s adminCollectSvc) DisableSchedule(ctx context.Context, sourceID int64) error { return nil }
func (s adminCollectSvc) CollectNow(ctx context.Context, sourceID int64, req *admindto.CollectNowRequest) error {
	return nil
}
func (s adminCollectSvc) ReloadScheduler(ctx context.Context) error { return nil }
func (s adminCollectSvc) StartScheduler(ctx context.Context) error  { return nil }
func (s adminCollectSvc) StopScheduler(ctx context.Context) error   { return nil }

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

type adminLogSvc struct{}

func (s adminLogSvc) ListSystemLogs(ctx context.Context, req *admindto.SystemLogListRequest) ([]admindto.SystemLogItem, int, error) {
	return nil, 0, nil
}
func (s adminLogSvc) ListLoginLogs(ctx context.Context, req *admindto.LoginLogListRequest) ([]admindto.LoginLogItem, int, error) {
	return nil, 0, nil
}
func (s adminLogSvc) ListAppLogs(ctx context.Context, req *admindto.AppLogListRequest) (*admindto.AppLogListResponse, error) {
	return &admindto.AppLogListResponse{List: []admindto.AppLogItem{}, HasMore: false}, nil
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
	return &shareddto.VideoDetailResponse{ID: uint64(id)}, nil
}
func (s clientVideoSvc) Related(ctx context.Context, id int64, limit int) ([]shareddto.VideoListItem, error) {
	return nil, nil
}

type clientLiveSvc struct{}

func (s clientLiveSvc) List(ctx context.Context, req *clientdto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error) {
	return nil, 0, nil
}

func (s clientLiveSvc) GetStreamURL(ctx context.Context, id int64) (string, error) {
	return "", nil
}

func (s clientLiveSvc) AllowedStreamDomains(ctx context.Context) (map[string]struct{}, error) {
	return nil, nil
}

type clientLiveProxySvc struct{}

func (s clientLiveProxySvc) Proxy(c *gin.Context, channelID int64, segURL string) error { return nil }

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

// ===== Phase 5 stubs =====

type adminDataSvc struct{}

func (s adminDataSvc) Backup(ctx context.Context, w io.Writer, useNative bool) error { return nil }
func (s adminDataSvc) BatchUpdatePreview(ctx context.Context, req *admindto.BatchUpdatePreviewRequest) (int64, error) {
	return 0, nil
}
func (s adminDataSvc) BatchUpdateExecute(ctx context.Context, req *admindto.BatchUpdateExecuteRequest, adminID int64, ip string) (int64, error) {
	return 0, nil
}

type adminMgmtSvc struct{}

func (s adminMgmtSvc) Dashboard(ctx context.Context) (*admindto.DashboardResponse, error) {
	return &admindto.DashboardResponse{}, nil
}
func (s adminMgmtSvc) BatchUpdatePublishStatus(ctx context.Context, req *admindto.BatchVideoRequest) (*admindto.BatchVideoResponse, error) {
	return &admindto.BatchVideoResponse{}, nil
}
func (s adminMgmtSvc) BatchDeleteVideos(ctx context.Context, req *admindto.BatchVideoRequest) (*admindto.BatchVideoResponse, error) {
	return &admindto.BatchVideoResponse{}, nil
}
func (s adminMgmtSvc) ListAdmins(ctx context.Context, req *admindto.AdminListRequest) ([]admindto.AdminItem, int, error) {
	return nil, 0, nil
}
func (s adminMgmtSvc) CreateAdmin(ctx context.Context, req *admindto.CreateAdminRequest) (*admindto.AdminItem, error) {
	return &admindto.AdminItem{}, nil
}
func (s adminMgmtSvc) UpdateAdmin(ctx context.Context, id int64, req *admindto.UpdateAdminRequest) (*admindto.AdminItem, error) {
	return &admindto.AdminItem{ID: uint64(id)}, nil
}
func (s adminMgmtSvc) ResetAdminPassword(ctx context.Context, id int64, req *admindto.ResetAdminPasswordRequest) error {
	return nil
}
func (s adminMgmtSvc) DeleteAdmin(ctx context.Context, id int64) error { return nil }
func (s adminMgmtSvc) ListGroups(ctx context.Context, req *admindto.UserGroupListRequest) ([]admindto.UserGroupItem, int, error) {
	return nil, 0, nil
}
func (s adminMgmtSvc) CreateGroup(ctx context.Context, req *admindto.CreateUserGroupRequest) (*admindto.UserGroupItem, error) {
	return &admindto.UserGroupItem{}, nil
}
func (s adminMgmtSvc) UpdateGroup(ctx context.Context, id int64, req *admindto.UpdateUserGroupRequest) (*admindto.UserGroupItem, error) {
	return &admindto.UserGroupItem{ID: uint64(id)}, nil
}
func (s adminMgmtSvc) DeleteGroup(ctx context.Context, id int64) error { return nil }
func (s adminMgmtSvc) ListUsers(ctx context.Context, req *admindto.UserListRequest) ([]admindto.UserItem, int, error) {
	return nil, 0, nil
}
func (s adminMgmtSvc) CreateUser(ctx context.Context, req *admindto.CreateUserRequest) (*admindto.UserItem, error) {
	return &admindto.UserItem{}, nil
}
func (s adminMgmtSvc) UpdateUser(ctx context.Context, id int64, req *admindto.UpdateUserRequest) (*admindto.UserItem, error) {
	return &admindto.UserItem{ID: uint64(id)}, nil
}
func (s adminMgmtSvc) ResetUserPassword(ctx context.Context, id int64, req *admindto.ResetUserPasswordRequest) error {
	return nil
}
func (s adminMgmtSvc) DeleteUser(ctx context.Context, id int64) error { return nil }
func (s adminMgmtSvc) ListUserLoginLogs(ctx context.Context, userID int64, offset, limit int) ([]model.UserLoginLogs, int, error) {
	return nil, 0, nil
}
func (s adminMgmtSvc) ListBanners(ctx context.Context, offset, limit int) ([]admindto.BannerItem, int, error) {
	return nil, 0, nil
}
func (s adminMgmtSvc) CreateBanner(ctx context.Context, req *admindto.CreateBannerRequest) (*admindto.BannerItem, error) {
	return &admindto.BannerItem{}, nil
}
func (s adminMgmtSvc) UpdateBanner(ctx context.Context, id int64, req *admindto.UpdateBannerRequest) (*admindto.BannerItem, error) {
	return &admindto.BannerItem{ID: uint64(id)}, nil
}
func (s adminMgmtSvc) DeleteBanner(ctx context.Context, id int64) error { return nil }

type clientUserSvc struct{}

func (s clientUserSvc) Register(ctx context.Context, req *clientdto.RegisterRequest) (*clientdto.Profile, error) {
	return &clientdto.Profile{}, nil
}
func (s clientUserSvc) Login(ctx context.Context, req *clientdto.LoginRequest, ip, ua string) (*clientdto.LoginResponse, error) {
	return &clientdto.LoginResponse{}, nil
}
func (s clientUserSvc) Profile(ctx context.Context, userID int64) (*clientdto.Profile, error) {
	return &clientdto.Profile{ID: uint64(userID)}, nil
}
func (s clientUserSvc) ListFavorites(ctx context.Context, userID int64, req *clientdto.FavoriteListRequest) ([]clientdto.FavoriteItem, int, error) {
	return nil, 0, nil
}
func (s clientUserSvc) AddFavorite(ctx context.Context, userID, videoID int64) error    { return nil }
func (s clientUserSvc) RemoveFavorite(ctx context.Context, userID, videoID int64) error { return nil }
func (s clientUserSvc) ListHistory(ctx context.Context, userID int64, req *clientdto.HistoryListRequest) ([]clientdto.HistoryItem, int, error) {
	return nil, 0, nil
}
func (s clientUserSvc) UpsertHistory(ctx context.Context, userID int64, req *clientdto.UpsertHistoryRequest) error {
	return nil
}
func (s clientUserSvc) DeleteHistory(ctx context.Context, userID, videoID int64) error { return nil }
func (s clientUserSvc) ClearHistory(ctx context.Context, userID int64) error           { return nil }
func (s clientUserSvc) ListComments(ctx context.Context, videoID int64, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error) {
	return nil, 0, nil
}
func (s clientUserSvc) CreateComment(ctx context.Context, userID int64, req *clientdto.CreateCommentRequest) (*clientdto.CommentItem, error) {
	return &clientdto.CommentItem{}, nil
}
func (s clientUserSvc) DeleteComment(ctx context.Context, userID, commentID int64) error { return nil }

type clientBannerSvc struct{}

func (s clientBannerSvc) ListBanners(ctx context.Context) ([]clientdto.BannerItem, error) {
	return nil, nil
}

// NewBusinessHandlers builds no-op business handlers for tests.
func NewBusinessHandlers() BusinessHandlers {
	auth := authSvc{}
	mgmt := adminMgmtSvc{}
	userSvc := clientUserSvc{}
	bannerSvc := clientBannerSvc{}
	clientLiveSvc := clientLiveSvc{}
	return BusinessHandlers{
		AuthService:    auth,
		AdminAuth:      adminhandler.NewAuthHandler(auth),
		AdminCategory:  adminhandler.NewCategoryHandler(adminCategorySvc{}),
		AdminVideo:     adminhandler.NewVideoHandler(adminVideoSvc{}),
		AdminMetadata:  adminhandler.NewMetadataHandler(adminMetadataSvc{}),
		AdminPlay:      adminhandler.NewPlayHandler(adminPlaySvc{}),
		AdminLive:      adminhandler.NewLiveHandler(adminLiveSvc{}),
		AdminCollect:   adminhandler.NewCollectHandler(adminCollectSvc{}),
		AdminSettings:  adminhandler.NewSettingsHandler(adminSettingsSvc{}, nil),
		AdminLog:       adminhandler.NewLogHandler(adminLogSvc{}),
		AdminMgmt:      adminhandler.NewManagementHandler(mgmt),
		AdminData:      adminhandler.NewDataHandler(adminDataSvc{}),
		ClientCategory: clienthandler.NewCategoryHandler(clientCategorySvc{}),
		ClientVideo:    clienthandler.NewVideoHandler(clientVideoSvc{}),
		ClientLive:     clienthandler.NewLiveHandler(clientLiveSvc, clientLiveProxySvc{}),
		ClientSite:     clienthandler.NewSiteHandler(adminSettingsSvc{}),
		ClientUser:     clienthandler.NewUserHandler(userSvc),
		ClientBanner:   clienthandler.NewBannerHandler(bannerSvc),
		OpenResource:   openhandler.NewResourceHandler(openResourceSvc{}),
	}
}
