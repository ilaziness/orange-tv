package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// PlayService manages play sources and episodes.
type PlayService interface {
	ListSources(ctx context.Context) ([]dto.PlaySourceResponse, error)
	CreateSource(ctx context.Context, req *dto.CreatePlaySourceRequest) (*dto.PlaySourceResponse, error)
	UpdateSource(ctx context.Context, id int64, req *dto.UpdatePlaySourceRequest) (*dto.PlaySourceResponse, error)
	DeleteSource(ctx context.Context, id int64) error

	ListEpisodes(ctx context.Context, req *dto.PlayEpisodeListRequest) ([]dto.PlayEpisodeResponse, int, error)
	CreateEpisode(ctx context.Context, req *dto.CreatePlayEpisodeRequest) (*dto.PlayEpisodeResponse, error)
	UpdateEpisode(ctx context.Context, id int64, req *dto.UpdatePlayEpisodeRequest) (*dto.PlayEpisodeResponse, error)
	DeleteEpisode(ctx context.Context, id int64) error
}

type playService struct {
	playRepo  repository.PlayRepository
	videoRepo repository.VideoRepository
}

// NewPlayService creates a PlayService.
func NewPlayService(playRepo repository.PlayRepository, videoRepo repository.VideoRepository) PlayService {
	return &playService{playRepo: playRepo, videoRepo: videoRepo}
}

func (s *playService) ListSources(ctx context.Context) ([]dto.PlaySourceResponse, error) {
	items, err := s.playRepo.ListSources(ctx)
	if err != nil {
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
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlaySourceDTO(*m)
	return &resp, nil
}

func (s *playService) UpdateSource(ctx context.Context, id int64, req *dto.UpdatePlaySourceRequest) (*dto.PlaySourceResponse, error) {
	m, err := s.playRepo.GetSource(ctx, id)
	if err != nil {
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
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlaySourceDTO(*m)
	return &resp, nil
}

func (s *playService) DeleteSource(ctx context.Context, id int64) error {
	m, err := s.playRepo.GetSource(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.PlaySourceNotFound
	}
	n, err := s.playRepo.CountEpisodesBySource(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if n > 0 {
		return errcode.PlaySourceInUse
	}
	if err := s.playRepo.SoftDeleteSource(ctx, id); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *playService) ListEpisodes(ctx context.Context, req *dto.PlayEpisodeListRequest) ([]dto.PlayEpisodeResponse, int, error) {
	items, total, err := s.playRepo.ListEpisodes(ctx, req.VideoID, req.SourceID, req.GetOffset(), req.GetLimit())
	if err != nil {
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
		title = formatEpisodeTitle(req.EpisodeNumber)
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
	existing, err := s.playRepo.GetEpisodeByKey(ctx, req.VideoID, req.SourceID, req.EpisodeNumber)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if existing != nil {
		if existing.DeletedAt == nil {
			return nil, errcode.PlayEpisodeDuplicate
		}
		m.ID = existing.ID
		if err := s.playRepo.RestoreAndUpdateEpisode(ctx, m); err != nil {
			if isDuplicateKey(err) {
				return nil, errcode.PlayEpisodeDuplicate
			}
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		resp := toPlayEpisodeDTO(*m)
		return &resp, nil
	}

	if err := s.playRepo.CreateEpisode(ctx, m); err != nil {
		if isDuplicateKey(err) {
			return nil, errcode.PlayEpisodeDuplicate
		}
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlayEpisodeDTO(*m)
	return &resp, nil
}

func (s *playService) UpdateEpisode(ctx context.Context, id int64, req *dto.UpdatePlayEpisodeRequest) (*dto.PlayEpisodeResponse, error) {
	m, err := s.playRepo.GetEpisode(ctx, id)
	if err != nil {
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
		existing, err := s.playRepo.GetEpisodeByKey(ctx, videoID, sourceID, episodeNumber)
		if err != nil {
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if existing != nil && existing.ID != id {
			if existing.DeletedAt == nil {
				return nil, errcode.PlayEpisodeDuplicate
			}
			// Free soft-deleted unique key before reassigning.
			if err := s.playRepo.HardDeleteEpisodeByKey(ctx, videoID, sourceID, episodeNumber, id); err != nil {
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
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
		if isDuplicateKey(err) {
			return nil, errcode.PlayEpisodeDuplicate
		}
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp := toPlayEpisodeDTO(*m)
	return &resp, nil
}

func (s *playService) DeleteEpisode(ctx context.Context, id int64) error {
	m, err := s.playRepo.GetEpisode(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.PlayEpisodeNotFound
	}
	if err := s.playRepo.SoftDeleteEpisode(ctx, id); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *playService) validateEpisodeRefs(ctx context.Context, videoID, sourceID int64) error {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil {
		return errcode.VideoNotFound
	}
	source, err := s.playRepo.GetSource(ctx, sourceID)
	if err != nil {
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

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
