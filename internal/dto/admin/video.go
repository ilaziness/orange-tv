package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// VideoListRequest filters admin video list.
type VideoListRequest struct {
	dto.PaginationRequest
	// 关键词搜索（标题）
	Keyword string `form:"keyword"`
	// 分类ID筛选
	CategoryID uint32 `form:"category_id"`
	// 上架状态筛选（0=未上架，1=已上架）
	PublishStatus *uint8 `form:"publish_status"`
	// 年份筛选
	Year uint32 `form:"year"`
	// 地区筛选
	Region string `form:"region"`
	// 语言筛选
	Language string `form:"language"`
	// 导演ID筛选
	DirectorID uint32 `form:"director_id"`
	// 演员ID筛选
	ActorID uint32 `form:"actor_id"`
	// 标签ID筛选
	TagID uint32 `form:"tag_id"`
}

// CreateVideoRequest creates a video with associations.
type CreateVideoRequest struct {
	// 视频标题（必填）
	Title string `json:"title" binding:"required,min=1,max=255"`
	// 副标题
	Subtitle string `json:"subtitle" binding:"omitempty,max=255"`
	// 剧情简介
	Description string `json:"description" binding:"omitempty,max=10000"`
	// 分类ID（必填）
	CategoryID uint32 `json:"category_id" binding:"required,min=1"`
	// 上架状态（0=未上架，1=已上架）
	PublishStatus *uint8 `json:"publish_status" binding:"omitempty,oneof=0 1"`
	// 连载状态（1=连载中，2=已完结，3=未知）
	SerialStatus *uint8 `json:"serial_status" binding:"omitempty,oneof=1 2 3"`
	// 封面地址
	CoverImage string `json:"cover_image" binding:"omitempty,max=500"`
	// 海报地址
	PosterImage string `json:"poster_image" binding:"omitempty,max=500"`
	// 上映年份
	Year uint32 `json:"year" binding:"omitempty,min=0,max=9999"`
	// 地区
	Region string `json:"region" binding:"omitempty,max=50"`
	// 总时长（分钟）
	Duration uint32 `json:"duration" binding:"omitempty,min=0"`
	// 语言
	Language string `json:"language" binding:"omitempty,max=50"`
	// 上映日期
	ReleaseDate string `json:"release_date" binding:"omitempty,max=64"`
	// 导演ID列表
	DirectorIDs []uint32 `json:"director_ids" binding:"omitempty,dive,min=1"`
	// 演员列表
	Actors []VideoActorInput `json:"actors" binding:"omitempty,dive"`
	// 标签ID列表
	TagIDs []uint32 `json:"tag_ids" binding:"omitempty,dive,min=1"`
}

// UpdateVideoRequest updates a video and optional associations.
type UpdateVideoRequest struct {
	// 视频标题
	Title *string `json:"title" binding:"omitempty,min=1,max=255"`
	// 副标题
	Subtitle *string `json:"subtitle" binding:"omitempty,max=255"`
	// 剧情简介
	Description *string `json:"description" binding:"omitempty,max=10000"`
	// 分类ID
	CategoryID *uint32 `json:"category_id" binding:"omitempty,min=1"`
	// 上架状态（0=未上架，1=已上架）
	PublishStatus *uint8 `json:"publish_status" binding:"omitempty,oneof=0 1"`
	// 连载状态（1=连载中，2=已完结，3=未知）
	SerialStatus *uint8 `json:"serial_status" binding:"omitempty,oneof=1 2 3"`
	// 封面地址
	CoverImage *string `json:"cover_image" binding:"omitempty,max=500"`
	// 海报地址
	PosterImage *string `json:"poster_image" binding:"omitempty,max=500"`
	// 上映年份
	Year *uint32 `json:"year" binding:"omitempty,min=0,max=9999"`
	// 地区
	Region *string `json:"region" binding:"omitempty,max=50"`
	// 总时长（分钟）
	Duration *uint32 `json:"duration" binding:"omitempty,min=0"`
	// 语言
	Language *string `json:"language" binding:"omitempty,max=50"`
	// 上映日期
	ReleaseDate *string `json:"release_date" binding:"omitempty,max=64"`
	// 导演ID列表
	DirectorIDs *[]uint32 `json:"director_ids" binding:"omitempty,dive,min=1"`
	// 演员列表
	Actors *[]VideoActorInput `json:"actors" binding:"omitempty,dive"`
	// 标签ID列表
	TagIDs *[]uint32 `json:"tag_ids" binding:"omitempty,dive,min=1"`
}

// VideoActorInput binds actor to video.
type VideoActorInput struct {
	// 演员ID（必填）
	ActorID uint32 `json:"actor_id" binding:"required,min=1"`
}

// VideoListItem is a compact video card payload for admin (includes publish_status, timestamps).
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
	// 分类名称
	CategoryName string `json:"category_name,omitempty"`
	// 上架状态（0=未上架，1=已上架）
	PublishStatus uint8 `json:"publish_status"`
	// 连载状态（1=连载中，2=已完结，3=未知）
	SerialStatus uint8 `json:"serial_status"`
	// 总时长（分钟）
	Duration uint32 `json:"duration"`
	// 播放量
	ViewCount uint32 `json:"view_count"`
	// 标签列表
	Tags []dto.NamedItem `json:"tags,omitempty"`
	// 创建时间
	CreatedAt string `json:"created_at,omitempty"`
	// 更新时间
	UpdatedAt string `json:"updated_at,omitempty"`
}

// VideoSourceEpisode is one playable episode under a source for admin (includes URL).
type VideoSourceEpisode struct {
	// 剧集记录ID
	ID uint32 `json:"id"`
	// 集数
	Episode uint32 `json:"episode"`
	// 剧集标题
	Title string `json:"title"`
	// 播放地址
	URL string `json:"url"`
	// 清晰度
	Quality string `json:"quality"`
	// 播放格式
	Format string `json:"format"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
}

// VideoSourceGroup groups episodes by play source for admin (includes URL).
type VideoSourceGroup struct {
	// 播放源ID
	ID uint32 `json:"id"`
	// 播放源名称
	Name string `json:"name"`
	// 剧集列表（含播放地址）
	Episodes []VideoSourceEpisode `json:"episodes"`
}

// VideoDetailResponse is a full video detail payload for admin (includes publish_status & play URLs).
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
	// 上架状态（0=未上架，1=已上架）
	PublishStatus uint8 `json:"publish_status"`
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
	// 播放量
	ViewCount uint32 `json:"view_count"`
	// 导演列表
	Directors []dto.NamedItem `json:"directors"`
	// 演员列表
	Actors []dto.NamedItem `json:"actors"`
	// 标签列表
	Tags []dto.NamedItem `json:"tags"`
	// 播放源分组列表（含播放地址）
	Sources []VideoSourceGroup `json:"sources"`
}
