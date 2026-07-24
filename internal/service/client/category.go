package client

import (
	"context"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

const categoryTreeCacheKey = "category:tree:client"
const categoryTreeCacheTTL = 5 * time.Minute

// CategoryService provides client category queries.
type CategoryService interface {
	ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error)
}

type categoryService struct {
	repo  repository.CategoryRepository
	cache cache.Cache
	log   *zap.Logger
}

// NewCategoryService creates a client CategoryService.
func NewCategoryService(repo repository.CategoryRepository, c cache.Cache, log *zap.Logger) CategoryService {
	if c == nil {
		c = cache.NewNopCache()
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &categoryService{repo: repo, cache: c, log: log}
}

func (s *categoryService) ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	if v, err := s.cache.Get(ctx, categoryTreeCacheKey); err == nil {
		if tree, ok := v.([]shareddto.CategoryResponse); ok {
			return tree, nil
		}
	}
	items, err := s.repo.List(ctx, true)
	if err != nil {
		s.log.Error("client category: list tree failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tree := buildCategoryTree(items)
	_ = s.cache.Set(ctx, categoryTreeCacheKey, tree, categoryTreeCacheTTL)
	return tree, nil
}

// InvalidateCategoryCache clears client category tree cache (call after admin category writes).
func InvalidateCategoryCache(ctx context.Context, c cache.Cache) {
	if c == nil {
		return
	}
	_ = c.Delete(ctx, categoryTreeCacheKey)
}

func buildCategoryTree(items []model.Categories) []shareddto.CategoryResponse {
	byParent := make(map[uint64][]model.Categories, len(items))
	for _, item := range items {
		byParent[item.ParentID] = append(byParent[item.ParentID], item)
	}
	var build func(parentID uint64) []shareddto.CategoryResponse
	build = func(parentID uint64) []shareddto.CategoryResponse {
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
