package captcha

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// newTestCaptcha 创建用于测试的验证码实例（内置内存 store）。
func newTestCaptcha(opts ...Option) Captcha {
	return New(opts...)
}

func TestGenerateAndVerify(t *testing.T) {
	c := newTestCaptcha()
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if img.ID == "" {
		t.Fatal("ID is empty")
	}
	if !strings.HasPrefix(img.Content, "data:image/png;base64,") {
		t.Fatalf("Content is not a png data URI: %q", img.Content[:min(40, len(img.Content))])
	}
	if img.ExpiresIn <= 0 {
		t.Fatalf("ExpiresIn should be positive, got %d", img.ExpiresIn)
	}
}

func TestVerifySuccess(t *testing.T) {
	c := newTestCaptcha()
	// base64Captcha DriverString 的 answer == content（生成的随机字符），
	// 测试中无法直接拿到明文答案，故通过 store 探测获取答案。
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	pc := c.(*pngCaptcha)
	answer := pc.captcha.Store.Get(img.ID, false)
	if answer == "" {
		t.Fatal("cannot retrieve answer from store")
	}
	if err := c.Verify(context.Background(), img.ID, answer); err != nil {
		t.Fatalf("Verify should succeed, got %v", err)
	}
}

func TestVerifyCaseInsensitiveAndTrim(t *testing.T) {
	c := newTestCaptcha()
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	pc := c.(*pngCaptcha)
	answer := pc.captcha.Store.Get(img.ID, false)
	if answer == "" {
		t.Fatal("cannot retrieve answer from store")
	}
	// 大小写不敏感 + 前后空格
	upper := " " + strings.ToUpper(answer) + " "
	if err := c.Verify(context.Background(), img.ID, upper); err != nil {
		t.Fatalf("Verify should be case-insensitive and trim, got %v", err)
	}
}

func TestVerifyNotMatched(t *testing.T) {
	c := newTestCaptcha()
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	err = c.Verify(context.Background(), img.ID, "wrongAnswer")
	if !errors.Is(err, ErrNotMatched) {
		t.Fatalf("expected ErrNotMatched, got %v", err)
	}
}

func TestVerifyNotFound(t *testing.T) {
	c := newTestCaptcha()
	err := c.Verify(context.Background(), "nonexistent-id", "any")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// 校验一次性：成功校验后再次校验应返回 ErrNotFound。
func TestVerifyOneTimeAfterSuccess(t *testing.T) {
	c := newTestCaptcha()
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	pc := c.(*pngCaptcha)
	answer := pc.captcha.Store.Get(img.ID, false)
	if err := c.Verify(context.Background(), img.ID, answer); err != nil {
		t.Fatalf("first Verify should succeed, got %v", err)
	}
	if err := c.Verify(context.Background(), img.ID, answer); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second Verify should be ErrNotFound (one-time), got %v", err)
	}
}

// 校验一次性：失败校验后再次校验也应作废。
func TestVerifyOneTimeAfterFailure(t *testing.T) {
	c := newTestCaptcha()
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if err := c.Verify(context.Background(), img.ID, "wrong"); !errors.Is(err, ErrNotMatched) {
		t.Fatalf("expected ErrNotMatched, got %v", err)
	}
	// 错误答案也会作废，再次用正确答案校验应 NotFound。
	pc := c.(*pngCaptcha)
	answer := pc.captcha.Store.Get(img.ID, false)
	if answer != "" {
		t.Fatal("answer should be cleared after failed verify")
	}
}

// nil store 降级为内置内存 store，仍可正常工作。
func TestNilStoreFallbackToMemory(t *testing.T) {
	c := New() // 不传 store
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate with nil store failed: %v", err)
	}
	pc := c.(*pngCaptcha)
	answer := pc.captcha.Store.Get(img.ID, false)
	if answer == "" {
		t.Fatal("fallback memory store should hold answer")
	}
	if err := c.Verify(context.Background(), img.ID, answer); err != nil {
		t.Fatalf("Verify with fallback store failed: %v", err)
	}
}
