// Package testutil provides shared test helpers.
package testutil

import (
	"context"

	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	adminhandler "github.com/ilaziness/orange-tv/internal/handler/http/admin"
	clienthandler "github.com/ilaziness/orange-tv/internal/handler/http/client"
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
	ClientCategory *clienthandler.CategoryHandler
	ClientVideo    *clienthandler.VideoHandler
}

type authSvc struct{}

func (s authSvc) Login(ctx context.Context, req *admindto.LoginRequest) (*admindto.LoginResponse, error) {
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
	return []admindto.PlaySourceResponse{}, nil
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

type clientCategorySvc struct{}

func (s clientCategorySvc) ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	return []shareddto.CategoryResponse{}, nil
}

type clientVideoSvc struct{}

func (s clientVideoSvc) List(ctx context.Context, req *clientdto.VideoListRequest) ([]shareddto.VideoListItem, int, error) {
	return []shareddto.VideoListItem{}, 0, nil
}
func (s clientVideoSvc) Search(ctx context.Context, req *clientdto.SearchRequest) ([]shareddto.VideoListItem, int, error) {
	return []shareddto.VideoListItem{}, 0, nil
}
func (s clientVideoSvc) Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error) {
	return &shareddto.VideoDetailResponse{ID: id}, nil
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
		ClientCategory: clienthandler.NewCategoryHandler(clientCategorySvc{}),
		ClientVideo:    clienthandler.NewVideoHandler(clientVideoSvc{}),
	}
}
