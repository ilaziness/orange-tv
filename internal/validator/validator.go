// Package validator provides request validation with multi-language support.
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	validate   *validator.Validate
	translator ut.Translator
)

func init() {
	validate = validator.New()

	// 注册中文翻译器
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)
	translator, _ = uni.GetTranslator("zh")

	// 注册中文翻译
	_ = zhTranslations.RegisterDefaultTranslations(validate, translator)

	// 注册自定义标签名函数
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			return field.Name
		}
		// 去除 omitempty 等选项
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}
		return jsonTag
	})
}

// Validate validates a struct and returns translated error messages.
func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := make([]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				messages = append(messages, e.Translate(translator))
			}
			return fmt.Errorf("%s", strings.Join(messages, "; "))
		}
		return err
	}
	return nil
}

// ValidateVar validates a single variable with given tag.
func ValidateVar(field any, tag string) error {
	if err := validate.Var(field, tag); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fe := validationErrors[0]
			return fmt.Errorf("%s (%s)", fe.Translate(translator), fe.Tag())
		}
		return err
	}
	return nil
}

// GetValidator returns the underlying validator instance.
func GetValidator() *validator.Validate {
	return validate
}

// GetTranslator returns the translator instance.
func GetTranslator() ut.Translator {
	return translator
}

// RegisterValidation registers a custom validation function.
func RegisterValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	return validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

// RegisterTranslation registers a custom translation for a validation tag.
func RegisterTranslation(tag string, message string) error {
	return validate.RegisterTranslation(tag, translator, func(ut ut.Translator) error {
		return ut.Add(tag, message, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return fe.Translate(ut)
	})
}
