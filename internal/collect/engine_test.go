package collect

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/require"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"normal https", "https://img.example.com/path/cover.jpg", "img.example.com"},
		{"normal http", "http://cdn.old.com:8080/x/y.m3u8", "cdn.old.com:8080"},
		{"with port", "https://api.site.com:443/video", "api.site.com:443"},
		{"trailing slash", "https://a.com/", "a.com"},
		{"no path", "https://a.com", "a.com"},
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
		{"no scheme returns empty (Go url.Parse treats it as path)", "img.example.com/path", ""},
		{"scheme-relative", "//cdn.example.com/x", "cdn.example.com"},
		{"invalid url", "://bad", ""},
		{"with query", "https://h.com/p?a=b&c=d", "h.com"},
		{"with fragment", "https://h.com/p#frag", "h.com"},
		{"userinfo stripped", "https://user:pass@h.com/p", "h.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHost(tt.raw)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestApplySupplementFields(t *testing.T) {
	// bound category 5 is a second-level category whose parent is 1.
	const catID, parentCatID = 5, 1

	t.Run("fills all empty fields", func(t *testing.T) {
		v := &model.Videos{}
		item := Item{
			Subtitle:    "副标题",
			Description: "描述",
			Cover:       "https://img.new.com/c.jpg",
			Year:        2024,
			Region:      "美国",
			Language:    "英语",
			Duration:    120,
			ReleaseDate: "2024-01-01",
		}
		applySupplementFields(v, item, catID, parentCatID, constant.SerialStatusFinished, true)
		require.Equal(t, "副标题", v.Subtitle)
		require.Equal(t, "描述", v.Description)
		require.Equal(t, "https://img.new.com/c.jpg", v.CoverImage)
		require.Equal(t, uint32(2024), v.Year)
		require.Equal(t, "美国", v.Region)
		require.Equal(t, "英语", v.Language)
		require.Equal(t, uint32(120), v.Duration)
		require.Equal(t, "2024-01-01", v.ReleaseDate)
		require.Equal(t, uint8(constant.SerialStatusFinished), v.SerialStatus)
		require.Equal(t, uint32(5), v.CategoryID)
		require.Equal(t, uint32(1), v.ParentCategoryID)
	})

	t.Run("does not overwrite non-empty fields", func(t *testing.T) {
		v := &model.Videos{
			Subtitle:         "原副标题",
			Description:      "原有描述",
			CoverImage:       "https://img.old.com/old.jpg",
			Year:             2020,
			Region:           "中国",
			Language:         "中文",
			Duration:         90,
			ReleaseDate:      "2020-05-05",
			SerialStatus:     constant.SerialStatusOngoing,
			CategoryID:       3,
			ParentCategoryID: 2,
		}
		item := Item{
			Subtitle:    "新副标题",
			Description: "新描述",
			Cover:       "https://img.new.com/new.jpg",
			Year:        2024,
			Region:      "美国",
			Language:    "英语",
			Duration:    120,
			ReleaseDate: "2024-01-01",
		}
		applySupplementFields(v, item, catID, parentCatID, constant.SerialStatusFinished, true)
		// all basic fields keep original values (supplement semantics)
		require.Equal(t, "原副标题", v.Subtitle)
		require.Equal(t, "原有描述", v.Description)
		require.Equal(t, uint32(2020), v.Year)
		require.Equal(t, "中国", v.Region)
		require.Equal(t, "中文", v.Language)
		require.Equal(t, uint32(90), v.Duration)
		require.Equal(t, "2020-05-05", v.ReleaseDate)
		require.Equal(t, uint8(constant.SerialStatusOngoing), v.SerialStatus)
		require.Equal(t, uint32(3), v.CategoryID)
		require.Equal(t, uint32(2), v.ParentCategoryID)
	})

	t.Run("cover overridden when updateCover true (same-source)", func(t *testing.T) {
		v := &model.Videos{CoverImage: "https://img.old.com/old.jpg"}
		item := Item{Cover: "https://img.new.com/new.jpg"}
		applySupplementFields(v, item, catID, parentCatID, 0, true)
		require.Equal(t, "https://img.new.com/new.jpg", v.CoverImage)
	})

	t.Run("cover not touched when updateCover false (cross-source)", func(t *testing.T) {
		v := &model.Videos{CoverImage: "https://img.old.com/old.jpg"}
		item := Item{Cover: "https://img.new.com/new.jpg"}
		applySupplementFields(v, item, catID, parentCatID, 0, false)
		require.Equal(t, "https://img.old.com/old.jpg", v.CoverImage)
	})

	t.Run("cover not changed when item cover empty", func(t *testing.T) {
		v := &model.Videos{CoverImage: "https://img.old.com/old.jpg"}
		item := Item{Cover: ""}
		applySupplementFields(v, item, catID, parentCatID, 0, true)
		require.Equal(t, "https://img.old.com/old.jpg", v.CoverImage)
	})

	t.Run("description empty string filled", func(t *testing.T) {
		v := &model.Videos{}
		item := Item{Description: "新描述"}
		applySupplementFields(v, item, catID, parentCatID, 0, false)
		require.Equal(t, "新描述", v.Description)
	})

	t.Run("description not overwritten when non-empty", func(t *testing.T) {
		v := &model.Videos{Description: "原有描述"}
		item := Item{Description: "新描述"}
		applySupplementFields(v, item, catID, parentCatID, 0, false)
		require.Equal(t, "原有描述", v.Description)
	})

	t.Run("release date trimmed", func(t *testing.T) {
		v := &model.Videos{}
		item := Item{ReleaseDate: "  2024-01-01  "}
		applySupplementFields(v, item, catID, parentCatID, 0, false)
		require.Equal(t, "2024-01-01", v.ReleaseDate)
	})

	t.Run("serial status not filled when zero", func(t *testing.T) {
		v := &model.Videos{}
		item := Item{}
		applySupplementFields(v, item, catID, parentCatID, 0, false)
		require.Equal(t, uint8(0), v.SerialStatus)
	})

	t.Run("publish status never modified", func(t *testing.T) {
		v := &model.Videos{PublishStatus: 0}
		item := Item{}
		applySupplementFields(v, item, catID, parentCatID, constant.SerialStatusFinished, true)
		require.Equal(t, uint8(0), v.PublishStatus) // still 0, not set to Online
	})
}

func TestResolveCategoryFields(t *testing.T) {
	parentMap := map[uint32]uint32{
		1: 0, // 一级分类
		2: 1, // 二级分类，父为 1
		5: 1, // 二级分类，父为 1
	}

	t.Run("top-level category writes to parent_category_id", func(t *testing.T) {
		catID, parentCatID := resolveCategoryFields(1, parentMap)
		require.Equal(t, uint32(0), catID)
		require.Equal(t, uint32(1), parentCatID)
	})

	t.Run("second-level category writes to category_id", func(t *testing.T) {
		catID, parentCatID := resolveCategoryFields(5, parentMap)
		require.Equal(t, uint32(5), catID)
		require.Equal(t, uint32(1), parentCatID)
	})

	t.Run("unknown category treated as top-level", func(t *testing.T) {
		catID, parentCatID := resolveCategoryFields(999, parentMap)
		require.Equal(t, uint32(0), catID)
		require.Equal(t, uint32(999), parentCatID)
	})
}
