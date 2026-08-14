package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// CreatePlaySourceRequest creates a global play source.
type CreatePlaySourceRequest struct {
	// 播放源名称（必填）
	Name string `json:"name" binding:"required,min=1,max=100"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdatePlaySourceRequest updates a global play source.
type UpdatePlaySourceRequest struct {
	// 播放源名称
	Name *string `json:"name" binding:"omitempty,min=1,max=100"`
	// 排序权重，值越小越靠前
	SortOrder *uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// PlaySourceResponse is a play source payload.
type PlaySourceResponse struct {
	// 播放源ID
	ID uint32 `json:"id"`
	// 播放源名称
	Name string `json:"name"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
}

// PlayEpisodeListRequest filters episode list.
type PlayEpisodeListRequest struct {
	dto.PaginationRequest
	// 视频ID（必填）
	VideoID uint32 `form:"video_id" binding:"required,min=1"`
	// 播放源ID（必填）
	SourceID uint32 `form:"source_id" binding:"required,min=1"`
}

// CreatePlayEpisodeRequest creates a play episode.
type CreatePlayEpisodeRequest struct {
	// 播放源ID（必填）
	SourceID uint32 `json:"source_id" binding:"required,min=1"`
	// 视频ID（必填）
	VideoID uint32 `json:"video_id" binding:"required,min=1"`
	// 集数（必填）
	EpisodeNumber uint32 `json:"episode_number" binding:"required,min=1"`
	// 剧集标题
	Title string `json:"title" binding:"omitempty,max=255"`
	// 播放地址（必填）
	PlayURL string `json:"play_url" binding:"required,min=1,max=1000"`
	// 清晰度（如 高清、超清）
	Quality string `json:"quality" binding:"omitempty,max=50"`
	// 播放格式（必填：hls=HLS，mp4=MP4，dash=DASH，flv=FLV）
	Format string `json:"format" binding:"required,oneof=hls mp4 dash flv"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdatePlayEpisodeRequest updates a play episode.
type UpdatePlayEpisodeRequest struct {
	// 播放源ID
	SourceID *uint32 `json:"source_id" binding:"omitempty,min=1"`
	// 视频ID
	VideoID *uint32 `json:"video_id" binding:"omitempty,min=1"`
	// 集数
	EpisodeNumber *uint32 `json:"episode_number" binding:"omitempty,min=1"`
	// 剧集标题
	Title *string `json:"title" binding:"omitempty,max=255"`
	// 播放地址
	PlayURL *string `json:"play_url" binding:"omitempty,min=1,max=1000"`
	// 清晰度（如 高清、超清）
	Quality *string `json:"quality" binding:"omitempty,max=50"`
	// 播放格式（hls=HLS，mp4=MP4，dash=DASH，flv=FLV）
	Format *string `json:"format" binding:"omitempty,oneof=hls mp4 dash flv"`
	// 排序权重，值越小越靠前
	SortOrder *uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// PlayEpisodeResponse is a play episode payload.
type PlayEpisodeResponse struct {
	// 剧集记录ID
	ID uint32 `json:"id"`
	// 播放源ID
	SourceID uint32 `json:"source_id"`
	// 视频ID
	VideoID uint32 `json:"video_id"`
	// 集数
	EpisodeNumber uint32 `json:"episode_number"`
	// 剧集标题
	Title string `json:"title"`
	// 播放地址
	PlayURL string `json:"play_url"`
	// 清晰度
	Quality string `json:"quality"`
	// 播放格式
	Format string `json:"format"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
}

// BatchUpdateEpisodeStatusRequest 批量更新某影视下指定播放源的全部剧集状态。
type BatchUpdateEpisodeStatusRequest struct {
	// 视频ID（必填）
	VideoID uint32 `json:"video_id" binding:"required,min=1"`
	// 播放源ID（必填）
	SourceID uint32 `json:"source_id" binding:"required,min=1"`
	// 目标状态（0=禁用，1=启用）
	Status uint8 `json:"status" binding:"oneof=0 1"`
}

// BatchUpdateEpisodeStatusResponse 批量更新剧集状态响应。
type BatchUpdateEpisodeStatusResponse struct {
	// 受影响的行数
	Affected int `json:"affected"`
}
