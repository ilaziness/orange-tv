package collect

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"go.uber.org/zap"
)

// Collector abstracts format-specific collect logic (Apple CMS / system default).
// The Engine orchestrates pagination, category mapping, upsert and time filtering
// on top of the normalized ListPage / Page / RemoteCategory returned by these methods.
type Collector interface {
	// FetchListPage fetches one page of the source's list endpoint.
	// dataRange filters by time (today/last1d/last3d/last1w/last1m/all).
	FetchListPage(ctx context.Context, source *model.CollectSources, page int, dataRange string) (*ListPage, error)
	// FetchDetail fetches full detail for the given external IDs (caller batches).
	FetchDetail(ctx context.Context, source *model.CollectSources, ids []uint32) (*Page, error)
	// FetchCategories fetches the remote category list (Apple CMS class / Open API categories).
	FetchCategories(ctx context.Context, source *model.CollectSources) ([]RemoteCategory, error)
}

// newCollector returns a Collector implementation for the given source type.
func newCollector(source *model.CollectSources, fetcher *Fetcher, log *zap.Logger) (Collector, error) {
	switch source.Type {
	case constant.CollectTypeDefault:
		return &defaultCollector{fetcher: fetcher, log: log}, nil
	case constant.CollectTypeAppleCMS:
		return &appleCMSCollector{fetcher: fetcher, log: log}, nil
	default:
		return nil, fmt.Errorf("不支持的采集源类型: %d", source.Type)
	}
}

// FetchCategories is an exported helper for admin services to fetch remote
// categories of a source without instantiating an Engine. It builds a throwaway
// Fetcher + Collector for the source type.
func FetchCategories(ctx context.Context, source *model.CollectSources, log *zap.Logger) ([]RemoteCategory, error) {
	fetcher := NewFetcher(log)
	c, err := newCollector(source, fetcher, log)
	if err != nil {
		return nil, err
	}
	return c.FetchCategories(ctx, source)
}
