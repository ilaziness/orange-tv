package admin

import (
	"context"
	"testing"

	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateCategoryRejectsBlankName(t *testing.T) {
	t.Parallel()
	svc := NewLiveService(nil, nil, nil, nil, zap.NewNop())
	_, err := svc.CreateCategory(context.Background(), &admindto.CreateLiveCategoryRequest{Name: " \t "})
	got, ok := errcode.As(err)
	require.True(t, ok)
	assert.Equal(t, errcode.ParamError.Code, got.Code)
	assert.Equal(t, "分类名称不能为空", got.Message)
}
