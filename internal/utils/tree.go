package utils

import (
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	"github.com/ilaziness/orange-tv/internal/model"
)

// BuildCategoryTree builds a tree of CategoryResponse from a flat list of Categories.
func BuildCategoryTree(items []model.Categories) []shareddto.CategoryResponse {
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
			node := shareddto.CategoryResponse{
				ID:        c.ID,
				Name:      c.Name,
				ParentID:  c.ParentID,
				SortOrder: c.SortOrder,
				Status:    c.Status,
				Children:  childNodes,
			}
			out = append(out, node)
		}
		return out
	}
	roots := build(0)
	if roots == nil {
		return []shareddto.CategoryResponse{}
	}
	return roots
}
