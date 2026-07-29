package admin

import (
	"context"

	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// CommentService manages admin comment operations.
type CommentService interface {
	List(ctx context.Context, req *dto.CommentListRequest) ([]dto.CommentListItem, int, error)
	UpdateStatus(ctx context.Context, id int64, req *dto.UpdateCommentStatusRequest) error
	Delete(ctx context.Context, id int64) error
	GetParents(ctx context.Context, id int64) ([]dto.CommentParentItem, error)
}

type commentService struct {
	repo repository.CommentRepository
	log  *zap.Logger
}

// NewCommentService creates a CommentService.
func NewCommentService(repo repository.CommentRepository, log *zap.Logger) CommentService {
	if log == nil {
		log = zap.NewNop()
	}
	return &commentService{repo: repo, log: log}
}

func (s *commentService) List(ctx context.Context, req *dto.CommentListRequest) ([]dto.CommentListItem, int, error) {
	items, total, err := s.repo.List(ctx, repository.CommentListFilter{
		Keyword: req.Keyword,
		Status:  req.Status,
		VideoID: req.VideoID,
		Offset:  req.GetOffset(),
		Limit:   req.GetLimit(),
	})
	if err != nil {
		s.log.Error("comment: list failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapCommentList(items), total, nil
}

func (s *commentService) UpdateStatus(ctx context.Context, id int64, req *dto.UpdateCommentStatusRequest) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("comment: get by id for status update failed", zap.Int64("comment_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return errcode.CommentNotFound
	}
	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		s.log.Error("comment: update status failed", zap.Int64("comment_id", id), zap.Uint8("status", req.Status), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *commentService) Delete(ctx context.Context, id int64) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("comment: get by id for delete failed", zap.Int64("comment_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return errcode.CommentNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("comment: delete failed", zap.Int64("comment_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *commentService) GetParents(ctx context.Context, id int64) ([]dto.CommentParentItem, error) {
	chain, err := s.repo.GetParentChain(ctx, id)
	if err != nil {
		s.log.Error("comment: get parent chain failed", zap.Int64("comment_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if chain == nil {
		return nil, errcode.CommentNotFound
	}
	return mapCommentParents(chain), nil
}

func mapCommentList(items []model.VideoComments) []dto.CommentListItem {
	out := make([]dto.CommentListItem, 0, len(items))
	for _, c := range items {
		username := "-"
		if c.User != nil {
			username = c.User.Username
		}
		videoTitle := "-"
		if c.Video != nil {
			videoTitle = c.Video.Title
		}
		out = append(out, dto.CommentListItem{
			ID:           c.ID,
			VideoID:      c.VideoID,
			VideoTitle:   videoTitle,
			Content:      c.Content,
			UserID:       c.UserID,
			Username:     username,
			Status:       c.Status,
			LikeCount:    c.LikeCount,
			DislikeCount: c.DislikeCount,
			ParentID:     c.ParentID,
			CreatedAt:    utils.FormatTimeStr(&c.CreatedAt),
		})
	}
	return out
}

func mapCommentParents(chain []model.VideoComments) []dto.CommentParentItem {
	out := make([]dto.CommentParentItem, 0, len(chain))
	for _, c := range chain {
		username := "-"
		if c.User != nil {
			username = c.User.Username
		}
		out = append(out, dto.CommentParentItem{
			ID:        c.ID,
			UserID:    c.UserID,
			Username:  username,
			ParentID:  c.ParentID,
			Content:   c.Content,
			CreatedAt: utils.FormatTimeStr(&c.CreatedAt),
		})
	}
	return out
}
