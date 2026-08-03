package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// CategoryService manages categories.
type CategoryService interface {
	ListTree(ctx context.Context, onlyEnabled bool) ([]dto.CategoryResponse, error)
	Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	Update(ctx context.Context, id uint32, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	Delete(ctx context.Context, id uint32) error
}

type categoryService struct {
	repo  repository.CategoryRepository
	cache *cache.Manager
	log   *zap.Logger
}

// NewCategoryService creates a CategoryService.
func NewCategoryService(repo repository.CategoryRepository, c *cache.Manager, log *zap.Logger) CategoryService {
	if log == nil {
		log = zap.NewNop()
	}
	return &categoryService{repo: repo, cache: c, log: log}
}

func (s *categoryService) ListTree(ctx context.Context, onlyEnabled bool) ([]dto.CategoryResponse, error) {
	items, err := s.repo.List(ctx, onlyEnabled)
	if err != nil {
		s.log.Error("category: list failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return utils.BuildCategoryTree(items, func(c model.Categories, children []dto.CategoryResponse) dto.CategoryResponse {
		return dto.CategoryResponse{
			ID:        c.ID,
			Name:      c.Name,
			ParentID:  c.ParentID,
			SortOrder: c.SortOrder,
			Status:    c.Status,
			Children:  children,
		}
	}), nil
}

func (s *categoryService) Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
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
		parent, err := s.repo.GetByID(ctx, req.ParentID)
		if err != nil {
			s.log.Error("category: get parent by id failed", zap.Uint32("parent_id", req.ParentID), zap.Error(err))
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
	s.cache.InvalidateCategory(ctx)
	return toCategoryDTO(item), nil
}

func (s *categoryService) Update(ctx context.Context, id uint32, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("category: get by id failed", zap.Uint32("category_id", id), zap.Error(err))
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
			s.log.Error("category: check name exists for update failed", zap.Uint32("category_id", id), zap.String("name", name), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if exists {
			return nil, errcode.CategoryNameExists
		}
		item.Name = name
	}
	if req.ParentID != nil {
		parentID := *req.ParentID
		if parentID == id {
			return nil, errcode.CategoryCycle
		}
		if parentID > 0 {
			parent, err := s.repo.GetByID(ctx, parentID)
			if err != nil {
				s.log.Error("category: get parent for update failed", zap.Uint32("category_id", id), zap.Uint32("parent_id", parentID), zap.Error(err))
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
			if parent == nil {
				return nil, errcode.CategoryNotFound
			}
			// prevent cycle: new parent cannot be a descendant of current node
			all, err := s.repo.List(ctx, false)
			if err != nil {
				s.log.Error("category: list for cycle check failed", zap.Uint32("category_id", id), zap.Error(err))
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
			if isDescendant(all, id, parentID) {
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
		s.log.Error("category: update failed", zap.Uint32("category_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateCategory(ctx)
	return toCategoryDTO(item), nil
}

func (s *categoryService) Delete(ctx context.Context, id uint32) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("category: get by id for delete failed", zap.Uint32("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return errcode.CategoryNotFound
	}
	children, err := s.repo.CountChildren(ctx, id)
	if err != nil {
		s.log.Error("category: count children failed", zap.Uint32("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if children > 0 {
		return errcode.CategoryHasChildren
	}
	videos, err := s.repo.CountVideos(ctx, id)
	if err != nil {
		s.log.Error("category: count videos failed", zap.Uint32("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if videos > 0 {
		return errcode.CategoryHasVideos
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		s.log.Error("category: soft delete failed", zap.Uint32("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateCategory(ctx)
	return nil
}

func toCategoryDTO(item *model.Categories) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:        item.ID,
		Name:      item.Name,
		ParentID:  item.ParentID,
		SortOrder: item.SortOrder,
		Status:    item.Status,
	}
}

func isDescendant(all []model.Categories, ancestorID, nodeID uint32) bool {
	childrenOf := make(map[uint32][]uint32, len(all))
	for _, c := range all {
		childrenOf[c.ParentID] = append(childrenOf[c.ParentID], c.ID)
	}
	stack := []uint32{ancestorID}
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
