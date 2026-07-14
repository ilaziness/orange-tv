// Package model provides database models.
package model

import (
	"time"

	"github.com/uptrace/bun"
)

// User represents the users table.
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        int64     `bun:"id,autoincrement,pk" json:"id"`
	Email     string    `bun:"email,unique" json:"email" validate:"required,email"`
	Password  string    `bun:"password,notnull" json:"-" validate:"required"`
	Name      string    `bun:"name,notnull" json:"name" validate:"required"`
	Phone     string    `bun:"phone" json:"phone"`
	Status    int       `bun:"status,notnull,default:1" json:"status"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt time.Time `bun:"deleted_at,nullzero" json:"deleted_at,omitempty"`
}
