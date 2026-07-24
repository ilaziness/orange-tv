package admin

import (
	"context"
	"strings"

	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// MetadataService manages directors/actors/tags.
type MetadataService interface {
	ListDirectors(ctx context.Context, req *dto.NameSearchRequest) ([]dto.NamedResponse, int, error)
	CreateDirector(ctx context.Context, req *dto.CreateNamedRequest) (*dto.NamedResponse, error)
	UpdateDirector(ctx context.Context, id int64, req *dto.UpdateNamedRequest) (*dto.NamedResponse, error)
	DeleteDirector(ctx context.Context, id int64) error

	ListActors(ctx context.Context, req *dto.NameSearchRequest) ([]dto.NamedResponse, int, error)
	CreateActor(ctx context.Context, req *dto.CreateNamedRequest) (*dto.NamedResponse, error)
	UpdateActor(ctx context.Context, id int64, req *dto.UpdateNamedRequest) (*dto.NamedResponse, error)
	DeleteActor(ctx context.Context, id int64) error

	ListTags(ctx context.Context, req *dto.NameSearchRequest) ([]dto.NamedResponse, int, error)
	CreateTag(ctx context.Context, req *dto.CreateNamedRequest) (*dto.NamedResponse, error)
	UpdateTag(ctx context.Context, id int64, req *dto.UpdateNamedRequest) (*dto.NamedResponse, error)
	DeleteTag(ctx context.Context, id int64) error
}

type metadataService struct {
	repo repository.MetadataRepository
	log  *zap.Logger
}

// NewMetadataService creates a MetadataService.
func NewMetadataService(repo repository.MetadataRepository, log *zap.Logger) MetadataService {
	if log == nil {
		log = zap.NewNop()
	}
	return &metadataService{repo: repo, log: log}
}

func (s *metadataService) ListDirectors(ctx context.Context, req *dto.NameSearchRequest) ([]dto.NamedResponse, int, error) {
	items, total, err := s.repo.ListDirectors(ctx, strings.TrimSpace(req.Keyword), req.GetOffset(), req.GetLimit())
	if err != nil {
		s.log.Error("metadata: list directors failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapNamedDirectors(items), total, nil
}

func (s *metadataService) CreateDirector(ctx context.Context, req *dto.CreateNamedRequest) (*dto.NamedResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "名称不能为空")
	}
	exists, err := s.repo.ExistsDirectorName(ctx, name, 0)
	if err != nil {
		s.log.Error("metadata: check director name exists failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.DirectorNameExists
	}
	m := &model.Directors{Name: name}
	if err := s.repo.CreateDirector(ctx, m); err != nil {
		s.log.Error("metadata: create director failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.NamedResponse{ID: m.ID, Name: m.Name}, nil
}

func (s *metadataService) UpdateDirector(ctx context.Context, id int64, req *dto.UpdateNamedRequest) (*dto.NamedResponse, error) {
	m, err := s.repo.GetDirector(ctx, id)
	if err != nil {
		s.log.Error("metadata: get director failed", zap.Int64("director_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.DirectorNotFound
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "名称不能为空")
	}
	exists, err := s.repo.ExistsDirectorName(ctx, name, id)
	if err != nil {
		s.log.Error("metadata: check director name exists for update failed", zap.Int64("director_id", id), zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.DirectorNameExists
	}
	m.Name = name
	if err := s.repo.UpdateDirector(ctx, m); err != nil {
		s.log.Error("metadata: update director failed", zap.Int64("director_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.NamedResponse{ID: m.ID, Name: m.Name}, nil
}

func (s *metadataService) DeleteDirector(ctx context.Context, id int64) error {
	m, err := s.repo.GetDirector(ctx, id)
	if err != nil {
		s.log.Error("metadata: get director for delete failed", zap.Int64("director_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.DirectorNotFound
	}
	n, err := s.repo.CountDirectorRefs(ctx, id)
	if err != nil {
		s.log.Error("metadata: count director refs failed", zap.Int64("director_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if n > 0 {
		return errcode.DirectorInUse
	}
	if err := s.repo.SoftDeleteDirector(ctx, id); err != nil {
		s.log.Error("metadata: soft delete director failed", zap.Int64("director_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *metadataService) ListActors(ctx context.Context, req *dto.NameSearchRequest) ([]dto.NamedResponse, int, error) {
	items, total, err := s.repo.ListActors(ctx, strings.TrimSpace(req.Keyword), req.GetOffset(), req.GetLimit())
	if err != nil {
		s.log.Error("metadata: list actors failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.NamedResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.NamedResponse{ID: item.ID, Name: item.Name})
	}
	return out, total, nil
}

func (s *metadataService) CreateActor(ctx context.Context, req *dto.CreateNamedRequest) (*dto.NamedResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "名称不能为空")
	}
	exists, err := s.repo.ExistsActorName(ctx, name, 0)
	if err != nil {
		s.log.Error("metadata: check actor name exists failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.ActorNameExists
	}
	m := &model.Actors{Name: name}
	if err := s.repo.CreateActor(ctx, m); err != nil {
		s.log.Error("metadata: create actor failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.NamedResponse{ID: m.ID, Name: m.Name}, nil
}

func (s *metadataService) UpdateActor(ctx context.Context, id int64, req *dto.UpdateNamedRequest) (*dto.NamedResponse, error) {
	m, err := s.repo.GetActor(ctx, id)
	if err != nil {
		s.log.Error("metadata: get actor failed", zap.Int64("actor_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.ActorNotFound
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "名称不能为空")
	}
	exists, err := s.repo.ExistsActorName(ctx, name, id)
	if err != nil {
		s.log.Error("metadata: check actor name exists for update failed", zap.Int64("actor_id", id), zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.ActorNameExists
	}
	m.Name = name
	if err := s.repo.UpdateActor(ctx, m); err != nil {
		s.log.Error("metadata: update actor failed", zap.Int64("actor_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.NamedResponse{ID: m.ID, Name: m.Name}, nil
}

func (s *metadataService) DeleteActor(ctx context.Context, id int64) error {
	m, err := s.repo.GetActor(ctx, id)
	if err != nil {
		s.log.Error("metadata: get actor for delete failed", zap.Int64("actor_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.ActorNotFound
	}
	n, err := s.repo.CountActorRefs(ctx, id)
	if err != nil {
		s.log.Error("metadata: count actor refs failed", zap.Int64("actor_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if n > 0 {
		return errcode.ActorInUse
	}
	if err := s.repo.SoftDeleteActor(ctx, id); err != nil {
		s.log.Error("metadata: soft delete actor failed", zap.Int64("actor_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *metadataService) ListTags(ctx context.Context, req *dto.NameSearchRequest) ([]dto.NamedResponse, int, error) {
	items, total, err := s.repo.ListTags(ctx, strings.TrimSpace(req.Keyword), req.GetOffset(), req.GetLimit())
	if err != nil {
		s.log.Error("metadata: list tags failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.NamedResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.NamedResponse{ID: item.ID, Name: item.Name})
	}
	return out, total, nil
}

func (s *metadataService) CreateTag(ctx context.Context, req *dto.CreateNamedRequest) (*dto.NamedResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "名称不能为空")
	}
	exists, err := s.repo.ExistsTagName(ctx, name, 0)
	if err != nil {
		s.log.Error("metadata: check tag name exists failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.TagNameExists
	}
	m := &model.Tags{Name: name}
	if err := s.repo.CreateTag(ctx, m); err != nil {
		s.log.Error("metadata: create tag failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.NamedResponse{ID: m.ID, Name: m.Name}, nil
}

func (s *metadataService) UpdateTag(ctx context.Context, id int64, req *dto.UpdateNamedRequest) (*dto.NamedResponse, error) {
	m, err := s.repo.GetTag(ctx, id)
	if err != nil {
		s.log.Error("metadata: get tag failed", zap.Int64("tag_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.TagNotFound
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "名称不能为空")
	}
	exists, err := s.repo.ExistsTagName(ctx, name, id)
	if err != nil {
		s.log.Error("metadata: check tag name exists for update failed", zap.Int64("tag_id", id), zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.TagNameExists
	}
	m.Name = name
	if err := s.repo.UpdateTag(ctx, m); err != nil {
		s.log.Error("metadata: update tag failed", zap.Int64("tag_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.NamedResponse{ID: m.ID, Name: m.Name}, nil
}

func (s *metadataService) DeleteTag(ctx context.Context, id int64) error {
	m, err := s.repo.GetTag(ctx, id)
	if err != nil {
		s.log.Error("metadata: get tag for delete failed", zap.Int64("tag_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.TagNotFound
	}
	n, err := s.repo.CountTagRefs(ctx, id)
	if err != nil {
		s.log.Error("metadata: count tag refs failed", zap.Int64("tag_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if n > 0 {
		return errcode.TagInUse
	}
	if err := s.repo.SoftDeleteTag(ctx, id); err != nil {
		s.log.Error("metadata: soft delete tag failed", zap.Int64("tag_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func mapNamedDirectors(items []model.Directors) []dto.NamedResponse {
	out := make([]dto.NamedResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.NamedResponse{ID: item.ID, Name: item.Name})
	}
	return out
}
