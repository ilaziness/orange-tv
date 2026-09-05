package captcha

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
)

// 默认参数按 DriverString 的绘制规则选取：
// 字号约为 height*(7~13)/16，末字从 width/length 处起笔。
// 高度过大时字会画出右边界；故用较扁画布，保证 4 字完整落入图内。
const (
	defaultLength = 4   // 验证码字符数
	defaultWidth  = 120 // 图片宽度（像素）
	defaultHeight = 32  // 图片高度（像素）；字号随高度增大，过高会裁掉末字
	defaultNoise  = 6   // 噪点字符；库内噪点为浅色，过多会糊成一片
	// defaultExpireTTL 验证码默认有效期，同时用于内置内存 store 的过期回收
	// 与外部 cache store 的 TTL。调用方可通过 WithExpireTTL / WithCacheStore 覆盖。
	defaultExpireTTL = 5 * time.Minute
	// defaultCollectNum 内存 store 触发过期回收的条目阈值。
	defaultCollectNum = 1024
)

// defaultBgColor 中灰底。DriverString 字符色走 RandDeepColor，可能偏浅；
// 中灰底让浅色字仍能看出轮廓，深色字对比也足够。
var defaultBgColor = color.RGBA{R: 200, G: 200, B: 200, A: 255}

// defaultShowLineOptions 只用 HollowLine（浅色空心曲线）。
// SlimeLine / SineLine 与正文一样走 RandDeepColor，线、字会糊在一起。
const defaultShowLineOptions = base64Captcha.OptionShowHollowLine

// defaultFonts 库内嵌 chromohv，字形相对规整、宽度可控。
// Fonts 留空会混入 3Dumb 等过宽装饰体，4 字更容易画出画布。
var defaultFonts = []string{"chromohv.ttf"}

// Options 默认实现的配置项。
type Options struct {
	// Store 验证码存储。为 nil 时使用 base64Captcha 内置内存 store，
	// 保证应用未配置外部 Cache 时验证码仍可用。
	Store base64Captcha.Store

	// Length 验证码字符数，默认 4。
	Length int

	// Width 图片宽度（像素），默认 120。
	Width int

	// Height 图片高度（像素），默认 32。
	Height int

	// NoiseCount 噪点字符数，默认 6。
	NoiseCount int

	// ExpireTTL 验证码有效期，默认 5 分钟。
	// 同时用于内置内存 store 的过期回收与外部 store 的 TTL。
	ExpireTTL time.Duration
}

// Option 配置函数。
type Option func(*Options)

// WithStore 设置外部存储（如基于 pkg/cache 的 Redis/内存共享存储）。
func WithStore(s base64Captcha.Store) Option {
	return func(o *Options) { o.Store = s }
}

// WithLength 设置验证码字符数。
func WithLength(n int) Option {
	return func(o *Options) { o.Length = n }
}

// WithSize 设置图片尺寸（像素）。
func WithSize(width, height int) Option {
	return func(o *Options) { o.Width, o.Height = width, height }
}

// WithNoiseCount 设置噪点字符数。
func WithNoiseCount(n int) Option {
	return func(o *Options) { o.NoiseCount = n }
}

// WithExpireTTL 设置验证码有效期。
func WithExpireTTL(ttl time.Duration) Option {
	return func(o *Options) { o.ExpireTTL = ttl }
}

func applyOptions(opts []Option) *Options {
	o := &Options{
		Length:     defaultLength,
		Width:      defaultWidth,
		Height:     defaultHeight,
		NoiseCount: defaultNoise,
		ExpireTTL:  defaultExpireTTL,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.Length <= 0 {
		o.Length = defaultLength
	}
	if o.Width <= 0 {
		o.Width = defaultWidth
	}
	if o.Height <= 0 {
		o.Height = defaultHeight
	}
	if o.NoiseCount < 0 {
		o.NoiseCount = defaultNoise
	}
	if o.ExpireTTL <= 0 {
		o.ExpireTTL = defaultExpireTTL
	}
	return o
}

// pngCaptcha 基于 base64Captcha 的默认图像验证码实现。
//
// 底层渲染委托给 base64Captcha.DriverString，本结构仅做接口适配与错误归一化。
type pngCaptcha struct {
	captcha   *base64Captcha.Captcha
	expireTTL time.Duration
}

// New 创建默认图像验证码实现。
//
// opts.Store 为 nil 时使用 base64Captcha 内置内存 store（NewMemoryStore），
// 保证应用未配置外部 Cache 时验证码仍可用；传入外部 store（如基于
// pkg/cache 的共享存储）则用于多实例部署下的验证码共享。
func New(opts ...Option) Captcha {
	o := applyOptions(opts)
	store := o.Store
	if store == nil {
		// 内置内存 store：collectNum 触发过期回收阈值，expiration 为有效期。
		store = base64Captcha.NewMemoryStore(defaultCollectNum, o.ExpireTTL)
	}
	// DriverString 内置字段：扁平画布 + chromohv + 浅色空心线 + 中灰底。
	driver := (&base64Captcha.DriverString{
		Height:          o.Height,
		Width:           o.Width,
		NoiseCount:      o.NoiseCount,
		ShowLineOptions: defaultShowLineOptions,
		Length:          o.Length,
		Source:          base64Captcha.TxtSimpleCharaters,
		BgColor:         &defaultBgColor,
		Fonts:           defaultFonts,
	}).ConvertFonts()
	return &pngCaptcha{
		captcha:   base64Captcha.NewCaptcha(driver, store),
		expireTTL: o.ExpireTTL,
	}
}

// Generate 生成并存储一个验证码，返回展示图片与标识。
//
// 注意：scene 参数当前不参与存储 key 的构造。base64Captcha 使用全局
// 唯一的随机 id（RandomId）作为存储键，本身已保证不同场景、不同请求的
// 验证码互不冲突。保留 scene 参数是为了接口约定的稳定性，便于未来
// 扩展实现（如按场景差异化配置字符集、长度或限流）时无需改动调用方。
func (p *pngCaptcha) Generate(_ context.Context, _ string) (*Image, error) {
	id, b64s, _, err := p.captcha.Generate()
	if err != nil {
		return nil, fmt.Errorf("captcha: generate: %w", err)
	}
	return &Image{
		ID:        id,
		Content:   b64s,
		ExpiresIn: int(p.expireTTL.Seconds()),
	}, nil
}

// Verify 校验验证码答案，校验后作废（一次性）。
//
// 为区分「不存在/已作废」与「答案不匹配」两种失败，这里不直接使用
// base64Captcha.Verify（它在 clear=true 时无条件删除，失败后无法探测），
// 而是自行控制 clear 语义：
//  1. 先 Get(id, false) 探测是否存在（不删除）；
//  2. 不存在 → ErrNotFound；
//  3. 答案不匹配 → ErrNotMatched（此时再删除，作废本次验证码，防穷举）；
//  4. 匹配 → 删除并返回 nil。
//
// 答案比对与 base64Captcha 保持一致：TrimSpace + 大小写不敏感。
func (p *pngCaptcha) Verify(_ context.Context, id, answer string) error {
	stored := p.captcha.Store.Get(id, false)
	if stored == "" {
		return ErrNotFound
	}
	if !stringsEqualFoldTrim(stored, answer) {
		// 答案错误也要作废，避免同一验证码被反复穷举。
		_ = p.captcha.Store.Get(id, true)
		return ErrNotMatched
	}
	// 匹配，删除（一次性）。
	_ = p.captcha.Store.Get(id, true)
	return nil
}

// stringsEqualFoldTrim 复刻 base64Captcha.Verify 的答案比对逻辑：
// 先 TrimSpace 再 EqualFold（大小写不敏感）。
func stringsEqualFoldTrim(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
