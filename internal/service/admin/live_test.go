package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubLiveSyncSettingsService struct {
	m         map[string]map[string]model.SystemSettings
	loadErr   error
	upsertErr error
}

func (s *stubLiveSyncSettingsService) LoadMapByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error) {
	if s.m == nil {
		return nil, s.loadErr
	}
	return s.m[group], s.loadErr
}

func (s *stubLiveSyncSettingsService) LoadGroupMaps(ctx context.Context, groups []string) (map[string]map[string]model.SystemSettings, error) {
	return s.m, s.loadErr
}

func (s *stubLiveSyncSettingsService) MapGroupToResponse(group string, m map[string]model.SystemSettings) (any, error) {
	return nil, nil
}

func (s *stubLiveSyncSettingsService) MapGroupsToResponse(groups []string, maps map[string]map[string]model.SystemSettings) (any, error) {
	return nil, nil
}

func (s *stubLiveSyncSettingsService) UpsertMany(ctx context.Context, group string, upserts []repository.SettingUpsert) error {
	if s.upsertErr != nil {
		return s.upsertErr
	}
	if s.m == nil {
		s.m = make(map[string]map[string]model.SystemSettings)
	}
	if s.m[group] == nil {
		s.m[group] = make(map[string]model.SystemSettings)
	}
	for _, u := range upserts {
		s.m[group][u.Key] = model.SystemSettings{
			SettingKey:   u.Key,
			SettingGroup: u.Group,
			SettingValue: u.Value,
			SettingType:  u.SettingType,
			Description:  u.Description,
		}
	}
	return nil
}

func TestLiveService_GetAndSaveSyncSourceURL(t *testing.T) {
	tests := []struct {
		name      string
		loadErr   error
		upsertErr error
		steps     func(t *testing.T, svc LiveService, settings *stubLiveSyncSettingsService)
	}{
		{
			name: "save trims and get returns saved url",
			steps: func(t *testing.T, svc LiveService, settings *stubLiveSyncSettingsService) {
				ctx := context.Background()

				url, err := svc.GetSyncSourceURL(ctx)
				require.NoError(t, err)
				assert.Empty(t, url)

				require.NoError(t, svc.SaveSyncSourceURL(ctx, "  https://example.com/live.txt  "))

				url, err = svc.GetSyncSourceURL(ctx)
				require.NoError(t, err)
				assert.Equal(t, "https://example.com/live.txt", url)

				require.NotNil(t, settings.m[constant.SettingGroupLive][constant.SettingLiveSyncSourceURL])
				assert.Equal(t, "https://example.com/live.txt", settings.m[constant.SettingGroupLive][constant.SettingLiveSyncSourceURL].SettingValue)
			},
		},
		{
			name: "save whitespace only returns param error",
			steps: func(t *testing.T, svc LiveService, _ *stubLiveSyncSettingsService) {
				err := svc.SaveSyncSourceURL(context.Background(), "   ")
				require.Error(t, err)
				var code *errcode.Code
				require.True(t, errors.As(err, &code))
				assert.Equal(t, errcode.ParamError.Code, code.Code)
			},
		},
		{
			name:    "load error returns database error",
			loadErr: errcode.DatabaseError,
			steps: func(t *testing.T, svc LiveService, _ *stubLiveSyncSettingsService) {
				_, err := svc.GetSyncSourceURL(context.Background())
				require.Error(t, err)
				assert.ErrorIs(t, err, errcode.DatabaseError)
			},
		},
		{
			name:      "upsert error is returned",
			upsertErr: errcode.DatabaseError,
			steps: func(t *testing.T, svc LiveService, _ *stubLiveSyncSettingsService) {
				err := svc.SaveSyncSourceURL(context.Background(), "https://example.com/live.txt")
				require.Error(t, err)
				assert.ErrorIs(t, err, errcode.DatabaseError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &stubLiveSyncSettingsService{loadErr: tt.loadErr, upsertErr: tt.upsertErr}
			svc := NewLiveService(nil, nil, settings, zap.NewNop())
			tt.steps(t, svc, settings)
		})
	}
}
