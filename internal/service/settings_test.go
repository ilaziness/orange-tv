package service

import (
	"encoding/json"
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestParsePlatformFlags(t *testing.T) {
	t.Parallel()

	t.Run("empty uses default", func(t *testing.T) {
		got := ParsePlatformFlags("", true)
		assert.Equal(t, dto.PlatformFlags{Web: true, Desktop: true, App: true, TV: true}, got)
	})

	t.Run("invalid json uses default", func(t *testing.T) {
		got := ParsePlatformFlags("not-json", false)
		assert.Equal(t, dto.PlatformFlags{}, got)
	})

	t.Run("partial fills missing with default", func(t *testing.T) {
		got := ParsePlatformFlags(`{"web":true,"app":false}`, true)
		assert.Equal(t, dto.PlatformFlags{Web: true, Desktop: true, App: false, TV: true}, got)
	})

	t.Run("full matrix", func(t *testing.T) {
		got := ParsePlatformFlags(`{"web":false,"desktop":true,"app":false,"tv":true}`, false)
		assert.Equal(t, dto.PlatformFlags{Web: false, Desktop: true, App: false, TV: true}, got)
	})
}

func TestPlatformBoolVal(t *testing.T) {
	t.Parallel()

	m := map[string]model.SystemSettings{
		constant.SettingFeatureLiveTVEnabled: {
			SettingValue: `{"web":true,"desktop":false,"app":true,"tv":false}`,
			SettingType:  constant.SettingTypeJSON,
		},
	}

	assert.True(t, PlatformBoolVal(m, constant.SettingFeatureLiveTVEnabled, constant.ClientTypeWeb, false))
	assert.False(t, PlatformBoolVal(m, constant.SettingFeatureLiveTVEnabled, constant.ClientTypeDesktop, false))
	assert.True(t, PlatformBoolVal(m, constant.SettingFeatureLiveTVEnabled, constant.ClientTypeApp, false))
	assert.False(t, PlatformBoolVal(m, constant.SettingFeatureLiveTVEnabled, constant.ClientTypeTV, false))
	assert.False(t, PlatformBoolVal(m, "missing", constant.ClientTypeWeb, false))
	assert.True(t, PlatformBoolVal(m, "missing", constant.ClientTypeWeb, true))
}

func TestMapToFeatureSettings(t *testing.T) {
	t.Parallel()

	m := map[string]model.SystemSettings{
		constant.SettingFeatureLiveTVEnabled:  {SettingValue: `{"web":false,"desktop":true,"app":false,"tv":true}`},
		constant.SettingFeatureCommentEnabled: {SettingValue: `{"web":true,"desktop":true,"app":false,"tv":false}`},
		constant.SettingFeatureCommentReview:  {SettingValue: `{"web":false,"desktop":false,"app":true,"tv":true}`},
		constant.SettingFeatureRatingEnabled:  {SettingValue: `{"web":true,"desktop":false,"app":true,"tv":false}`},
	}

	web := MapToFeatureSettings(m, constant.ClientTypeWeb)
	assert.Equal(t, dto.FeatureSettings{
		LiveTVEnabled: false, CommentEnabled: true, CommentReview: false, RatingEnabled: true,
	}, web)

	app := MapToFeatureSettings(m, constant.ClientTypeApp)
	assert.Equal(t, dto.FeatureSettings{
		LiveTVEnabled: false, CommentEnabled: false, CommentReview: false, RatingEnabled: true,
	}, app)
}

func TestMarshalPlatformFlags(t *testing.T) {
	t.Parallel()
	got := MarshalPlatformFlags(dto.PlatformFlags{Web: true, Desktop: false, App: true, TV: false})
	assert.JSONEq(t, `{"web":true,"desktop":false,"app":true,"tv":false}`, got)
}

func TestDecodePlatformFlags(t *testing.T) {
	t.Parallel()

	t.Run("requires all platforms", func(t *testing.T) {
		_, err := DecodePlatformFlags(json.RawMessage(`{"web":true,"app":false}`))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "desktop")
	})

	t.Run("rejects null platform", func(t *testing.T) {
		_, err := DecodePlatformFlags(json.RawMessage(`{"web":true,"desktop":null,"app":false,"tv":true}`))
		assert.Error(t, err)
	})

	t.Run("rejects non-boolean", func(t *testing.T) {
		_, err := DecodePlatformFlags(json.RawMessage(`{"web":1,"desktop":true,"app":false,"tv":true}`))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "布尔")
	})

	t.Run("full matrix ok", func(t *testing.T) {
		got, err := DecodePlatformFlags(json.RawMessage(`{"web":false,"desktop":true,"app":false,"tv":true}`))
		assert.NoError(t, err)
		assert.Equal(t, dto.PlatformFlags{Web: false, Desktop: true, App: false, TV: true}, got)
	})
}

func TestMapToFeatureMatrixNormalizesReview(t *testing.T) {
	t.Parallel()

	m := map[string]model.SystemSettings{
		constant.SettingFeatureCommentEnabled: {SettingValue: `{"web":false,"desktop":true,"app":false,"tv":true}`},
		constant.SettingFeatureCommentReview:  {SettingValue: `{"web":true,"desktop":true,"app":true,"tv":true}`},
	}
	got := MapToFeatureMatrix(m)
	assert.Equal(t, dto.PlatformFlags{Web: false, Desktop: true, App: false, TV: true}, got.CommentReview)
}
