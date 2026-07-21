package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Users represents the users table.
type Users struct {
	bun.BaseModel `bun:"table:users,alias:us"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 用户名
	Username string `bun:"username,notnull,unique" json:"username"`
	// 密码（加密存储）
	Password string `bun:"password,notnull" json:"-"`
	// 邮箱
	Email string `bun:"email,notnull" json:"email"`
	// 头像
	Avatar string `bun:"avatar,notnull" json:"avatar"`
	// 状态：1启用 0禁用
	Status uint8 `bun:"status,notnull" json:"status"`
	// 最后登录时间
	LastLoginAt *time.Time `bun:"last_login_at" json:"last_login_at"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
	UserFavorites []*UserFavorites `bun:"rel:has-many,join:id=user_id" json:"-"`
	UserLoginLogs []*UserLoginLogs `bun:"rel:has-many,join:id=user_id" json:"-"`
	UserPlayHistories []*UserPlayHistory `bun:"rel:has-many,join:id=user_id" json:"-"`
	VideoComments []*VideoComments `bun:"rel:has-many,join:id=user_id" json:"-"`
}
