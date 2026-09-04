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
	sources  map[uint32]*model.PlaySources
	episodes map[uint32]*model.PlayEpisodes
	nextID   uint32
}

func newFakePlayRepo() *fakePlayRepo {
	return &fakePlayRepo{
		sources:  map[uint32]*model.PlaySources{},
		episodes: map[uint32]*model.PlayEpisodes{},
		nextID:   1,
	}
}

func (f *fakePlayRepo) ListSources(ctx context.Context) ([]model.PlaySources, error) {
	return nil, nil
}
func (f *fakePlayRepo) GetSource(ctx context.Context, id uint32) (*model.PlaySources, error) {
	s, ok := f.sources[id]
	if !ok || s.DeletedAt != nil {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}
func (f *fakePlayRepo) ExistsSourceName(ctx context.Context, name string, excludeID uint32) (bool, error) {
	return false, nil
}
func (f *fakePlayRepo) CreateSource(ctx context.Context, m *model.PlaySources) error { return nil }
func (f *fakePlayRepo) UpdateSource(ctx context.Context, m *model.PlaySources) error { return nil }
func (f *fakePlayRepo) SoftDeleteSource(ctx context.Context, id uint32) error        { return nil }
func (f *fakePlayRepo) CountEpisodesBySource(ctx context.Context, sourceID uint32) (int, error) {
	return 0, nil
}
func (f *fakePlayRepo) ListEpisodes(ctx context.Context, videoID, sourceID uint32, offset, limit int) ([]model.PlayEpisodes, int, error) {
	return nil, 0, nil
}
func (f *fakePlayRepo) ListEpisodesByVideo(ctx context.Context, videoID uint32, onlyEnabled bool) ([]model.PlayEpisodes, error) {
	return nil, nil
}
func (f *fakePlayRepo) GetEpisode(ctx context.Context, id uint32) (*model.PlayEpisodes, error) {
	ep, ok := f.episodes[id]
	if !ok {
		return nil, nil
	}
	cp := *ep
	return &cp, nil
}
func (f *fakePlayRepo) GetPlayableEpisodeByID(ctx context.Context, videoID, episodeID uint32) (*model.PlayEpisodes, error) {
	ep, ok := f.episodes[episodeID]
	if !ok || ep.VideoID != videoID || ep.Status != 1 {
		return nil, nil
	}
	cp := *ep
	return &cp, nil
}
func (f *fakePlayRepo) GetEpisodeByKey(ctx context.Context, videoID, sourceID uint32, episodeNumber int32) (*model.PlayEpisodes, error) {
	for _, ep := range f.episodes {
		if ep.VideoID == videoID && ep.SourceID == sourceID && ep.EpisodeNumber == uint32(episodeNumber) {
			cp := *ep
			return &cp, nil
		}
	}
	return nil, nil
}
func (f *fakePlayRepo) ExistsEpisode(ctx context.Context, videoID, sourceID uint32, episodeNumber int32, excludeID uint32) (bool, error) {
	for id, ep := range f.episodes {
		if id == excludeID {
			continue
		}
		if ep.VideoID == videoID && ep.SourceID == sourceID && ep.EpisodeNumber == uint32(episodeNumber) {
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
func (f *fakePlayRepo) SoftDeleteEpisode(ctx context.Context, id uint32) error {
	delete(f.episodes, id)
	return nil
}
func (f *fakePlayRepo) UpdateEpisodeStatusBySource(ctx context.Context, videoID, sourceID uint32, status uint8) (int, error) {
	n := 0
	for _, ep := range f.episodes {
		if ep.VideoID == videoID && ep.SourceID == sourceID {
			ep.Status = status
			n++
		}
	}
	return n, nil
}
func (f *fakePlayRepo) UpdatePlayURLDomainBySource(ctx context.Context, playSourceID uint32, oldHost, newHost string) (int, error) {
	return 0, nil
}

func (f *fakePlayRepo) WithTx(tx bun.Tx) repository.PlayRepository { return f }

type videoRepoStub struct {
	videos map[uint32]*model.Videos
}

func (v *videoRepoStub) List(ctx context.Context, f repository.VideoListFilter) ([]model.Videos, int, error) {
	return nil, 0, nil
}
func (v *videoRepoStub) GetByID(ctx context.Context, id uint32) (*model.Videos, error) {
	item, ok := v.videos[id]
	if !ok {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}
func (v *videoRepoStub) GetByIDs(ctx context.Context, ids []uint32) ([]model.Videos, error) {
	return nil, nil
}
func (v *videoRepoStub) Create(ctx context.Context, video *model.Videos) error { return nil }
func (v *videoRepoStub) BatchCreate(ctx context.Context, videos []*model.Videos) error {
	return nil
}
func (v *videoRepoStub) Update(ctx context.Context, video *model.Videos) error { return nil }
func (v *videoRepoStub) SoftDelete(ctx context.Context, id uint32) error       { return nil }
func (v *videoRepoStub) ReplaceDirectors(ctx context.Context, videoID uint32, directorIDs []uint32) error {
	return nil
}
func (v *videoRepoStub) ReplaceActors(ctx context.Context, videoID uint32, actors []model.VideoActors) error {
	return nil
}
func (v *videoRepoStub) ReplaceTags(ctx context.Context, videoID uint32, tagIDs []uint32) error {
	return nil
}
func (v *videoRepoStub) ListDirectorIDs(ctx context.Context, videoID uint32) ([]uint32, error) {
	return nil, nil
}
func (v *videoRepoStub) ListActorRels(ctx context.Context, videoID uint32) ([]model.VideoActors, error) {
	return nil, nil
}
func (v *videoRepoStub) ListTagIDs(ctx context.Context, videoID uint32) ([]uint32, error) {
	return nil, nil
}
func (v *videoRepoStub) RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return nil
}
func (v *videoRepoStub) WithTx(tx bun.Tx) repository.VideoRepository { return v }
func (v *videoRepoStub) UpdateRatingStats(ctx context.Context, videoID uint32, rating float64, count uint32) error {
	return nil
}
func (v *videoRepoStub) BatchUpdatePublishStatus(ctx context.Context, ids []uint32, status uint8) (int, error) {
	return 0, nil
}
func (v *videoRepoStub) BatchSoftDelete(ctx context.Context, ids []uint32) (int, error) {
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
func (v *videoRepoStub) CountOnlineForSitemap(ctx context.Context) (int, error) {
	return 0, nil
}
func (v *videoRepoStub) ListOnlineForSitemap(ctx context.Context, afterID uint32, limit int) ([]repository.SitemapVideoRow, error) {
	return nil, nil
}
func (v *videoRepoStub) OnlineIDAtOffset(ctx context.Context, offset int) (uint32, bool, error) {
	return 0, false, nil
}
func (v *videoRepoStub) ListTagsByVideoIDs(ctx context.Context, videoIDs []uint32) ([]repository.VideoTagRow, error) {
	return nil, nil
}
func (v *videoRepoStub) UpdateCoverDomainByCollectSource(ctx context.Context, collectSourceID uint32, oldHost, newHost string) (int, error) {
	return 0, nil
}

func TestPlayService_CreateEpisodeSuccess(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint32]*model.Videos{1: {ID: 1, Title: "v"}}}, nil)

	resp, err := svc.CreateEpisode(context.Background(), &dto.CreatePlayEpisodeRequest{
		SourceID: 1, VideoID: 1, EpisodeNumber: 1, PlayURL: "https://example.com/1.m3u8", Format: "hls",
	})
	require.NoError(t, err)
	require.Equal(t, uint32(1), resp.ID)
	require.Equal(t, "https://example.com/1.m3u8", resp.PlayURL)
}

func TestPlayService_CreateEpisodeRejectsActiveDuplicate(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	playRepo.episodes[9] = &model.PlayEpisodes{
		ID: 9, VideoID: 1, SourceID: 1, EpisodeNumber: 1, Title: "旧", PlayURL: "old", Format: "hls",
	}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint32]*model.Videos{1: {ID: 1, Title: "v"}}}, nil)

	_, err := svc.CreateEpisode(context.Background(), &dto.CreatePlayEpisodeRequest{
		SourceID: 1, VideoID: 1, EpisodeNumber: 1, PlayURL: "https://example.com/1.m3u8", Format: "hls",
	})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.PlayEpisodeDuplicate.Code, code.Code)
}

func TestPlayService_BatchUpdateEpisodeStatus(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	playRepo.episodes[10] = &model.PlayEpisodes{ID: 10, VideoID: 1, SourceID: 1, EpisodeNumber: 1, Format: "hls", Status: 1}
	playRepo.episodes[11] = &model.PlayEpisodes{ID: 11, VideoID: 1, SourceID: 1, EpisodeNumber: 2, Format: "hls", Status: 1}
	playRepo.episodes[12] = &model.PlayEpisodes{ID: 12, VideoID: 1, SourceID: 2, EpisodeNumber: 1, Format: "hls", Status: 1}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint32]*model.Videos{1: {ID: 1, Title: "v"}}}, nil)

	// 下架 source 1 的全部剧集
	resp, err := svc.BatchUpdateEpisodeStatus(context.Background(), &dto.BatchUpdateEpisodeStatusRequest{
		VideoID: 1, SourceID: 1, Status: 0,
	})
	require.NoError(t, err)
	require.Equal(t, 2, resp.Affected)
	require.Equal(t, uint8(0), playRepo.episodes[10].Status)
	require.Equal(t, uint8(0), playRepo.episodes[11].Status)
	// source 2 不受影响
	require.Equal(t, uint8(1), playRepo.episodes[12].Status)

	// 上架 source 1 的全部剧集
	resp, err = svc.BatchUpdateEpisodeStatus(context.Background(), &dto.BatchUpdateEpisodeStatusRequest{
		VideoID: 1, SourceID: 1, Status: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 2, resp.Affected)
	require.Equal(t, uint8(1), playRepo.episodes[10].Status)
	require.Equal(t, uint8(1), playRepo.episodes[11].Status)
}

func TestPlayService_BatchUpdateEpisodeStatusRejectsMissingRefs(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint32]*model.Videos{1: {ID: 1, Title: "v"}}}, nil)

	// video 不存在
	_, err := svc.BatchUpdateEpisodeStatus(context.Background(), &dto.BatchUpdateEpisodeStatusRequest{
		VideoID: 99, SourceID: 1, Status: 0,
	})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.VideoNotFound.Code, code.Code)

	// source 不存在
	_, err = svc.BatchUpdateEpisodeStatus(context.Background(), &dto.BatchUpdateEpisodeStatusRequest{
		VideoID: 1, SourceID: 99, Status: 0,
	})
	require.Error(t, err)
	code, ok = errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.PlaySourceNotFound.Code, code.Code)
}
