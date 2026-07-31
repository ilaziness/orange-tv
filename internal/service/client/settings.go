package client

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/service"
)

// ClientSettingsService manages system settings for the client API surface.
// Only whitelisted groups are exposed; non-whitelisted groups return an error.
type ClientSettingsService interface {
	// GetByGroups returns settings for one or more groups.
	// If any group is not in the client whitelist, an error is returned.
	// Single group → flat DTO; multiple groups → map[string]any.
	GetByGroups(ctx context.Context, groups []string) (any, error)
}

type clientSettingsService struct {
	shared service.SettingsService
}

// NewClientSettingsService creates a ClientSettingsService.
func NewClientSettingsService(shared service.SettingsService) ClientSettingsService {
	return &clientSettingsService{shared: shared}
}

func (s *clientSettingsService) GetByGroups(ctx context.Context, groups []string) (any, error) {
	for _, g := range groups {
		if !constant.IsValidSettingGroup(g) {
			return nil, errcode.WithMessage(errcode.ParamError, "无效的设置分组")
		}
		if !constant.IsClientAllowedGroup(g) {
			return nil, errcode.WithMessage(errcode.ParamError, "无权限访问该设置分组")
		}
	}

	maps, err := s.shared.LoadGroupMaps(ctx, groups)
	if err != nil {
		return nil, err
	}
	return s.shared.MapGroupsToResponse(groups, maps)
}
