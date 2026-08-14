package admin

// TargetField 定义批量更新支持的目标字段。
const (
	TargetVideoCover = "video_cover"
	TargetEpisodeURL = "episode_url"
)

// BackupQuery binds the backup query parameter.
type BackupQuery struct {
	// 是否使用原生 SQL 备份（true/false）
	Native string `form:"native"`
}

// BatchUpdatePreviewRequest 批量更新预览请求。
type BatchUpdatePreviewRequest struct {
	// 目标字段（video_cover=视频封面，episode_url=剧集地址，必填）
	Target string `json:"target" binding:"required,oneof=video_cover episode_url"`
	// 旧值（必填）
	OldValue string `json:"old_value" binding:"required,max=2000"`
}

// BatchUpdatePreviewResponse 批量更新预览响应。
type BatchUpdatePreviewResponse struct {
	// 匹配到的数据行数
	MatchedRows int64 `json:"matched_rows"`
}

// BatchUpdateExecuteRequest 批量更新执行请求。
type BatchUpdateExecuteRequest struct {
	// 目标字段（video_cover=视频封面，episode_url=剧集地址，必填）
	Target string `json:"target" binding:"required,oneof=video_cover episode_url"`
	// 旧值（必填）
	OldValue string `json:"old_value" binding:"required,max=2000"`
	// 新值（必填）
	NewValue string `json:"new_value" binding:"required,max=2000"`
}

// BatchUpdateExecuteResponse 批量更新执行响应。
type BatchUpdateExecuteResponse struct {
	// 更新成功的数据行数
	UpdatedRows int64 `json:"updated_rows"`
}
