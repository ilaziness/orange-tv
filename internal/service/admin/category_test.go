package admin

import (
	"context"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/require"
)

type fakeCategoryRepo struct {
	items   map[uint64]*model.Categories
	nextID  uint64
	videos  map[int64]int
	created []*model.Categories
}

func newFakeCategoryRepo() *fakeCategoryRepo {
	return &fakeCategoryRepo{items: map[uint64]*model.Categories{}, nextID: 1, videos: map[int64]int{}}
}

func (f *fakeCategoryRepo) List(ctx context.Context, onlyEnabled bool) ([]model.Categories, error) {
	out := make([]model.Categories, 0, len(f.items))
	for _, item := range f.items {
		if item.DeletedAt != nil {
			continue
		}
		if onlyEnabled && item.Status != constant.StatusEnabled {
			continue
		}
		out = append(out, *item)
	}
	return out, nil
}

func (f *fakeCategoryRepo) GetByID(ctx context.Context, id int64) (*model.Categories, error) {
	item, ok := f.items[uint64(id)]
	if !ok || item.DeletedAt != nil {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}

func (f *fakeCategoryRepo) ExistsName(ctx context.Context, name string, excludeID int64) (bool, error) {
	for id, item := range f.items {
		if item.DeletedAt != nil || id == uint64(excludeID) {
			continue
		}
		if item.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeCategoryRepo) Create(ctx context.Context, c *model.Categories) error {
	c.ID = f.nextID
	f.nextID++
	cp := *c
	f.items[c.ID] = &cp
	f.created = append(f.created, c)
	return nil
}

func (f *fakeCategoryRepo) Update(ctx context.Context, c *model.Categories) error {
	cp := *c
	f.items[c.ID] = &cp
	return nil
}

func (f *fakeCategoryRepo) SoftDelete(ctx context.Context, id int64) error {
	if item, ok := f.items[uint64(id)]; ok {
		t := time.Now()
		item.DeletedAt = &t
	}
	return nil
}

func (f *fakeCategoryRepo) CountChildren(ctx context.Context, parentID int64) (int, error) {
	n := 0
	for _, item := range f.items {
		if item.DeletedAt == nil && item.ParentID == uint64(parentID) {
			n++
		}
	}
	return n, nil
}

func (f *fakeCategoryRepo) CountVideos(ctx context.Context, categoryID int64) (int, error) {
	return f.videos[categoryID], nil
}

func (f *fakeCategoryRepo) ListIDs(ctx context.Context) ([]model.Categories, error) {
	return f.List(ctx, false)
}

func TestCategoryService_CreateRejectsDuplicateName(t *testing.T) {
	repo := newFakeCategoryRepo()
	repo.items[1] = &model.Categories{ID: 1, Name: "电影", Status: 1}
	svc := NewCategoryService(repo, nil)

	_, err := svc.Create(context.Background(), &dto.CreateCategoryRequest{Name: "电影"})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.CategoryNameExists.Code, code.Code)
}

func TestCategoryService_DeleteRejectsWhenHasVideos(t *testing.T) {
	repo := newFakeCategoryRepo()
	repo.items[1] = &model.Categories{ID: 1, Name: "电影", Status: 1}
	repo.videos[1] = 2
	svc := NewCategoryService(repo, nil)

	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.CategoryHasVideos.Code, code.Code)
}

func TestCategoryService_UpdateRejectsCycle(t *testing.T) {
	repo := newFakeCategoryRepo()
	repo.items[1] = &model.Categories{ID: 1, Name: "电影", ParentID: 0, Status: 1}
	repo.items[2] = &model.Categories{ID: 2, Name: "动作", ParentID: 1, Status: 1}
	svc := NewCategoryService(repo, nil)

	parent := uint64(2)
	_, err := svc.Update(context.Background(), 1, &dto.UpdateCategoryRequest{ParentID: &parent})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.CategoryCycle.Code, code.Code)
}

func TestCategoryService_ListTree(t *testing.T) {
	repo := newFakeCategoryRepo()
	repo.items[1] = &model.Categories{ID: 1, Name: "电影", ParentID: 0, Status: 1}
	repo.items[2] = &model.Categories{ID: 2, Name: "动作", ParentID: 1, Status: 1}
	svc := NewCategoryService(repo, nil)

	tree, err := svc.ListTree(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, tree, 1)
	require.Equal(t, "电影", tree[0].Name)
	require.Len(t, tree[0].Children, 1)
	require.Equal(t, "动作", tree[0].Children[0].Name)
}
