package client

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/cache"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// CategoryService provides client category queries.
type CategoryService interface {
	ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error)
}

type categoryService struct {
	repo  repository.CategoryRepository
	cache *cache.Manager
	log   *zap.Logger
}

// NewCategoryService creates a client CategoryService.
func NewCategoryService(repo repository.CategoryRepository, c *cache.Manager, log *zap.Logger) CategoryService {
	if log == nil {
		log = zap.NewNop()
	}
	return &categoryService{repo: repo, cache: c, log: log}
}

func (s *categoryService) ListTree(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	if tree, err := s.cache.GetCategoryTreeClient(ctx); err == nil && tree != nil {
		return tree, nil
	}
	items, err := s.repo.List(ctx, true)
	if err != nil {
		s.log.Error("client category: list tree failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tree := utils.BuildCategoryTree(items)
	_ = s.cache.SetCategoryTreeClient(ctx, tree)
	return tree, nil
}
