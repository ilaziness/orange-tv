package admin

// TargetField 定义批量更新支持的目标字段。
const (
	TargetVideoCover = "video_cover"
	TargetEpisodeURL = "episode_url"
)

// BatchUpdatePreviewRequest 批量更新预览请求。
type BatchUpdatePreviewRequest struct {
	Target   string `json:"target" validate:"required,oneof=video_cover episode_url"`
	OldValue string `json:"old_value" validate:"required,max=2000"`
}

// BatchUpdatePreviewResponse 批量更新预览响应。
type BatchUpdatePreviewResponse struct {
	MatchedRows int64 `json:"matched_rows"`
}

// BatchUpdateExecuteRequest 批量更新执行请求。
type BatchUpdateExecuteRequest struct {
	Target   string `json:"target" validate:"required,oneof=video_cover episode_url"`
	OldValue string `json:"old_value" validate:"required,max=2000"`
	NewValue string `json:"new_value" validate:"required,max=2000"`
}

// BatchUpdateExecuteResponse 批量更新执行响应。
type BatchUpdateExecuteResponse struct {
	UpdatedRows int64 `json:"updated_rows"`
}
