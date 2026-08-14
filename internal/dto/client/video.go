package client

import "github.com/ilaziness/orange-tv/internal/dto"

// VideoListRequest filters public video list.
type VideoListRequest struct {
	dto.PaginationRequest
	// 分类ID筛选
	CategoryID uint32 `form:"category_id"`
	// 父分类ID筛选
	ParentCategoryID uint32 `form:"parent_category_id"`
	// 起始年份筛选
	YearStart uint32 `form:"year_start"`
	// 结束年份筛选
	YearEnd uint32 `form:"year_end"`
	// 地区筛选
	Region string `form:"region"`
	// 排序方式（如 latest=最新，rating=评分）
	Sort string `form:"sort"`
}

// SearchRequest is the public search query with optional filters.
type SearchRequest struct {
	dto.PaginationRequest
	// 搜索关键词（1-10 字，必填）
	Keyword string `form:"keyword" binding:"required,min=1,max=10,search"`
	// 分类ID筛选
	CategoryID uint32 `form:"category_id"`
	// 父分类ID筛选
	ParentCategoryID uint32 `form:"parent_category_id"`
	// 起始年份筛选
	YearStart uint32 `form:"year_start"`
	// 结束年份筛选
	YearEnd uint32 `form:"year_end"`
	// 地区筛选
	Region string `form:"region"`
	// 排序方式（如 latest=最新，rating=评分）
	Sort string `form:"sort"`
}

// RelatedRequest loads related videos for a detail page.
type RelatedRequest struct {
	// 返回数量（1-50，默认由服务端控制）
	Limit int `form:"limit" binding:"omitempty,min=1,max=50"`
}

// EpisodeURI captures route params for single-episode play URL.
type EpisodeURI struct {
	// 视频ID（路径参数）
	ID uint32 `uri:"id" binding:"required,gt=0"`
	// 播放源ID（路径参数）
	SourceID uint32 `uri:"source_id" binding:"required,gt=0"`
	// 剧集ID（路径参数）
	EpisodeID uint32 `uri:"episode_id" binding:"required,gt=0"`
}

// VideoListItem is a compact video card payload for client (no publish_status, no timestamps).
type VideoListItem struct {
	// 视频ID
	ID uint32 `json:"id"`
	// 视频标题
	Title string `json:"title"`
	// 副标题
	Subtitle string `json:"subtitle"`
	// 封面地址
	Cover string `json:"cover"`
	// 海报地址
	Poster string `json:"poster"`
	// 上映年份
	Year uint32 `json:"year"`
	// 地区
	Region string `json:"region"`
	// 语言
	Language string `json:"language"`
	// 评分
	Rating float64 `json:"rating"`
	// 分类ID
	CategoryID uint32 `json:"category_id"`
	// 连载状态（1=连载中，2=已完结，3=未知）
	SerialStatus uint8 `json:"serial_status"`
	// 总时长（分钟）
	Duration uint32 `json:"duration"`
	// 播放量
	ViewCount uint32 `json:"view_count"`
	// 标签列表
	Tags []dto.NamedItem `json:"tags,omitempty"`
}

// VideoDetailEpisode is an episode summary without play URL (client detail API).
type VideoDetailEpisode struct {
	// 剧集记录ID
	ID uint32 `json:"id"`
	// 集数
	Episode uint32 `json:"episode"`
	// 剧集标题
	Title string `json:"title"`
}

// VideoDetailSourceGroup groups episode summaries by play source (client detail API, no URL).
type VideoDetailSourceGroup struct {
	// 播放源ID
	ID uint32 `json:"id"`
	// 播放源名称
	Name string `json:"name"`
	// 剧集列表（不含播放地址）
	Episodes []VideoDetailEpisode `json:"episodes"`
}

// VideoDetailResponse is the client video detail payload (no play URLs).
type VideoDetailResponse struct {
	// 视频ID
	ID uint32 `json:"id"`
	// 视频标题
	Title string `json:"title"`
	// 副标题
	Subtitle string `json:"subtitle"`
	// 剧情简介
	Description string `json:"description"`
	// 分类ID
	CategoryID uint32 `json:"category_id"`
	// 连载状态（1=连载中，2=已完结，3=未知）
	SerialStatus uint8 `json:"serial_status"`
	// 封面地址
	Cover string `json:"cover"`
	// 海报地址
	Poster string `json:"poster"`
	// 上映年份
	Year uint32 `json:"year"`
	// 地区
	Region string `json:"region"`
	// 语言
	Language string `json:"language"`
	// 总时长（分钟）
	Duration uint32 `json:"duration"`
	// 上映日期
	ReleaseDate string `json:"release_date,omitempty"`
	// 评分
	Rating float64 `json:"rating"`
	// 评分人数
	RatingCount uint32 `json:"rating_count"`
	// 播放量
	ViewCount uint32 `json:"view_count"`
	// 导演列表
	Directors []dto.NamedItem `json:"directors"`
	// 演员列表
	Actors []dto.NamedItem `json:"actors"`
	// 标签列表
	Tags []dto.NamedItem `json:"tags"`
	// 播放源分组列表（不含播放地址）
	Sources []VideoDetailSourceGroup `json:"sources"`
}

// PlayEpisodeResponse is the single-episode play URL response.
type PlayEpisodeResponse struct {
	// 播放地址
	URL string `json:"url"`
	// 清晰度（如 高清、超清）
	Quality string `json:"quality"`
	// 播放格式（如 hls、mp4）
	Format string `json:"format"`
}
