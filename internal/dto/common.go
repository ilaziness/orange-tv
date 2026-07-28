// Package dto provides shared data transfer objects used by multiple API surfaces.
package dto

// PaginationRequest 分页请求参数（兼容 page_size / limit）。
type PaginationRequest struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1,max=1000000"`
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
	Limit    int `form:"limit" json:"limit" binding:"omitempty,min=1,max=100"`
}

// GetPage 获取页码，如果未设置则返回默认值1
func (p *PaginationRequest) GetPage() int {
	if p.Page < 1 {
		return 1
	}
	return p.Page
}

// GetLimit 获取每页数量，优先 page_size，其次 limit，默认 20。
func (p *PaginationRequest) GetLimit() int {
	size := p.PageSize
	if size < 1 {
		size = p.Limit
	}
	if size < 1 || size > 100 {
		return 20
	}
	return size
}

// GetPageSize is an alias for GetLimit used by handlers expecting PRD field names.
func (p *PaginationRequest) GetPageSize() int {
	return p.GetLimit()
}

// GetOffset 获取偏移量
func (p *PaginationRequest) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

// GetTotalPages 计算总页数
func (p *PaginationRequest) GetTotalPages(total int) int {
	if total == 0 || p.GetLimit() == 0 {
		return 0
	}
	pages := total / p.GetLimit()
	if total%p.GetLimit() > 0 {
		pages++
	}
	return pages
}

// IDURI is a common URI path parameter.
type IDURI struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

// NamedItem is a simple id/name pair shared by content APIs.
type NamedItem struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// CategoryResponse is a category tree node.
type CategoryResponse struct {
	ID        uint64             `json:"id"`
	Name      string             `json:"name"`
	ParentID  uint64             `json:"parent_id"`
	SortOrder uint32             `json:"sort_order"`
	Status    uint8              `json:"status"`
	Children  []CategoryResponse `json:"children"`
}

// VideoListItem is a compact video card payload.
type VideoListItem struct {
	ID            uint64  `json:"id"`
	Title         string  `json:"title"`
	Subtitle      string  `json:"subtitle"`
	Cover         string  `json:"cover"`
	Poster        string  `json:"poster"`
	Year          uint32  `json:"year"`
	Region        string  `json:"region"`
	Language      string  `json:"language"`
	Rating        float64 `json:"rating"`
	CategoryID    uint64  `json:"category_id"`
	PublishStatus uint8   `json:"publish_status,omitempty"`
	SerialStatus  uint8   `json:"serial_status"`
	Duration      uint32  `json:"duration"`
	ViewCount     uint32  `json:"view_count"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

// VideoSourceEpisode is one playable episode under a source (admin/open API, includes URL).
type VideoSourceEpisode struct {
	Episode uint32 `json:"episode"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Format  string `json:"format"`
}

// VideoSourceGroup groups episodes by play source (admin/open API, includes URL).
type VideoSourceGroup struct {
	ID       uint64               `json:"id"`
	Name     string               `json:"name"`
	Episodes []VideoSourceEpisode `json:"episodes"`
}

// VideoDetailEpisode is an episode summary without play URL (client detail API).
type VideoDetailEpisode struct {
	Episode uint32 `json:"episode"`
	Title   string `json:"title"`
}

// VideoDetailSourceGroup groups episode summaries by play source (client detail API, no URL).
type VideoDetailSourceGroup struct {
	ID       uint64               `json:"id"`
	Name     string               `json:"name"`
	Episodes []VideoDetailEpisode `json:"episodes"`
}

// PlayEpisodeResponse is the single-episode play URL response.
type PlayEpisodeResponse struct {
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Format  string `json:"format"`
}

// LiveChannelItem is a live channel payload for admin and client.
type LiveChannelItem struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	StreamURL   string `json:"stream_url"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	SortOrder   uint32 `json:"sort_order"`
	Status      uint8  `json:"status,omitempty"`
}

// LiveSyncResult is the result of a live source synchronization.
type LiveSyncResult struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
	Deleted int `json:"deleted"`
}

// CollectSourceItem is an admin collect source payload.
// API key is never returned in list/detail responses.
type CollectSourceItem struct {
	ID              uint64 `json:"id"`
	Name            string `json:"name"`
	Type            uint8  `json:"type"`
	CollectURL      string `json:"collect_url"`
	CronExpr        string `json:"cron_expr"`
	PlaySourceID    uint64 `json:"play_source_id"`
	PlaySourceName  string `json:"play_source_name,omitempty"`
	LastCollectAt   string `json:"last_collect_at,omitempty"`
	Status          uint8  `json:"status"`
	ScheduleEnabled uint8  `json:"schedule_enabled"`
	DataRange       string `json:"data_range,omitempty"`
}

// CollectCategoryMapItem is an external→internal category mapping.
type CollectCategoryMapItem struct {
	ID                 uint64 `json:"id"`
	SourceID           uint64 `json:"source_id"`
	ExternalCategoryID uint64 `json:"external_category_id"`
	CategoryID         uint64 `json:"category_id"`
}

// CollectLogItem is one collect run log entry.
type CollectLogItem struct {
	ID           uint64 `json:"id"`
	SourceID     uint64 `json:"source_id"`
	SourceName   string `json:"source_name,omitempty"`
	Status       uint8  `json:"status"`
	CollectCount uint32 `json:"collect_count"`
	DurationSec  uint32 `json:"duration_sec"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// VideoDetailResponse is a full video detail payload.
type VideoDetailResponse struct {
	ID            uint64             `json:"id"`
	Title         string             `json:"title"`
	Subtitle      string             `json:"subtitle"`
	Description   string             `json:"description"`
	CategoryID    uint64             `json:"category_id"`
	PublishStatus uint8              `json:"publish_status,omitempty"`
	SerialStatus  uint8              `json:"serial_status"`
	Cover         string             `json:"cover"`
	Poster        string             `json:"poster"`
	Year          uint32             `json:"year"`
	Region        string             `json:"region"`
	Language      string             `json:"language"`
	Duration      uint32             `json:"duration"`
	ReleaseDate   string             `json:"release_date,omitempty"`
	Rating        float64            `json:"rating"`
	ViewCount     uint32             `json:"view_count"`
	Directors     []NamedItem        `json:"directors"`
	Actors        []NamedItem        `json:"actors"`
	Tags          []NamedItem        `json:"tags"`
	Sources       []VideoSourceGroup `json:"sources"`
}
