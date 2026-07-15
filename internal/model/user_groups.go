package model

import (
	"time"

	"github.com/uptrace/bun"
)

// UserGroups represents the user_groups table.
type UserGroups struct {
	bun.BaseModel `bun:"table:user_groups,alias:u"`

	ID int64 `bun:"id" json:"id"`
	// 用户组名称
	Name string `bun:"name" json:"name"`
	// 权限列表
	Permissions *string `bun:"permissions" json:"permissions"`
	// 描述
	Description string     `bun:"description" json:"description"`
	CreatedAt   *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}
