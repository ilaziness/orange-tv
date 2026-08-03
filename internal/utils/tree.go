package utils

import (
	"github.com/ilaziness/orange-tv/internal/model"
)

// BuildCategoryTree builds a tree from a flat list of Categories using the provided mapper.
// The mapper receives a category model and its already-built child nodes, and returns
// the endpoint-specific tree node. This allows admin and client to produce their own
// DTO types without sharing a response struct.
func BuildCategoryTree[T any](items []model.Categories, mapper func(c model.Categories, children []T) T) []T {
	byParent := make(map[uint32][]model.Categories, len(items))
	for _, item := range items {
		byParent[item.ParentID] = append(byParent[item.ParentID], item)
	}
	var build func(parentID uint32) []T
	build = func(parentID uint32) []T {
		children := byParent[parentID]
		out := make([]T, 0, len(children))
		for _, c := range children {
			childNodes := build(c.ID)
			if childNodes == nil {
				childNodes = []T{}
			}
			out = append(out, mapper(c, childNodes))
		}
		return out
	}
	roots := build(0)
	if roots == nil {
		return []T{}
	}
	return roots
}
