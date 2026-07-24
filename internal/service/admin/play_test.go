package admin

import (
	"context"
	"testing"
	"time"

	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type fakePlayRepo struct {
	sources  map[uint64]*model.PlaySources
	episodes map[uint64]*model.PlayEpisodes
	nextID   uint64
}

func newFakePlayRepo() *fakePlayRepo {
	return &fakePlayRepo{
		sources:  map[uint64]*model.PlaySources{},
		episodes: map[uint64]*model.PlayEpisodes{},
		nextID:   1,
	}
}

func (f *fakePlayRepo) ListSources(ctx context.Context) ([]model.PlaySources, error) {
	return nil, nil
}
func (f *fakePlayRepo) GetSource(ctx context.Context, id int64) (*model.PlaySources, error) {
	s, ok := f.sources[uint64(id)]
	if !ok || s.DeletedAt != nil {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}
func (f *fakePlayRepo) ExistsSourceName(ctx context.Context, name string, excludeID int64) (bool, error) {
	return false, nil
}
func (f *fakePlayRepo) CreateSource(ctx context.Context, m *model.PlaySources) error { return nil }
func (f *fakePlayRepo) UpdateSource(ctx context.Context, m *model.PlaySources) error { return nil }
func (f *fakePlayRepo) SoftDeleteSource(ctx context.Context, id int64) error         { return nil }
func (f *fakePlayRepo) CountEpisodesBySource(ctx context.Context, sourceID int64) (int, error) {
	return 0, nil
}
func (f *fakePlayRepo) ListEpisodes(ctx context.Context, videoID, sourceID int64, offset, limit int) ([]model.PlayEpisodes, int, error) {
	return nil, 0, nil
}
func (f *fakePlayRepo) ListEpisodesByVideo(ctx context.Context, videoID int64, onlyEnabled bool) ([]model.PlayEpisodes, error) {
	return nil, nil
}
func (f *fakePlayRepo) GetEpisode(ctx context.Context, id int64) (*model.PlayEpisodes, error) {
	ep, ok := f.episodes[uint64(id)]
	if !ok || ep.DeletedAt != nil {
		return nil, nil
	}
	cp := *ep
	return &cp, nil
}
func (f *fakePlayRepo) GetEpisodeByKey(ctx context.Context, videoID, sourceID int64, episodeNumber int32) (*model.PlayEpisodes, error) {
	for _, ep := range f.episodes {
		if ep.VideoID == uint64(videoID) && ep.SourceID == uint64(sourceID) && ep.EpisodeNumber == uint32(episodeNumber) {
			cp := *ep
			return &cp, nil
		}
	}
	return nil, nil
}
func (f *fakePlayRepo) ExistsEpisode(ctx context.Context, videoID, sourceID int64, episodeNumber int32, excludeID int64) (bool, error) {
	for id, ep := range f.episodes {
		if id == uint64(excludeID) || ep.DeletedAt != nil {
			continue
		}
		if ep.VideoID == uint64(videoID) && ep.SourceID == uint64(sourceID) && ep.EpisodeNumber == uint32(episodeNumber) {
			return true, nil
		}
	}
	return false, nil
}
func (f *fakePlayRepo) CreateEpisode(ctx context.Context, m *model.PlayEpisodes) error {
	m.ID = f.nextID
	f.nextID++
	cp := *m
	f.episodes[m.ID] = &cp
	return nil
}
func (f *fakePlayRepo) UpdateEpisode(ctx context.Context, m *model.PlayEpisodes) error {
	cp := *m
	f.episodes[m.ID] = &cp
	return nil
}
func (f *fakePlayRepo) RestoreAndUpdateEpisode(ctx context.Context, m *model.PlayEpisodes) error {
	m.DeletedAt = nil
	cp := *m
	f.episodes[m.ID] = &cp
	return nil
}
func (f *fakePlayRepo) HardDeleteEpisodeByKey(ctx context.Context, videoID, sourceID int64, episodeNumber int32, excludeID int64) error {
	for id, ep := range f.episodes {
		if id == uint64(excludeID) || ep.DeletedAt == nil {
			continue
		}
		if ep.VideoID == uint64(videoID) && ep.SourceID == uint64(sourceID) && ep.EpisodeNumber == uint32(episodeNumber) {
			delete(f.episodes, id)
		}
	}
	return nil
}
func (f *fakePlayRepo) SoftDeleteEpisode(ctx context.Context, id int64) error {
	if ep, ok := f.episodes[uint64(id)]; ok {
		now := time.Now()
		ep.DeletedAt = &now
	}
	return nil
}

func (f *fakePlayRepo) WithTx(tx bun.Tx) repository.PlayRepository { return f }

type videoRepoStub struct {
	videos map[uint64]*model.Videos
}

func (v *videoRepoStub) List(ctx context.Context, f repository.VideoListFilter) ([]model.Videos, int, error) {
	return nil, 0, nil
}
func (v *videoRepoStub) GetByID(ctx context.Context, id uint64) (*model.Videos, error) {
	item, ok := v.videos[id]
	if !ok {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}
func (v *videoRepoStub) Create(ctx context.Context, video *model.Videos) error { return nil }
func (v *videoRepoStub) BatchCreate(ctx context.Context, videos []*model.Videos) error {
	return nil
}
func (v *videoRepoStub) Update(ctx context.Context, video *model.Videos) error { return nil }
func (v *videoRepoStub) SoftDelete(ctx context.Context, id uint64) error       { return nil }
func (v *videoRepoStub) ReplaceDirectors(ctx context.Context, videoID uint64, directorIDs []uint64) error {
	return nil
}
func (v *videoRepoStub) ReplaceActors(ctx context.Context, videoID uint64, actors []model.VideoActors) error {
	return nil
}
func (v *videoRepoStub) ReplaceTags(ctx context.Context, videoID uint64, tagIDs []uint64) error {
	return nil
}
func (v *videoRepoStub) ListDirectorIDs(ctx context.Context, videoID uint64) ([]uint64, error) {
	return nil, nil
}
func (v *videoRepoStub) ListActorRels(ctx context.Context, videoID uint64) ([]model.VideoActors, error) {
	return nil, nil
}
func (v *videoRepoStub) ListTagIDs(ctx context.Context, videoID uint64) ([]uint64, error) {
	return nil, nil
}
func (v *videoRepoStub) RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return nil
}
func (v *videoRepoStub) WithTx(tx bun.Tx) repository.VideoRepository { return v }
func (v *videoRepoStub) BatchUpdatePublishStatus(ctx context.Context, ids []uint64, status uint8) (int, error) {
	return 0, nil
}
func (v *videoRepoStub) BatchSoftDelete(ctx context.Context, ids []uint64) (int, error) {
	return 0, nil
}
func (v *videoRepoStub) CountVideos(ctx context.Context) (int, error) { return 0, nil }
func (v *videoRepoStub) CountVideosToday(ctx context.Context, since time.Time) (int, error) {
	return 0, nil
}
func (v *videoRepoStub) CountVideosByStatus(ctx context.Context, status uint8) (int, error) {
	return 0, nil
}
func (v *videoRepoStub) CountCategories(ctx context.Context) (int, error) { return 0, nil }

func TestPlayService_CreateEpisodeRestoresSoftDeleted(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	now := time.Now()
	playRepo.episodes[9] = &model.PlayEpisodes{
		ID: 9, VideoID: 1, SourceID: 1, EpisodeNumber: 1, Title: "旧", PlayURL: "old", Format: "hls", DeletedAt: &now,
	}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint64]*model.Videos{1: {ID: 1, Title: "v"}}})

	resp, err := svc.CreateEpisode(context.Background(), &dto.CreatePlayEpisodeRequest{
		SourceID: 1, VideoID: 1, EpisodeNumber: 1, PlayURL: "https://example.com/1.m3u8", Format: "hls",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(9), resp.ID)
	require.Equal(t, "https://example.com/1.m3u8", resp.PlayURL)
	require.Nil(t, playRepo.episodes[9].DeletedAt)
}

func TestPlayService_CreateEpisodeRejectsActiveDuplicate(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	playRepo.episodes[9] = &model.PlayEpisodes{
		ID: 9, VideoID: 1, SourceID: 1, EpisodeNumber: 1, Title: "旧", PlayURL: "old", Format: "hls",
	}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint64]*model.Videos{1: {ID: 1, Title: "v"}}})

	_, err := svc.CreateEpisode(context.Background(), &dto.CreatePlayEpisodeRequest{
		SourceID: 1, VideoID: 1, EpisodeNumber: 1, PlayURL: "https://example.com/1.m3u8", Format: "hls",
	})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.PlayEpisodeDuplicate.Code, code.Code)
}
