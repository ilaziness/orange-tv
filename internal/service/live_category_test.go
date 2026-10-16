package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// liveCategoryStub implements only the category reads used by pair validation.
type liveCategoryStub struct {
	repository.LiveRepository
	byID   map[uint32]*model.LiveCategories
	getErr error
}

func (s liveCategoryStub) GetCategory(_ context.Context, id uint32) (*model.LiveCategories, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	c, ok := s.byID[id]
	if !ok || c == nil || c.DeletedAt != nil {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (s liveCategoryStub) CountEnabledCategoryChildren(_ context.Context, parentID uint32) (int, error) {
	if s.getErr != nil {
		return 0, s.getErr
	}
	n := 0
	for _, c := range s.byID {
		if c != nil && c.DeletedAt == nil && c.ParentID == parentID && c.Status == constant.StatusEnabled {
			n++
		}
	}
	return n, nil
}

func assertLiveCode(t *testing.T, err error, code int, message string) {
	t.Helper()
	got, ok := errcode.As(err)
	require.True(t, ok)
	assert.Equal(t, code, got.Code)
	assert.Equal(t, message, got.Message)
}

func TestValidateLiveCategoryPair(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deleted := time.Now()
	repo := liveCategoryStub{byID: map[uint32]*model.LiveCategories{
		1: {ID: 1, ParentID: 0, Name: "游戏", Status: constant.StatusEnabled},
		2: {ID: 2, ParentID: 0, Name: "闲置", Status: constant.StatusDisabled},
		3: {ID: 3, ParentID: 1, Name: "MOBA", Status: constant.StatusEnabled},
		4: {ID: 4, ParentID: 1, Name: "停用子类", Status: constant.StatusDisabled},
		5: {ID: 5, ParentID: 2, Name: "孤儿", Status: constant.StatusEnabled},
		6: {ID: 6, ParentID: 0, Name: "已删", Status: constant.StatusEnabled, DeletedAt: &deleted},
		7: {ID: 7, ParentID: 0, Name: "无子类", Status: constant.StatusEnabled},
	}}

	require.NoError(t, ValidateLiveCategoryPair(ctx, repo, 7, 0))
	require.NoError(t, ValidateLiveCategoryPair(ctx, repo, 1, 3))

	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 0, 0), errcode.ParamError.Code, "请选择直播分类")
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 99, 0), errcode.LiveCategoryDisabled.Code, errcode.LiveCategoryDisabled.Message)
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 6, 0), errcode.LiveCategoryDisabled.Code, errcode.LiveCategoryDisabled.Message)
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 2, 0), errcode.LiveCategoryDisabled.Code, errcode.LiveCategoryDisabled.Message)
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 3, 0), errcode.LiveCategoryDisabled.Code, errcode.LiveCategoryDisabled.Message)
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 7, 3), errcode.ParamError.Code, "该一级分类无启用子类，二级分类须为0")
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 1, 0), errcode.ParamError.Code, "请选择二级分类")
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 1, 4), errcode.LiveCategoryDisabled.Code, "二级分类不可用或不属于所选一级分类")
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 1, 5), errcode.LiveCategoryDisabled.Code, "二级分类不可用或不属于所选一级分类")
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, repo, 1, 99), errcode.LiveCategoryDisabled.Code, "二级分类不可用或不属于所选一级分类")

	dbErr := errors.New("db down")
	assertLiveCode(t, ValidateLiveCategoryPair(ctx, liveCategoryStub{getErr: dbErr}, 1, 0), errcode.DatabaseError.Code, errcode.DatabaseError.Message)
}

func TestResolveLiveCategoryOverride(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := liveCategoryStub{byID: map[uint32]*model.LiveCategories{
		1: {ID: 1, ParentID: 0, Status: constant.StatusEnabled},
		3: {ID: 3, ParentID: 1, Status: constant.StatusEnabled},
		7: {ID: 7, ParentID: 0, Status: constant.StatusEnabled},
	}}

	catID, subID, err := ResolveLiveCategoryOverride(ctx, liveCategoryStub{}, nil, nil, 5, 6)
	require.NoError(t, err)
	assert.Equal(t, uint32(5), catID)
	assert.Equal(t, uint32(6), subID)

	onlySub := uint32(3)
	_, _, err = ResolveLiveCategoryOverride(ctx, repo, nil, &onlySub, 1, 0)
	assertLiveCode(t, err, errcode.ParamError.Code, "请选择一级分类")

	onlyL1 := uint32(7)
	catID, subID, err = ResolveLiveCategoryOverride(ctx, repo, &onlyL1, nil, 1, 3)
	require.NoError(t, err)
	assert.Equal(t, uint32(7), catID)
	assert.Equal(t, uint32(0), subID)

	needsChild := uint32(1)
	_, _, err = ResolveLiveCategoryOverride(ctx, repo, &needsChild, nil, 7, 0)
	assertLiveCode(t, err, errcode.ParamError.Code, "请选择二级分类")

	sub := uint32(3)
	catID, subID, err = ResolveLiveCategoryOverride(ctx, repo, &needsChild, &sub, 7, 0)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), catID)
	assert.Equal(t, uint32(3), subID)

	zeroSub := uint32(0)
	_, _, err = ResolveLiveCategoryOverride(ctx, repo, &needsChild, &zeroSub, 7, 0)
	assertLiveCode(t, err, errcode.ParamError.Code, "请选择二级分类")
}

func TestNewLiveSessionSnapshotCopiesSubCategory(t *testing.T) {
	t.Parallel()
	now := time.Now()
	session := NewLiveSessionSnapshot(
		&model.LiveRooms{ID: 1, UserID: 2, Title: "t", CategoryID: 3, SubCategoryID: 4},
		&model.Users{Nickname: "n", Avatar: "a"},
		now,
	)
	assert.Equal(t, uint32(3), session.CategoryID)
	assert.Equal(t, uint32(4), session.SubCategoryID)
	assert.Equal(t, now, session.StartedAt)
}
