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
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ActorItem is actor info with role.
type ActorItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// CategoryResponse is a category tree node.
type CategoryResponse struct {
	ID        int64              `json:"id"`
	Name      string             `json:"name"`
	ParentID  int64              `json:"parent_id"`
	SortOrder int32              `json:"sort_order"`
	Status    int8               `json:"status"`
	Children  []CategoryResponse `json:"children"`
}

// VideoListItem is a compact video card payload.
type VideoListItem struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Subtitle      string  `json:"subtitle"`
	Cover         string  `json:"cover"`
	Poster        string  `json:"poster"`
	Year          int32   `json:"year"`
	Region        string  `json:"region"`
	Language      string  `json:"language"`
	Rating        float64 `json:"rating"`
	CategoryID    int64   `json:"category_id"`
	PublishStatus int8    `json:"publish_status,omitempty"`
	SerialStatus  int8    `json:"serial_status"`
	Duration      int32   `json:"duration"`
	ViewCount     int32   `json:"view_count"`
}

// VideoSourceEpisode is one playable episode under a source.
type VideoSourceEpisode struct {
	Episode int32  `json:"episode"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Format  string `json:"format"`
}

// VideoSourceGroup groups episodes by play source.
type VideoSourceGroup struct {
	ID       int64                `json:"id"`
	Name     string               `json:"name"`
	Episodes []VideoSourceEpisode `json:"episodes"`
}

// LiveChannelItem is a live channel payload for admin and client.
type LiveChannelItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	StreamURL   string `json:"stream_url"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	SortOrder   int32  `json:"sort_order"`
	Status      int8   `json:"status,omitempty"`
}

// ThemeCurrentResponse is the merged active theme for the client.
type ThemeCurrentResponse struct {
	Name       string         `json:"name"`
	Identifier string         `json:"identifier"`
	Version    string         `json:"version"`
	Config     map[string]any `json:"config"`
	Templates  map[string]any `json:"templates,omitempty"`
	CustomCSS  string         `json:"custom_css"`
	CustomJS   string         `json:"custom_js"`
}

// ThemeListItem is an admin theme list entry.
type ThemeListItem struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	Identifier   string         `json:"identifier"`
	Version      string         `json:"version"`
	Author       string         `json:"author"`
	Description  string         `json:"description"`
	PreviewImage string         `json:"preview_image"`
	Config       map[string]any `json:"config"`
	CustomCSS    string         `json:"custom_css"`
	CustomJS     string         `json:"custom_js"`
	IsDefault    int8           `json:"is_default"`
	IsActive     int8           `json:"is_active"`
}

// CollectSourceItem is an admin collect source payload.
// API key is never returned in list/detail responses.
type CollectSourceItem struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Type          int8   `json:"type"`
	CollectURL    string `json:"collect_url"`
	Config        string `json:"config,omitempty"`
	CronExpr      string `json:"cron_expr"`
	PlaySourceID  int64  `json:"play_source_id"`
	LastCollectAt string `json:"last_collect_at,omitempty"`
	Status        int8   `json:"status"`
}

// CollectCategoryMapItem is an external→internal category mapping.
type CollectCategoryMapItem struct {
	ID               int64  `json:"id"`
	SourceID         int64  `json:"source_id"`
	ExternalCategory string `json:"external_category"`
	CategoryID       int64  `json:"category_id"`
}

// CollectLogItem is one collect run log entry.
type CollectLogItem struct {
	ID           int64  `json:"id"`
	SourceID     int64  `json:"source_id"`
	Status       int8   `json:"status"`
	TotalCount   int32  `json:"total_count"`
	SuccessCount int32  `json:"success_count"`
	FailedCount  int32  `json:"failed_count"`
	ErrorMessage string `json:"error_message,omitempty"`
	DurationMs   int32  `json:"duration_ms"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// VideoDetailResponse is a full video detail payload.
type VideoDetailResponse struct {
	ID            int64              `json:"id"`
	Title         string             `json:"title"`
	Subtitle      string             `json:"subtitle"`
	Description   string             `json:"description"`
	CategoryID    int64              `json:"category_id"`
	PublishStatus int8               `json:"publish_status,omitempty"`
	SerialStatus  int8               `json:"serial_status"`
	Cover         string             `json:"cover"`
	Poster        string             `json:"poster"`
	Year          int32              `json:"year"`
	Region        string             `json:"region"`
	Language      string             `json:"language"`
	Duration      int32              `json:"duration"`
	ReleaseDate   string             `json:"release_date,omitempty"`
	Rating        float64            `json:"rating"`
	ViewCount     int32              `json:"view_count"`
	Directors     []NamedItem        `json:"directors"`
	Actors        []ActorItem        `json:"actors"`
	Tags          []NamedItem        `json:"tags"`
	Sources       []VideoSourceGroup `json:"sources"`
}
