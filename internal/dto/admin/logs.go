package admin

import shareddto "github.com/ilaziness/orange-tv/internal/dto"

// SystemLogListRequest queries system_logs.
type SystemLogListRequest struct {
	shareddto.PaginationRequest
	Level   *uint8  `form:"level" validate:"omitempty,oneof=1 2 3 4"`
	Module  string  `form:"module"`
	AdminID *uint32 `form:"admin_id" validate:"omitempty,min=1"`
	// Start / End are RFC3339 or date strings; optional.
	Start string `form:"start"`
	End   string `form:"end"`
}

// AdminLoginLogListRequest queries admin_login_logs.
type AdminLoginLogListRequest struct {
	shareddto.PaginationRequest
	Username string `form:"username"`
	Status   *uint8 `form:"status" validate:"omitempty,oneof=1 2"`
	Start    string `form:"start"`
	End      string `form:"end"`
}

// UserLoginLogListRequest queries user_login_logs.
type UserLoginLogListRequest struct {
	shareddto.PaginationRequest
	UserID   *uint32 `form:"user_id" validate:"omitempty,min=1"`
	Username string  `form:"username"`
	Status   *uint8  `form:"status" validate:"omitempty,oneof=1 2"`
	Start    string  `form:"start"`
	End      string  `form:"end"`
}

// SystemLogItem is one system log row.
type SystemLogItem struct {
	ID        uint32 `json:"id"`
	Level     uint8  `json:"level"`
	Module    string `json:"module"`
	Action    string `json:"action"`
	AdminID   uint32 `json:"admin_id"`
	Content   string `json:"content"`
	IPAddress string `json:"ip_address"`
	CreatedAt string `json:"created_at"`
}

// AdminLoginLogItem is one admin login log row.
type AdminLoginLogItem struct {
	ID        uint32 `json:"id"`
	UserID    uint32 `json:"user_id"`
	Username  string `json:"username"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Status    uint8  `json:"status"`
	CreatedAt string `json:"created_at"`
}

// UserLoginLogItem is one user login log row.
type UserLoginLogItem struct {
	ID        uint32 `json:"id"`
	UserID    uint32 `json:"user_id"`
	Username  string `json:"username"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Status    uint8  `json:"status"`
	CreatedAt string `json:"created_at"`
}

// AppLogListRequest reads the zap log file from the end backwards.
type AppLogListRequest struct {
	// Offset is the byte offset from the end of file; 0 means start from the last byte.
	Offset int64 `form:"offset"`
	// Limit is the max number of log lines to return (default 50, max 200).
	Limit int `form:"limit" validate:"omitempty,min=1,max=200"`
}

// GetLimit returns a sanitized limit value.
func (r *AppLogListRequest) GetLimit() int {
	if r == nil || r.Limit <= 0 {
		return 50
	}
	if r.Limit > 200 {
		return 200
	}
	return r.Limit
}

// AppLogItem is one parsed zap log line.
type AppLogItem struct {
	Time   string         `json:"time"`
	Level  string         `json:"level"`
	Msg    string         `json:"msg"`
	Fields map[string]any `json:"fields,omitempty"`
}

// AppLogListResponse is the paginated response for app log reading.
type AppLogListResponse struct {
	List       []AppLogItem `json:"list"`
	HasMore    bool         `json:"has_more"`
	NextOffset int64        `json:"next_offset"`
}
