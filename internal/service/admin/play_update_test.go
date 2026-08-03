package admin

import (
	"context"
	"testing"

	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/require"
)

func TestPlayService_UpdateEpisodeRejectsKeyConflict(t *testing.T) {
	playRepo := newFakePlayRepo()
	playRepo.sources[1] = &model.PlaySources{ID: 1, Name: "源1", Status: 1}
	// episode #1 occupies key (v1,s1,1)
	playRepo.episodes[1] = &model.PlayEpisodes{
		ID: 1, VideoID: 1, SourceID: 1, EpisodeNumber: 1, Title: "E1", PlayURL: "a", Format: "hls",
	}
	// episode #2 on a different key (v1,s1,2)
	playRepo.episodes[2] = &model.PlayEpisodes{
		ID: 2, VideoID: 1, SourceID: 1, EpisodeNumber: 2, Title: "E2", PlayURL: "b", Format: "hls",
	}
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint32]*model.Videos{1: {ID: 1, Title: "v"}}}, nil)

	// Moving episode #2 to key (v1,s1,1) conflicts with episode #1
	epNum := uint32(1)
	_, err := svc.UpdateEpisode(context.Background(), 2, &dto.UpdatePlayEpisodeRequest{EpisodeNumber: &epNum})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.PlayEpisodeDuplicate.Code, code.Code)
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
	svc := NewPlayService(playRepo, &videoRepoStub{videos: map[uint32]*model.Videos{1: {ID: 1, Title: "v"}}}, nil)

	epNum := uint32(1)
	_, err := svc.UpdateEpisode(context.Background(), 2, &dto.UpdatePlayEpisodeRequest{EpisodeNumber: &epNum})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.PlayEpisodeDuplicate.Code, code.Code)
}
