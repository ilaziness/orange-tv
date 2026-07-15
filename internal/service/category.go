package service

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// CategoryService manages categories.
type CategoryService interface {
	ListTree(ctx context.Context, onlyEnabled bool) ([]dto.CategoryResponse, error)
	Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	Update(ctx context.Context, id int64, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	Delete(ctx context.Context, id int64) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService creates a CategoryService.
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) ListTree(ctx context.Context, onlyEnabled bool) ([]dto.CategoryResponse, error) {
	items, err := s.repo.List(ctx, onlyEnabled)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return buildCategoryTree(items), nil
}

func (s *categoryService) Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "分类名称不能为空")
	}
	exists, err := s.repo.ExistsName(ctx, name, 0)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.CategoryNameExists
	}
	if req.ParentID > 0 {
		parent, err := s.repo.GetByID(ctx, req.ParentID)
		if err != nil {
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
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toCategoryDTO(item), nil
}

func (s *categoryService) Update(ctx context.Context, id int64, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
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
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
			if parent == nil {
				return nil, errcode.CategoryNotFound
			}
			// prevent cycle: parent cannot be a descendant of id
			all, err := s.repo.List(ctx, false)
			if err != nil {
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
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toCategoryDTO(item), nil
}

func (s *categoryService) Delete(ctx context.Context, id int64) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return errcode.CategoryNotFound
	}
	children, err := s.repo.CountChildren(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if children > 0 {
		return errcode.CategoryHasChildren
	}
	videos, err := s.repo.CountVideos(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if videos > 0 {
		return errcode.CategoryHasVideos
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
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

func buildCategoryTree(items []model.Categories) []dto.CategoryResponse {
	byParent := make(map[int64][]model.Categories, len(items))
	for _, item := range items {
		byParent[item.ParentID] = append(byParent[item.ParentID], item)
	}
	var build func(parentID int64) []dto.CategoryResponse
	build = func(parentID int64) []dto.CategoryResponse {
		children := byParent[parentID]
		out := make([]dto.CategoryResponse, 0, len(children))
		for _, c := range children {
			node := dto.CategoryResponse{
				ID:        c.ID,
				Name:      c.Name,
				ParentID:  c.ParentID,
				SortOrder: c.SortOrder,
				Status:    c.Status,
				Children:  build(c.ID),
			}
			out = append(out, node)
		}
		return out
	}
	return build(0)
}

func isDescendant(all []model.Categories, ancestorID, nodeID int64) bool {
	// walk from nodeID up via parent map
	parentOf := make(map[int64]int64, len(all))
	for _, c := range all {
		parentOf[c.ID] = c.ParentID
	}
	// alternatively: collect descendants of ancestor
	childrenOf := make(map[int64][]int64, len(all))
	for _, c := range all {
		childrenOf[c.ParentID] = append(childrenOf[c.ParentID], c.ID)
	}
	stack := []int64{ancestorID}
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
