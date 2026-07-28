package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"github.com/ilaziness/orange-tv/pkg/cache"
	"go.uber.org/zap"
)

// CategoryService manages categories.
type CategoryService interface {
	ListTree(ctx context.Context, onlyEnabled bool) ([]shareddto.CategoryResponse, error)
	Create(ctx context.Context, req *dto.CreateCategoryRequest) (*shareddto.CategoryResponse, error)
	Update(ctx context.Context, id int64, req *dto.UpdateCategoryRequest) (*shareddto.CategoryResponse, error)
	Delete(ctx context.Context, id int64) error
}

type categoryService struct {
	repo  repository.CategoryRepository
	cache cache.Cache
	log   *zap.Logger
}

// NewCategoryService creates a CategoryService.
func NewCategoryService(repo repository.CategoryRepository, c cache.Cache, log *zap.Logger) CategoryService {
	if c == nil {
		c = cache.NewNopCache()
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &categoryService{repo: repo, cache: c, log: log}
}

func (s *categoryService) ListTree(ctx context.Context, onlyEnabled bool) ([]shareddto.CategoryResponse, error) {
	items, err := s.repo.List(ctx, onlyEnabled)
	if err != nil {
		s.log.Error("category: list failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return utils.BuildCategoryTree(items), nil
}

func (s *categoryService) Create(ctx context.Context, req *dto.CreateCategoryRequest) (*shareddto.CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "分类名称不能为空")
	}
	exists, err := s.repo.ExistsName(ctx, name, 0)
	if err != nil {
		s.log.Error("category: check name exists failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.CategoryNameExists
	}
	if req.ParentID > 0 {
		parent, err := s.repo.GetByID(ctx, int64(req.ParentID))
		if err != nil {
			s.log.Error("category: get parent by id failed", zap.Uint64("parent_id", req.ParentID), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if parent == nil {
			return nil, errcode.CategoryNotFound
		}
	}

	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	item := &model.Categories{
		Name:      name,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		Status:    status,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		s.log.Error("category: create failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.invalidateCaches(ctx)
	return toCategoryDTO(item), nil
}

func (s *categoryService) Update(ctx context.Context, id int64, req *dto.UpdateCategoryRequest) (*shareddto.CategoryResponse, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("category: get by id failed", zap.Int64("category_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return nil, errcode.CategoryNotFound
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "分类名称不能为空")
		}
		exists, err := s.repo.ExistsName(ctx, name, id)
		if err != nil {
			s.log.Error("category: check name exists for update failed", zap.Int64("category_id", id), zap.String("name", name), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if exists {
			return nil, errcode.CategoryNameExists
		}
		item.Name = name
	}
	if req.ParentID != nil {
		parentID := *req.ParentID
		if parentID == uint64(id) {
			return nil, errcode.CategoryCycle
		}
		if parentID > 0 {
			parent, err := s.repo.GetByID(ctx, int64(parentID))
			if err != nil {
				s.log.Error("category: get parent for update failed", zap.Int64("category_id", id), zap.Uint64("parent_id", parentID), zap.Error(err))
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
			if parent == nil {
				return nil, errcode.CategoryNotFound
			}
			// prevent cycle: new parent cannot be a descendant of current node
			all, err := s.repo.List(ctx, false)
			if err != nil {
				s.log.Error("category: list for cycle check failed", zap.Int64("category_id", id), zap.Error(err))
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
			if isDescendant(all, uint64(id), parentID) {
				return nil, errcode.CategoryCycle
			}
		}
		item.ParentID = parentID
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := s.repo.Update(ctx, item); err != nil {
		s.log.Error("category: update failed", zap.Int64("category_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.invalidateCaches(ctx)
	return toCategoryDTO(item), nil
}

func (s *categoryService) Delete(ctx context.Context, id int64) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("category: get by id for delete failed", zap.Int64("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return errcode.CategoryNotFound
	}
	children, err := s.repo.CountChildren(ctx, id)
	if err != nil {
		s.log.Error("category: count children failed", zap.Int64("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if children > 0 {
		return errcode.CategoryHasChildren
	}
	videos, err := s.repo.CountVideos(ctx, id)
	if err != nil {
		s.log.Error("category: count videos failed", zap.Int64("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if videos > 0 {
		return errcode.CategoryHasVideos
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		s.log.Error("category: soft delete failed", zap.Int64("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	s.invalidateCaches(ctx)
	return nil
}

func (s *categoryService) invalidateCaches(ctx context.Context) {
	_ = s.cache.Delete(ctx, "category:tree:client")
	_ = s.cache.Delete(ctx, "open:categories")
}

func toCategoryDTO(item *model.Categories) *shareddto.CategoryResponse {
	return &shareddto.CategoryResponse{
		ID:        item.ID,
		Name:      item.Name,
		ParentID:  item.ParentID,
		SortOrder: item.SortOrder,
		Status:    item.Status,
	}
}

func isDescendant(all []model.Categories, ancestorID, nodeID uint64) bool {
	childrenOf := make(map[uint64][]uint64, len(all))
	for _, c := range all {
		childrenOf[c.ParentID] = append(childrenOf[c.ParentID], c.ID)
	}
	stack := []uint64{ancestorID}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, child := range childrenOf[cur] {
			if child == nodeID {
				return true
			}
			stack = append(stack, child)
		}
	}
	return false
}
