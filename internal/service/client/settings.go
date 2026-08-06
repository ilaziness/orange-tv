package client

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/service"
)

// ClientSettingsService manages system settings for the client API surface.
// Only client-visible groups (site/feature) are exposed via the handler's binding tag.
type ClientSettingsService interface {
	// GetByGroups returns settings for one or more groups.
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
	maps, err := s.shared.LoadGroupMaps(ctx, groups)
	if err != nil {
		return nil, err
	}
	return s.shared.MapGroupsToResponse(groups, maps)
}
