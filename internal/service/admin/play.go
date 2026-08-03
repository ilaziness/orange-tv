package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// PlayService manages play sources and episodes.
type PlayService interface {
	ListSources(ctx context.Context) ([]dto.PlaySourceResponse, error)
	CreateSource(ctx context.Context, req *dto.CreatePlaySourceRequest) (*dto.PlaySourceResponse, error)
	UpdateSource(ctx context.Context, id uint32, req *dto.UpdatePlaySourceRequest) (*dto.PlaySourceResponse, error)
	DeleteSource(ctx context.Context, id uint32) error

	ListEpisodes(ctx context.Context, req *dto.PlayEpisodeListRequest) ([]dto.PlayEpisodeResponse, int, error)
	CreateEpisode(ctx context.Context, req *dto.CreatePlayEpisodeRequest) (*dto.PlayEpisodeResponse, error)
	UpdateEpisode(ctx context.Context, id uint32, req *dto.UpdatePlayEpisodeRequest) (*dto.PlayEpisodeResponse, error)
	DeleteEpisode(ctx context.Context, id uint32) error
	BatchUpdateEpisodeStatus(ctx context.Context, req *dto.BatchUpdateEpisodeStatusRequest) (*dto.BatchUpdateEpisodeStatusResponse, error)
}

type playService struct {
	playRepo  repository.PlayRepository
	videoRepo repository.VideoRepository
	log       *zap.Logger
}

// NewPlayService creates a PlayService.
func NewPlayService(playRepo repository.PlayRepository, videoRepo repository.VideoRepository, log *zap.Logger) PlayService {
	if log == nil {
		log = zap.NewNop()
	}
	return &playService{playRepo: playRepo, videoRepo: videoRepo, log: log}
}

func (s *playService) ListSources(ctx context.Context) ([]dto.PlaySourceResponse, error) {
	items, err := s.playRepo.ListSources(ctx)
	if err != nil {
		s.log.Error("play: list sources failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.PlaySourceResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toPlaySourceDTO(item))
	}
	return out, nil
}

func (s *playService) CreateSource(ctx context.Context, req *dto.CreatePlaySourceRequest) (*dto.PlaySourceResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "播放源名称不能为空")
	}
	exists, err := s.playRepo.ExistsSourceName(ctx, name, 0)
	if err != nil {
		s.log.Error("play: check source name exists failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.PlaySourceNameExists
	}
	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	m := &model.PlaySources{Name: name, SortOrder: req.SortOrder, Status: status}
	if err := s.playRepo.CreateSource(ctx, m); err != nil {
		s.log.Error("play: create source failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlaySourceDTO(*m)
	return &resp, nil
}

func (s *playService) UpdateSource(ctx context.Context, id uint32, req *dto.UpdatePlaySourceRequest) (*dto.PlaySourceResponse, error) {
	m, err := s.playRepo.GetSource(ctx, id)
	if err != nil {
		s.log.Error("play: get source failed", zap.Uint32("source_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.PlaySourceNotFound
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "播放源名称不能为空")
		}
		exists, err := s.playRepo.ExistsSourceName(ctx, name, id)
		if err != nil {
			s.log.Error("play: check source name exists for update failed", zap.Uint32("source_id", id), zap.String("name", name), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if exists {
			return nil, errcode.PlaySourceNameExists
		}
		m.Name = name
	}
	if req.SortOrder != nil {
		m.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := s.playRepo.UpdateSource(ctx, m); err != nil {
		s.log.Error("play: update source failed", zap.Uint32("source_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlaySourceDTO(*m)
	return &resp, nil
}

func (s *playService) DeleteSource(ctx context.Context, id uint32) error {
	m, err := s.playRepo.GetSource(ctx, id)
	if err != nil {
		s.log.Error("play: get source for delete failed", zap.Uint32("source_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.PlaySourceNotFound
	}
	n, err := s.playRepo.CountEpisodesBySource(ctx, id)
	if err != nil {
		s.log.Error("play: count episodes by source failed", zap.Uint32("source_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if n > 0 {
		return errcode.PlaySourceInUse
	}
	if err := s.playRepo.SoftDeleteSource(ctx, id); err != nil {
		s.log.Error("play: soft delete source failed", zap.Uint32("source_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *playService) ListEpisodes(ctx context.Context, req *dto.PlayEpisodeListRequest) ([]dto.PlayEpisodeResponse, int, error) {
	items, total, err := s.playRepo.ListEpisodes(ctx, req.VideoID, req.SourceID, req.GetOffset(), req.GetLimit())
	if err != nil {
		s.log.Error("play: list episodes failed", zap.Uint32("video_id", req.VideoID), zap.Uint32("source_id", req.SourceID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.PlayEpisodeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toPlayEpisodeDTO(item))
	}
	return out, total, nil
}

func (s *playService) CreateEpisode(ctx context.Context, req *dto.CreatePlayEpisodeRequest) (*dto.PlayEpisodeResponse, error) {
	if err := s.validateEpisodeRefs(ctx, req.VideoID, req.SourceID); err != nil {
		return nil, err
	}
	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = formatEpisodeTitle(int32(req.EpisodeNumber))
	}
	m := &model.PlayEpisodes{
		SourceID:      req.SourceID,
		VideoID:       req.VideoID,
		EpisodeNumber: req.EpisodeNumber,
		Title:         title,
		PlayURL:       strings.TrimSpace(req.PlayURL),
		Quality:       strings.TrimSpace(req.Quality),
		Format:        strings.TrimSpace(req.Format),
		SortOrder:     req.SortOrder,
		Status:        status,
	}

	// Unique index covers soft-deleted rows. Reuse a soft-deleted slot when present.
	existing, err := s.playRepo.GetEpisodeByKey(ctx, req.VideoID, req.SourceID, int32(req.EpisodeNumber))
	if err != nil {
		s.log.Error("play: get episode by key failed", zap.Uint32("video_id", req.VideoID), zap.Uint32("source_id", req.SourceID), zap.Uint32("episode_number", req.EpisodeNumber), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if existing != nil {
		return nil, errcode.PlayEpisodeDuplicate
	}

	if err := s.playRepo.CreateEpisode(ctx, m); err != nil {
		if utils.IsDuplicateKey(err) {
			return nil, errcode.PlayEpisodeDuplicate
		}
		s.log.Error("play: create episode failed", zap.Uint32("video_id", req.VideoID), zap.Uint32("source_id", req.SourceID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlayEpisodeDTO(*m)
	return &resp, nil
}

func (s *playService) UpdateEpisode(ctx context.Context, id uint32, req *dto.UpdatePlayEpisodeRequest) (*dto.PlayEpisodeResponse, error) {
	m, err := s.playRepo.GetEpisode(ctx, id)
	if err != nil {
		s.log.Error("play: get episode failed", zap.Uint32("episode_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.PlayEpisodeNotFound
	}
	videoID, sourceID, episodeNumber := m.VideoID, m.SourceID, m.EpisodeNumber
	if req.VideoID != nil {
		videoID = *req.VideoID
	}
	if req.SourceID != nil {
		sourceID = *req.SourceID
	}
	if req.EpisodeNumber != nil {
		episodeNumber = *req.EpisodeNumber
	}
	if err := s.validateEpisodeRefs(ctx, videoID, sourceID); err != nil {
		return nil, err
	}
	if videoID != m.VideoID || sourceID != m.SourceID || episodeNumber != m.EpisodeNumber {
		existing, err := s.playRepo.GetEpisodeByKey(ctx, videoID, sourceID, int32(episodeNumber))
		if err != nil {
			s.log.Error("play: get episode by key for update failed", zap.Uint32("video_id", videoID), zap.Uint32("source_id", sourceID), zap.Uint32("episode_number", episodeNumber), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if existing != nil && existing.ID != id {
			return nil, errcode.PlayEpisodeDuplicate
		}
	}
	m.VideoID = videoID
	m.SourceID = sourceID
	m.EpisodeNumber = episodeNumber
	if req.Title != nil {
		m.Title = strings.TrimSpace(*req.Title)
	}
	if req.PlayURL != nil {
		m.PlayURL = strings.TrimSpace(*req.PlayURL)
	}
	if req.Quality != nil {
		m.Quality = strings.TrimSpace(*req.Quality)
	}
	if req.Format != nil {
		m.Format = strings.TrimSpace(*req.Format)
	}
	if req.SortOrder != nil {
		m.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := s.playRepo.UpdateEpisode(ctx, m); err != nil {
		if utils.IsDuplicateKey(err) {
			return nil, errcode.PlayEpisodeDuplicate
		}
		s.log.Error("play: update episode failed", zap.Uint32("episode_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlayEpisodeDTO(*m)
	return &resp, nil
}

func (s *playService) DeleteEpisode(ctx context.Context, id uint32) error {
	m, err := s.playRepo.GetEpisode(ctx, id)
	if err != nil {
		s.log.Error("play: get episode for delete failed", zap.Uint32("episode_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.PlayEpisodeNotFound
	}
	if err := s.playRepo.SoftDeleteEpisode(ctx, id); err != nil {
		s.log.Error("play: soft delete episode failed", zap.Uint32("episode_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// BatchUpdateEpisodeStatus 批量更新某影视下指定播放源的全部剧集状态。
func (s *playService) BatchUpdateEpisodeStatus(ctx context.Context, req *dto.BatchUpdateEpisodeStatusRequest) (*dto.BatchUpdateEpisodeStatusResponse, error) {
	if err := s.validateEpisodeRefs(ctx, req.VideoID, req.SourceID); err != nil {
		return nil, err
	}
	n, err := s.playRepo.UpdateEpisodeStatusBySource(ctx, req.VideoID, req.SourceID, req.Status)
	if err != nil {
		s.log.Error("play: batch update episode status failed",
			zap.Uint32("video_id", req.VideoID),
			zap.Uint32("source_id", req.SourceID),
			zap.Uint8("status", req.Status),
			zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.BatchUpdateEpisodeStatusResponse{Affected: n}, nil
}

func (s *playService) validateEpisodeRefs(ctx context.Context, videoID, sourceID uint32) error {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("play: validate episode refs get video failed", zap.Uint32("video_id", videoID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil {
		return errcode.VideoNotFound
	}
	source, err := s.playRepo.GetSource(ctx, sourceID)
	if err != nil {
		s.log.Error("play: validate episode refs get source failed", zap.Uint32("source_id", sourceID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if source == nil {
		return errcode.PlaySourceNotFound
	}
	return nil
}

func toPlaySourceDTO(m model.PlaySources) dto.PlaySourceResponse {
	return dto.PlaySourceResponse{
		ID:        m.ID,
		Name:      m.Name,
		SortOrder: m.SortOrder,
		Status:    m.Status,
	}
}

func toPlayEpisodeDTO(m model.PlayEpisodes) dto.PlayEpisodeResponse {
	return dto.PlayEpisodeResponse{
		ID:            m.ID,
		SourceID:      m.SourceID,
		VideoID:       m.VideoID,
		EpisodeNumber: m.EpisodeNumber,
		Title:         m.Title,
		PlayURL:       m.PlayURL,
		Quality:       m.Quality,
		Format:        m.Format,
		SortOrder:     m.SortOrder,
		Status:        m.Status,
	}
}

func formatEpisodeTitle(n int32) string {
	return "第" + itoa32(n) + "集"
}

func itoa32(n int32) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
