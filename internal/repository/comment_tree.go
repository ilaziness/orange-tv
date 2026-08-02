package repository

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

const commentDescendantMaxDepth = 50

func getCommentDescendantIDs(ctx context.Context, db bun.IDB, id int64) ([]int64, error) {
	visited := make(map[int64]struct{}, 64)
	current := []int64{id}
	descendants := make([]int64, 0, 64)
	for depth := 0; depth < commentDescendantMaxDepth && len(current) > 0; depth++ {
		var ids []int64
		err := db.NewSelect().TableExpr("video_comments").
			ColumnExpr("id").
			Where("parent_id IN (?)", bun.In(current)).
			Scan(ctx, &ids)
		if err != nil {
			return nil, fmt.Errorf("get comment descendants: %w", err)
		}
		next := make([]int64, 0, len(ids))
		for _, childID := range ids {
			if _, ok := visited[childID]; ok {
				continue
			}
			visited[childID] = struct{}{}
			descendants = append(descendants, childID)
			next = append(next, childID)
		}
		current = next
	}
	return descendants, nil
}

func deleteCommentTreeByID(ctx context.Context, db bun.IDB, id int64) error {
	descendants, err := getCommentDescendantIDs(ctx, db, id)
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(descendants)+1)
	ids = append(ids, id)
	ids = append(ids, descendants...)
	_, err = db.NewDelete().Model((*model.VideoComments)(nil)).
		Where("id IN (?)", bun.In(ids)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete comment tree: %w", err)
	}
	return nil
}
