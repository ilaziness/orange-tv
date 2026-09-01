package client

import (
	"context"
	"testing"

	"github.com/ilaziness/orange-tv/internal/clienttype"
	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiveTVService_List_CategoryFilter(t *testing.T) {
	svc := newTestLiveTVService(t)
	ctx := clienttype.WithContext(context.Background(), constant.ClientTypeApp)

	items, total, err := svc.List(ctx, &clientdto.LiveTVListRequest{Category: "央视"})
	require.NoError(t, err)
	// total 为全部在线频道数，category 仅过滤返回列表
	assert.Equal(t, 2, total)
	require.Len(t, items, 1)
	assert.Equal(t, "CCTV1", items[0].Name)
	assert.NotEmpty(t, items[0].StreamURL)
}

func TestLiveTVService_List_UnknownContextDefaultsToWeb(t *testing.T) {
	svc := newTestLiveTVService(t)
	items, _, err := svc.List(context.Background(), &clientdto.LiveTVListRequest{})
	require.NoError(t, err)
	for _, item := range items {
		assert.Empty(t, item.StreamURL, "unknown client type should default to web (no stream_url)")
	}
}

func TestLiveTVService_List_StreamURLByClientType(t *testing.T) {
	tests := []struct {
		name       string
		clientType string
		wantStream bool
	}{
		{"web no stream", constant.ClientTypeWeb, false},
		{"app stream", constant.ClientTypeApp, true},
		{"tv stream", constant.ClientTypeTV, true},
		{"desktop stream", constant.ClientTypeDesktop, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestLiveTVService(t)
			ctx := clienttype.WithContext(context.Background(), tt.clientType)

			items, total, err := svc.List(ctx, &clientdto.LiveTVListRequest{})
			require.NoError(t, err)
			require.Equal(t, 2, total) // 仅在线频道，禁用频道被过滤
			require.Len(t, items, 2)

			for _, item := range items {
				if tt.wantStream {
					assert.NotEmpty(t, item.StreamURL, "client_type=%s should include stream_url", tt.clientType)
				} else {
					assert.Empty(t, item.StreamURL, "client_type=%s should hide stream_url", tt.clientType)
				}
			}
		})
	}
}
