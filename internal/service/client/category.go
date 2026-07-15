package client

import (
	"context"

	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// CategoryService provides client category queries.
type CategoryService interface {
	ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error)
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService creates a client CategoryService.
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	items, err := s.repo.List(ctx, true)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return buildCategoryTree(items), nil
}

func buildCategoryTree(items []model.Categories) []shareddto.CategoryResponse {
	byParent := make(map[int64][]model.Categories, len(items))
	for _, item := range items {
		byParent[item.ParentID] = append(byParent[item.ParentID], item)
	}
	var build func(parentID int64) []shareddto.CategoryResponse
	build = func(parentID int64) []shareddto.CategoryResponse {
		children := byParent[parentID]
		out := make([]shareddto.CategoryResponse, 0, len(children))
		for _, c := range children {
			childNodes := build(c.ID)
			if childNodes == nil {
				childNodes = []shareddto.CategoryResponse{}
			}
			out = append(out, shareddto.CategoryResponse{
				ID:        c.ID,
				Name:      c.Name,
				ParentID:  c.ParentID,
				SortOrder: c.SortOrder,
				Status:    c.Status,
				Children:  childNodes,
			})
		}
		return out
	}
	roots := build(0)
	if roots == nil {
		return []shareddto.CategoryResponse{}
	}
	return roots
}
