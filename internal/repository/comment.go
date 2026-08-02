package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
)

// CommentListFilter filters admin comment list queries.
type CommentListFilter struct {
	Keyword string
	Status  *uint8
	VideoID *uint64
	Offset  int
	Limit   int
}

// CommentRepository provides admin comment persistence.
type CommentRepository interface {
	List(ctx context.Context, f CommentListFilter) ([]model.VideoComments, int, error)
	GetByID(ctx context.Context, id int64) (*model.VideoComments, error)
	GetParentChain(ctx context.Context, id int64) ([]model.VideoComments, error)
	UpdateStatus(ctx context.Context, id int64, status uint8) error
	Delete(ctx context.Context, id int64) error
	DeleteTree(ctx context.Context, id int64) error
}

type commentRepo struct {
	db *database.DB
}

// NewCommentRepo creates a CommentRepository.
func NewCommentRepo(db *database.DB) CommentRepository {
	return &commentRepo{db: db}
}

func (r *commentRepo) List(ctx context.Context, f CommentListFilter) ([]model.VideoComments, int, error) {
	items := make([]model.VideoComments, 0, f.Limit)
	q := r.db.NewSelect().Model(&items).
		Relation("User").
		Relation("Video")

	if f.Status != nil {
		q = q.Where("vc.status = ?", *f.Status)
	}
	if f.VideoID != nil {
		q = q.Where("vc.video_id = ?", *f.VideoID)
	}
	if f.Keyword != "" {
		q = q.Where("vc.content LIKE ?", "%"+f.Keyword+"%")
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}
	if err := q.Order("vc.id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list comments: %w", err)
	}
	return items, total, nil
}

func (r *commentRepo) GetByID(ctx context.Context, id int64) (*model.VideoComments, error) {
	c := new(model.VideoComments)
	err := r.db.NewSelect().Model(c).
		Relation("User").
		Relation("Video").
		Where("vc.id = ?", id).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return c, nil
}

func (r *commentRepo) GetParentChain(ctx context.Context, id int64) ([]model.VideoComments, error) {
	const maxDepth = 50

	c, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	if c.ParentID == 0 {
		return []model.VideoComments{}, nil
	}

	chain := make([]model.VideoComments, 0, 8)
	visited := make(map[uint64]struct{}, maxDepth)
	currentID := c.ParentID

	for i := 0; i < maxDepth; i++ {
		if currentID == 0 {
			break
		}
		if _, ok := visited[currentID]; ok {
			break
		}
		visited[currentID] = struct{}{}

		parent, err := r.GetByID(ctx, int64(currentID))
		if err != nil {
			return nil, err
		}
		if parent == nil {
			break
		}
		chain = append(chain, *parent)
		currentID = parent.ParentID
	}

	// Reverse so order is root -> direct parent.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain, nil
}

func (r *commentRepo) UpdateStatus(ctx context.Context, id int64, status uint8) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.VideoComments)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update comment status: %w", err)
	}
	return nil
}

func (r *commentRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*model.VideoComments)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return nil
}

func (r *commentRepo) DeleteTree(ctx context.Context, id int64) error {
	return deleteCommentTreeByID(ctx, r.db, id)
}
