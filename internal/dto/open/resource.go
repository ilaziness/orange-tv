package open

import (
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
)

// VideoListRequest is the query parameter for the open video list.
type VideoListRequest struct {
	shareddto.PaginationRequest
	// 按创建时间筛选（today/last1d/last3d/last1w/last1m/all），为空表示全部
	DataRange string `form:"data_range" json:"data_range" binding:"omitempty,oneof=today last1d last3d last1w last1m all"`
	// 按播放源名称筛选（精确匹配），为空表示不筛选
	Source string `form:"source" json:"source" binding:"omitempty,max=10"`
}

// VideoDetailRequest is the query parameter for the open video detail (multiple ids).
type VideoDetailRequest struct {
	// 视频ID列表（最多 50 个，可传多个 id 参数）
	IDs []uint32 `form:"id" binding:"required,max=50,dive,gt=0"`
}

// VideoListItem is one compact video item in the open video list.
type VideoListItem struct {
	// 视频ID
	ID uint32 `json:"id"`
	// 视频标题
	Title string `json:"title"`
	// 分类ID
	CategoryID uint32 `json:"category_id"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}

// VideoSource is a play source group used in the open detail response.
type VideoSource struct {
	// 播放源ID
	ID uint32 `json:"id"`
	// 播放源名称
	Name string `json:"name"`
	// 剧集列表（含播放地址）
	Episodes []VideoSourceEpisode `json:"episodes"`
}

// VideoSourceEpisode is one playable episode in the open detail response.
type VideoSourceEpisode struct {
	// 集数
	Episode uint32 `json:"episode"`
	// 剧集标题
	Title string `json:"title"`
	// 播放地址
	URL string `json:"url"`
}

// VideoDetailItem is the full video detail payload in the open API.
type VideoDetailItem struct {
	// 视频ID
	ID uint32 `json:"id"`
	// 视频标题
	Title string `json:"title"`
	// 副标题
	Subtitle string `json:"subtitle"`
	// 封面地址
	Cover string `json:"cover"`
	// 分类ID
	CategoryID uint32 `json:"category_id"`
	// 上映年份
	Year uint32 `json:"year"`
	// 评分
	Rating float64 `json:"rating"`
	// 上映日期
	ReleaseDate string `json:"release_date"`
	// 地区
	Region string `json:"region"`
	// 语言
	Language string `json:"language"`
	// 剧情简介
	Description string `json:"description"`
	// 导演列表
	Directors []string `json:"directors"`
	// 演员列表
	Actors []string `json:"actors"`
	// 播放源分组列表（含播放地址）
	Sources []VideoSource `json:"sources"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}
