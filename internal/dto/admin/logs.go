package admin

import shareddto "github.com/ilaziness/orange-tv/internal/dto"

// SystemLogListRequest queries system_logs.
type SystemLogListRequest struct {
	shareddto.PaginationRequest
	// 日志级别筛选（1=Debug，2=Info，3=Warn，4=Error）
	Level *uint8 `form:"level" binding:"omitempty,oneof=1 2 3 4"`
	// 模块筛选
	Module string `form:"module"`
	// 管理员ID筛选
	AdminID *uint32 `form:"admin_id" binding:"omitempty,min=1"`
	// 开始时间（RFC3339 或日期字符串）
	Start string `form:"start"`
	// 结束时间（RFC3339 或日期字符串）
	End string `form:"end"`
}

// AdminLoginLogListRequest queries admin_login_logs.
type AdminLoginLogListRequest struct {
	shareddto.PaginationRequest
	// 管理员用户名筛选
	Username string `form:"username"`
	// 登录结果筛选（1=成功，2=失败）
	Status *uint8 `form:"status" binding:"omitempty,oneof=1 2"`
	// 开始时间（RFC3339 或日期字符串）
	Start string `form:"start"`
	// 结束时间（RFC3339 或日期字符串）
	End string `form:"end"`
}

// UserLoginLogListRequest queries user_login_logs.
type UserLoginLogListRequest struct {
	shareddto.PaginationRequest
	// 用户ID筛选
	UserID *uint32 `form:"user_id" binding:"omitempty,min=1"`
	// 登录邮箱筛选
	Email string `form:"email"`
	// 登录结果筛选（1=成功，2=失败）
	Status *uint8 `form:"status" binding:"omitempty,oneof=1 2"`
	// 开始时间（RFC3339 或日期字符串）
	Start string `form:"start"`
	// 结束时间（RFC3339 或日期字符串）
	End string `form:"end"`
}

// SystemLogItem is one system log row.
type SystemLogItem struct {
	// 日志ID
	ID uint32 `json:"id"`
	// 日志级别（1=Debug，2=Info，3=Warn，4=Error）
	Level uint8 `json:"level"`
	// 模块
	Module string `json:"module"`
	// 操作动作
	Action string `json:"action"`
	// 操作管理员ID
	AdminID uint32 `json:"admin_id"`
	// 日志内容
	Content string `json:"content"`
	// 操作 IP 地址
	IPAddress string `json:"ip_address"`
	// 记录时间
	CreatedAt string `json:"created_at"`
}

// AdminLoginLogItem is one admin login log row.
type AdminLoginLogItem struct {
	// 日志ID
	ID uint32 `json:"id"`
	// 管理员用户ID
	UserID uint32 `json:"user_id"`
	// 管理员用户名
	Username string `json:"username"`
	// 登录 IP 地址
	IP string `json:"ip"`
	// User-Agent 信息
	UserAgent string `json:"user_agent"`
	// 登录结果（1=成功，2=失败）
	Status uint8 `json:"status"`
	// 登录时间
	CreatedAt string `json:"created_at"`
}

// UserLoginLogItem is one user login log row.
type UserLoginLogItem struct {
	// 日志ID
	ID uint32 `json:"id"`
	// 用户ID
	UserID uint32 `json:"user_id"`
	// 登录邮箱
	Email string `json:"email"`
	// 登录 IP 地址
	IP string `json:"ip"`
	// User-Agent 信息
	UserAgent string `json:"user_agent"`
	// 登录结果（1=成功，2=失败）
	Status uint8 `json:"status"`
	// 登录时间
	CreatedAt string `json:"created_at"`
}

// AppLogListRequest reads the zap log file from the end backwards.
type AppLogListRequest struct {
	// 距文件末尾的字节偏移量，0 表示从末尾开始
	Offset int64 `form:"offset"`
	// 返回的最大日志行数（默认 50，最大 200）
	Limit int `form:"limit" binding:"omitempty,min=1,max=200"`
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
	// 日志时间
	Time string `json:"time"`
	// 日志级别（如 info、error）
	Level string `json:"level"`
	// 日志消息
	Msg string `json:"msg"`
	// 附加字段
	Fields map[string]any `json:"fields,omitempty"`
}

// AppLogListResponse is the paginated response for app log reading.
type AppLogListResponse struct {
	// 日志列表
	List []AppLogItem `json:"list"`
	// 是否还有更多
	HasMore bool `json:"has_more"`
	// 下一次读取的偏移量
	NextOffset int64 `json:"next_offset"`
}
