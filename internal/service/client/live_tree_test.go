package client

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildLiveCategoryTree(t *testing.T) {
	t.Parallel()
	items := []model.LiveCategories{
		{ID: 3, ParentID: 1, Name: "MOBA", Sort: 1},
		{ID: 9, ParentID: 8, Name: "孤儿", Sort: 1},
		{ID: 1, ParentID: 0, Name: "游戏", Sort: 1},
		{ID: 4, ParentID: 1, Name: "FPS", Sort: 2},
		{ID: 10, ParentID: 3, Name: "更深一层", Sort: 1},
		{ID: 2, ParentID: 0, Name: "音乐", Sort: 2},
	}

	tree := buildLiveCategoryTree(items)
	require.Len(t, tree, 2)
	assert.Equal(t, uint32(1), tree[0].ID)
	assert.Equal(t, uint32(0), tree[0].ParentID)
	require.Len(t, tree[0].Children, 2)
	assert.Equal(t, uint32(3), tree[0].Children[0].ID)
	assert.Equal(t, uint32(1), tree[0].Children[0].ParentID)
	assert.Empty(t, tree[0].Children[0].Children)
	assert.Equal(t, uint32(4), tree[0].Children[1].ID)

	assert.Equal(t, uint32(2), tree[1].ID)
	require.NotNil(t, tree[1].Children)
	assert.Empty(t, tree[1].Children)

	assert.Empty(t, buildLiveCategoryTree(nil))
}
