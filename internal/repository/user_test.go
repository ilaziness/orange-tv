package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "github.com/uptrace/bun/driver/sqliteshim"
)

func setupTestDB(t *testing.T) *database.DB {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	db := bun.NewDB(sqldb, sqlitedialect.New())

	// 创建测试表
	ctx := context.Background()
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT,
			phone TEXT,
			status INTEGER DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return &database.DB{DB: db}
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *model.User
		wantErr bool
	}{
		{
			name: "创建用户成功",
			user: &model.User{
				Email:    "test@example.com",
				Password: "hashedpassword",
				Name:     "Test User",
				Phone:    "1234567890",
				Status:   1,
			},
			wantErr: false,
		},
		{
			name: "重复邮箱失败",
			user: &model.User{
				Email:    "test@example.com", // 与上一条重复
				Password: "hashedpassword2",
				Name:     "Test User 2",
				Status:   1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.user.ID)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	// 先创建一个用户
	user := &model.User{
		Email:    "getbyid@example.com",
		Password: "hashedpassword",
		Name:     "Get By ID User",
		Status:   1,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "存在的用户",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "不存在的用户",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, user.Email, got.Email)
				assert.Equal(t, user.Name, got.Name)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	// 先创建一个用户
	user := &model.User{
		Email:    "getbyemail@example.com",
		Password: "hashedpassword",
		Name:     "Get By Email User",
		Status:   1,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "存在的邮箱",
			email:   "getbyemail@example.com",
			wantErr: false,
		},
		{
			name:    "不存在的邮箱",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByEmail(ctx, tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, user.Email, got.Email)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	// 先创建一个用户
	user := &model.User{
		Email:    "update@example.com",
		Password: "hashedpassword",
		Name:     "Original Name",
		Status:   1,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// 更新用户
	user.Name = "Updated Name"
	user.Phone = "9876543210"

	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// 验证更新成功
	updated, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "9876543210", updated.Phone)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	// 先创建一个用户
	user := &model.User{
		Email:    "delete@example.com",
		Password: "hashedpassword",
		Name:     "Delete User",
		Status:   1,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// 删除用户（软删除）
	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// 验证用户无法通过正常查询获取
	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)

	// 再次删除应返回错误
	err = repo.Delete(ctx, user.ID)
	assert.Error(t, err)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	// 创建多个用户
	for i := 0; i < 5; i++ {
		user := &model.User{
			Email:    "list" + string(rune('a'+i)) + "@example.com",
			Password: "hashedpassword",
			Name:     "List User " + string(rune('A'+i)),
			Status:   1,
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		limit    int
		offset   int
		expected int
	}{
		{
			name:     "限制2条",
			limit:    2,
			offset:   0,
			expected: 2,
		},
		{
			name:     "偏移2条",
			limit:    10,
			offset:   2,
			expected: 3,
		},
		{
			name:     "空结果",
			limit:    10,
			offset:   100,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := repo.List(ctx, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, users, tt.expected)
		})
	}
}

func TestUserRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	// 创建用户
	for i := 0; i < 3; i++ {
		user := &model.User{
			Email:    "count" + string(rune('a'+i)) + "@example.com",
			Password: "hashedpassword",
			Name:     "Count User",
			Status:   1,
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

// 并发创建用户测试
func TestUserRepository_ConcurrentCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	// 并发创建用户
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer func() { done <- true }()

			user := &model.User{
				Email:    "concurrent" + string(rune('0'+index)) + "@example.com",
				Password: "hashedpassword",
				Name:     "Concurrent User",
				Status:   1,
			}
			err := repo.Create(ctx, user)
			assert.NoError(t, err)
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证创建的用户数量
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, concurrency, count)
}
