package admin

import (
	"context"
	"testing"
	"time"

	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/require"
)

func TestPlayService_UpdateEpisodeReclaimsSoftDeletedKey(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	now := time.Now()
	// active episode #1 on key (v1,s1,1)
	playRepo.episodes[1] = &model.PlayEpisodes{
		ID: 1, VideoID: 1, SourceID: 1, EpisodeNumber: 2, Title: "E2", PlayURL: "a", Format: "hls",
	}
	// soft-deleted occupies key (v1,s1,1)
	playRepo.episodes[2] = &model.PlayEpisodes{
		ID: 2, VideoID: 1, SourceID: 1, EpisodeNumber: 1, Title: "E1-old", PlayURL: "b", Format: "hls", DeletedAt: &now,
	}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint64]*model.Videos{1: {ID: 1, Title: "v"}}})

	epNum := uint32(1)
	resp, err := svc.UpdateEpisode(context.Background(), 1, &dto.UpdatePlayEpisodeRequest{EpisodeNumber: &epNum})
	require.NoError(t, err)
	require.Equal(t, uint32(1), resp.EpisodeNumber)
	// soft-deleted key owner should be hard-deleted
	_, ok := playRepo.episodes[2]
	require.False(t, ok)
}

func TestPlayService_UpdateEpisodeRejectsActiveKeyConflict(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	playRepo.episodes[1] = &model.PlayEpisodes{
		ID: 1, VideoID: 1, SourceID: 1, EpisodeNumber: 1, Title: "E1", PlayURL: "a", Format: "hls",
	}
	playRepo.episodes[2] = &model.PlayEpisodes{
		ID: 2, VideoID: 1, SourceID: 1, EpisodeNumber: 2, Title: "E2", PlayURL: "b", Format: "hls",
	}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint64]*model.Videos{1: {ID: 1, Title: "v"}}})

	epNum := uint32(1)
	_, err := svc.UpdateEpisode(context.Background(), 2, &dto.UpdatePlayEpisodeRequest{EpisodeNumber: &epNum})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.PlayEpisodeDuplicate.Code, code.Code)
}
