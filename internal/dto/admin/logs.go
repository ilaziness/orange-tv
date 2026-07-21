package admin

import shareddto "github.com/ilaziness/orange-tv/internal/dto"

// SystemLogListRequest queries system_logs.
type SystemLogListRequest struct {
	shareddto.PaginationRequest
	Level   *uint8  `form:"level" validate:"omitempty,oneof=1 2 3 4"`
	Module  string  `form:"module"`
	AdminID *uint64 `form:"admin_id" validate:"omitempty,min=1"`
	// Start / End are RFC3339 or date strings; optional.
	Start string `form:"start"`
	End   string `form:"end"`
}

// LoginLogListRequest queries login_logs.
type LoginLogListRequest struct {
	shareddto.PaginationRequest
	UserType *uint8 `form:"user_type" validate:"omitempty,oneof=1 2"`
	Username string `form:"username"`
	Status   *uint8 `form:"status" validate:"omitempty,oneof=1 2"`
	Start    string `form:"start"`
	End      string `form:"end"`
}

// SystemLogItem is one system log row.
type SystemLogItem struct {
	ID        uint64 `json:"id"`
	Level     uint8  `json:"level"`
	Module    string `json:"module"`
	Action    string `json:"action"`
	AdminID   uint64 `json:"admin_id"`
	Content   string `json:"content"`
	IPAddress string `json:"ip_address"`
	CreatedAt string `json:"created_at"`
}

// LoginLogItem is one login log row.
type LoginLogItem struct {
	ID        uint64 `json:"id"`
	UserType  uint8  `json:"user_type"`
	UserID    uint64 `json:"user_id"`
	Username  string `json:"username"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Status    uint8  `json:"status"`
	CreatedAt string `json:"created_at"`
}
