package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "正常密码",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "空密码",
			password: "",
			wantErr:  false,
		},
		{
			name:     "长密码",
			password: "this_is_a_very_long_password_with_special_chars_!@#$%^&*()_+",
			wantErr:  false,
		},
		{
			name:     "中文密码",
			password: "中文密码测试",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			if err == nil && hash == tt.password {
				t.Error("HashPassword() returned plaintext password")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "正确密码",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "错误密码",
			password: "wrongpassword",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "空密码",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashPasswordWithCost(t *testing.T) {
	tests := []struct {
		name     string
		password string
		cost     int
		wantErr  bool
	}{
		{
			name:     "默认 cost",
			password: "password123",
			cost:     0,
			wantErr:  false,
		},
		{
			name:     "cost 4",
			password: "password123",
			cost:     4,
			wantErr:  false,
		},
		{
			name:     "cost 10",
			password: "password123",
			cost:     10,
			wantErr:  false,
		},
		{
			name:     "cost 12",
			password: "password123",
			cost:     12,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPasswordWithCost(tt.password, tt.cost)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPasswordWithCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 验证生成的哈希可以正确验证
				if err := CheckPassword(tt.password, hash); err != nil {
					t.Errorf("Generated hash failed verification: %v", err)
				}
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmark_password"
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "benchmark_password"
	hash, _ := HashPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckPassword(password, hash)
	}
}
