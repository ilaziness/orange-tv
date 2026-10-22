package open

import (
	"context"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"go.uber.org/zap"
)

// SRSHookBody is the JSON SRS posts to http_hooks.
type SRSHookBody struct {
	Action string `json:"action"`
	App    string `json:"app"`
	Stream string `json:"stream"`
	Param  string `json:"param"`
}

// LiveHookService authorizes self-hosted SRS callbacks and Nginx play auth.
type LiveHookService interface {
	SRSHook(ctx context.Context, hookSecret string, body SRSHookBody) error
	PlayAuth(ctx context.Context, hookSecret, originalURI string) error
}

type liveHookService struct {
	repo     repository.LiveRepository
	settings service.SettingsService
	log      *zap.Logger
}

// NewLiveHookService creates a LiveHookService.
func NewLiveHookService(repo repository.LiveRepository, settings service.SettingsService, log *zap.Logger) LiveHookService {
	return &liveHookService{repo: repo, settings: settings, log: log}
}

func (s *liveHookService) SRSHook(ctx context.Context, hookSecret string, body SRSHookBody) error {
	cfg, err := s.load(ctx)
	if err != nil {
		return err
	}
	// Reject bad secrets before any room lookup to avoid unauthenticated DB probes.
	if err := service.VerifyLiveHookSecret(cfg, hookSecret); err != nil {
		s.log.Info("live hook: srs rejected",
			zap.String("action", body.Action),
			zap.String("app", body.App),
			zap.String("stream", body.Stream),
			zap.Error(err),
		)
		return err
	}
	if body.Action == "on_unpublish" {
		s.log.Info("live hook: srs unpublish", zap.String("app", body.App), zap.String("stream", body.Stream))
		return nil
	}
	var room *model.LiveRooms
	if body.Stream != "" {
		room, err = s.repo.GetRoomByStreamKey(ctx, body.Stream)
		if err != nil {
			s.log.Error("live hook: get room failed", zap.String("stream", body.Stream), zap.Error(err))
			return errcode.Wrap(errcode.DatabaseError, err)
		}
	}
	if err := service.DecideSRSHook(cfg, hookSecret, body.Action, body.App, body.Stream, body.Param, room, time.Now()); err != nil {
		s.log.Info("live hook: srs rejected",
			zap.String("action", body.Action),
			zap.String("app", body.App),
			zap.String("stream", body.Stream),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *liveHookService) PlayAuth(ctx context.Context, hookSecret, originalURI string) error {
	cfg, err := s.load(ctx)
	if err != nil {
		return err
	}
	err = service.DecidePlayAuth(cfg, hookSecret, originalURI, func(stream string) *model.LiveRooms {
		room, lookupErr := s.repo.GetRoomByStreamKey(ctx, stream)
		if lookupErr != nil {
			s.log.Error("live play auth: get room failed", zap.String("stream", stream), zap.Error(lookupErr))
			return nil
		}
		return room
	}, time.Now())
	if err != nil {
		s.log.Info("live play auth: rejected", zap.String("uri", originalURI), zap.Error(err))
		return err
	}
	return nil
}

func (s *liveHookService) load(ctx context.Context) (service.LiveStreamConfig, error) {
	m, err := s.settings.LoadMapByGroup(ctx, constant.SettingGroupLive)
	if err != nil {
		s.log.Error("live hook: load settings failed", zap.Error(err))
		return service.LiveStreamConfig{}, err
	}
	return service.ParseLiveStreamConfig(m), nil
}
