package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Age      int    `json:"age" validate:"gte=0,lte=130"`
	Password string `json:"password" validate:"required,min=8"`
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
		errMsg  string
	}{
		{
			name: "验证成功",
			input: TestStruct{
				Email:    "test@example.com",
				Name:     "Test User",
				Age:      25,
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "缺少必填字段",
			input: TestStruct{
				Email:    "",
				Name:     "Test User",
				Age:      25,
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email",
		},
		{
			name: "邮箱格式错误",
			input: TestStruct{
				Email:    "invalid-email",
				Name:     "Test User",
				Age:      25,
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email",
		},
		{
			name: "名称太短",
			input: TestStruct{
				Email:    "test@example.com",
				Name:     "AB",
				Age:      25,
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "年龄超出范围",
			input: TestStruct{
				Email:    "test@example.com",
				Name:     "Test User",
				Age:      150,
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "age",
		},
		{
			name: "密码太短",
			input: TestStruct{
				Email:    "test@example.com",
				Name:     "Test User",
				Age:      25,
				Password: "123",
			},
			wantErr: true,
			errMsg:  "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateVar(t *testing.T) {
	tests := []struct {
		name    string
		field   any
		tag     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "邮箱验证成功",
			field:   "test@example.com",
			tag:     "required,email",
			wantErr: false,
		},
		{
			name:    "邮箱验证失败",
			field:   "invalid-email",
			tag:     "required,email",
			wantErr: true,
			errMsg:  "email",
		},
		{
			name:    "数字范围验证成功",
			field:   25,
			tag:     "gte=0,lte=130",
			wantErr: false,
		},
		{
			name:    "数字范围验证失败",
			field:   150,
			tag:     "gte=0,lte=130",
			wantErr: true,
		},
		{
			name:    "字符串长度验证成功",
			field:   "password123",
			tag:     "required,min=8",
			wantErr: false,
		},
		{
			name:    "字符串长度验证失败",
			field:   "123",
			tag:     "required,min=8",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.field, tt.tag)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetValidator(t *testing.T) {
	v := GetValidator()
	assert.NotNil(t, v)
}

func TestGetTranslator(t *testing.T) {
	tr := GetTranslator()
	assert.NotNil(t, tr)
}

func TestRegisterValidation(t *testing.T) {
	// 创建新的验证器实例避免污染全局状态
	v := validator.New()
	err := v.RegisterValidation("custom", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "custom"
	})
	assert.NoError(t, err)

	// 测试自定义验证
	type CustomStruct struct {
		Field string `validate:"custom"`
	}

	// 验证成功
	err = v.Struct(CustomStruct{Field: "custom"})
	assert.NoError(t, err)

	// 验证失败
	err = v.Struct(CustomStruct{Field: "other"})
	assert.Error(t, err)
}

func TestRegisterTranslation(t *testing.T) {
	// 创建新的验证器实例避免污染全局状态
	v := validator.New()
	_ = v.RegisterValidation("custom2", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "custom"
	})

	// 测试自定义验证
	type CustomStruct struct {
		Field string `validate:"custom2"`
	}

	err := v.Struct(CustomStruct{Field: "other"})
	assert.Error(t, err)
}
