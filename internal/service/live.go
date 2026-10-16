package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

const liveStreamKeyBytes = 16 // 32 hex chars

// LiveStreamConfig holds admin-only push/play configuration.
type LiveStreamConfig struct {
	PushServer      string
	PlayHLSTemplate string
	PlayFLVTemplate string
}

// GenerateLiveStreamKey returns a 32+ hex stream key.
func GenerateLiveStreamKey() (string, error) {
	b := make([]byte, liveStreamKeyBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate live stream key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ApplyStreamKeyTemplate replaces {stream_key} in a play URL template.
func ApplyStreamKeyTemplate(template, streamKey string) string {
	template = strings.TrimSpace(template)
	if template == "" || streamKey == "" {
		return ""
	}
	return strings.ReplaceAll(template, "{stream_key}", streamKey)
}

// ParseLiveStreamConfig maps live settings rows to LiveStreamConfig.
func ParseLiveStreamConfig(m map[string]model.SystemSettings) LiveStreamConfig {
	return LiveStreamConfig{
		PushServer:      strings.TrimSpace(StrVal(m, constant.SettingLivePushServer)),
		PlayHLSTemplate: strings.TrimSpace(StrVal(m, constant.SettingLivePlayHLSTemplate)),
		PlayFLVTemplate: strings.TrimSpace(StrVal(m, constant.SettingLivePlayFLVTemplate)),
	}
}

// ReadyForBroadcast reports whether push server and at least one play template are set.
func (c LiveStreamConfig) ReadyForBroadcast() bool {
	if c.PushServer == "" {
		return false
	}
	return c.PlayHLSTemplate != "" || c.PlayFLVTemplate != ""
}

// ResolvePlayURLs returns resolved HLS/FLV URLs for a stream key (empty if template blank).
func (c LiveStreamConfig) ResolvePlayURLs(streamKey string) (hls, flv string) {
	return ApplyStreamKeyTemplate(c.PlayHLSTemplate, streamKey), ApplyStreamKeyTemplate(c.PlayFLVTemplate, streamKey)
}

// ResolvePublicPlayURLs returns play URLs only while the room is living.
// Idle/banned rooms must not expose URLs that embed stream_key.
func (c LiveStreamConfig) ResolvePublicPlayURLs(streamKey string, roomStatus uint8) (hls, flv string) {
	if roomStatus != constant.LiveRoomLiving {
		return "", ""
	}
	return c.ResolvePlayURLs(streamKey)
}

// ResolveStreamerApplyStatus maps DB state to profile streamer_apply_status.
// An approved apply after revoke (is_streamer=0) maps to "none" so the user can re-apply.
func ResolveStreamerApplyStatus(isStreamer uint8, apply *model.StreamerApplies) string {
	if isStreamer == 1 {
		return constant.StreamerApplyStatusApproved
	}
	if apply == nil {
		return constant.StreamerApplyStatusNone
	}
	switch apply.Status {
	case constant.StreamerApplyPending:
		return constant.StreamerApplyStatusPending
	case constant.StreamerApplyRejected:
		return constant.StreamerApplyStatusRejected
	case constant.StreamerApplyApproved:
		return constant.StreamerApplyStatusNone
	default:
		return constant.StreamerApplyStatusNone
	}
}

// NewLiveSessionSnapshot builds a session row from room + user snapshot fields.
func NewLiveSessionSnapshot(room *model.LiveRooms, user *model.Users, startedAt time.Time) *model.LiveSessions {
	return &model.LiveSessions{
		RoomID:           room.ID,
		UserID:           room.UserID,
		Title:            room.Title,
		CategoryID:       room.CategoryID,
		SubCategoryID:    room.SubCategoryID,
		StreamerNickname: user.Nickname,
		StreamerAvatar:   user.Avatar,
		StartedAt:        startedAt,
	}
}

// ValidateLiveCategoryPair checks that categoryID is an enabled L1 and subCategoryID
// matches the enabled-children rule (required L2 when children exist, else must be 0).
func ValidateLiveCategoryPair(ctx context.Context, repo repository.LiveRepository, categoryID, subCategoryID uint32) error {
	if categoryID == 0 {
		return errcode.WithMessage(errcode.ParamError, "请选择直播分类")
	}
	cat, err := repo.GetCategory(ctx, categoryID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if cat == nil || cat.Status != constant.StatusEnabled || cat.ParentID != 0 {
		return errcode.LiveCategoryDisabled
	}
	enabledChildren, err := repo.CountEnabledCategoryChildren(ctx, categoryID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if enabledChildren > 0 {
		if subCategoryID == 0 {
			return errcode.WithMessage(errcode.ParamError, "请选择二级分类")
		}
		sub, err := repo.GetCategory(ctx, subCategoryID)
		if err != nil {
			return errcode.Wrap(errcode.DatabaseError, err)
		}
		if sub == nil || sub.Status != constant.StatusEnabled || sub.ParentID != categoryID {
			return errcode.WithMessage(errcode.LiveCategoryDisabled, "二级分类不可用或不属于所选一级分类")
		}
		return nil
	}
	if subCategoryID != 0 {
		return errcode.WithMessage(errcode.ParamError, "该一级分类无启用子类，二级分类须为0")
	}
	return nil
}

// ResolveLiveCategoryOverride merges request category fields onto current values, then validates.
// Rules: both → use request; only L1 → sub=0 if no enabled children else error; only L2 → error.
func ResolveLiveCategoryOverride(
	ctx context.Context,
	repo repository.LiveRepository,
	reqCategoryID, reqSubCategoryID *uint32,
	currentCategoryID, currentSubCategoryID uint32,
) (categoryID, subCategoryID uint32, err error) {
	if reqCategoryID == nil && reqSubCategoryID == nil {
		return currentCategoryID, currentSubCategoryID, nil
	}
	if reqCategoryID == nil {
		return 0, 0, errcode.WithMessage(errcode.ParamError, "请选择一级分类")
	}
	categoryID = *reqCategoryID
	if reqSubCategoryID != nil {
		subCategoryID = *reqSubCategoryID
	} else {
		enabledChildren, err := repo.CountEnabledCategoryChildren(ctx, categoryID)
		if err != nil {
			return 0, 0, errcode.Wrap(errcode.DatabaseError, err)
		}
		if enabledChildren > 0 {
			return 0, 0, errcode.WithMessage(errcode.ParamError, "请选择二级分类")
		}
		subCategoryID = 0
	}
	if err := ValidateLiveCategoryPair(ctx, repo, categoryID, subCategoryID); err != nil {
		return 0, 0, err
	}
	return categoryID, subCategoryID, nil
}
