package captcha

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
)

// 默认参数。
const (
	defaultLength = 4   // 验证码字符数
	defaultWidth  = 120 // 图片宽度（像素）
	defaultHeight = 40  // 图片高度（像素）
	defaultNoise  = 15  // 噪点字符数（增强抗 OCR，仍可读）
	// defaultExpireTTL 验证码默认有效期，同时用于内置内存 store 的过期回收
	// 与外部 cache store 的 TTL。调用方可通过 WithExpireTTL / WithCacheStore 覆盖。
	defaultExpireTTL = 5 * time.Minute
	// defaultCollectNum 内存 store 触发过期回收的条目阈值。
	defaultCollectNum = 1024
)

// defaultBgColor 默认背景色：浅灰。
//
// base64Captcha.DriverString 的字符颜色调用 RandDeepColor，该函数对
// 随机色减去增量后取 math.Abs，可能产生浅色字符。纯白背景下浅色字符
// 几乎不可见，需多次刷新；改用浅灰背景可让浅色字符仍保持一定对比度，
// 同时深色字符仍清晰可辨，减少刷新次数。
var defaultBgColor = color.RGBA{R: 230, G: 230, B: 230, A: 255}

// defaultShowLineOptions 默认干扰线组合：HollowLine + SlimeLine + SineLine
// 三种干扰线全部开启，最大化抗 OCR 能力。
const defaultShowLineOptions = base64Captcha.OptionShowHollowLine |
	base64Captcha.OptionShowSlimeLine |
	base64Captcha.OptionShowSineLine

// Options 默认实现的配置项。
type Options struct {
	// Store 验证码存储。为 nil 时使用 base64Captcha 内置内存 store，
	// 保证应用未配置外部 Cache 时验证码仍可用。
	Store base64Captcha.Store

	// Length 验证码字符数，默认 4。
	Length int

	// Width 图片宽度（像素），默认 120。
	Width int

	// Height 图片高度（像素），默认 40。
	Height int

	// NoiseCount 噪点字符数，默认 15。
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
// 底层渲染委托给成熟的 github.com/mojocn/base64Captcha（内嵌字体、干扰线、
// 噪点、随机旋转等抗识别效果），本结构仅做接口适配与错误归一化。
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
	// 字符驱动：去除易混淆字符的字符集 + 三种干扰线 + 多噪点 + 浅灰底
	// + Go Regular 无衬线字体。
	//
	// 可读性优化：
	//   - 字体：使用 Go Regular 无衬线字体（newGoFontStorage），替换
	//     base64Captcha 默认的装饰性艺术字体集合（3Dumb/Flim-Flam/
	//     RitaSmith 等），字形规整清晰。
	//   - 背景：浅灰（defaultBgColor），相比纯白底让 RandDeepColor 偶尔
	//     产生的浅色字符仍保持可辨对比度，减少刷新次数。
	// 防御性优化：
	//   - 干扰线：HollowLine + SlimeLine + SineLine 全开（defaultShowLineOptions）。
	//   - 噪点：默认 15 个，平衡抗 OCR 与可读性。
	driver := base64Captcha.NewDriverString(
		o.Height, o.Width, o.NoiseCount,
		defaultShowLineOptions,
		o.Length,
		base64Captcha.TxtSimpleCharaters,
		&defaultBgColor,
		newGoFontStorage(), nil,
	).ConvertFonts()
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
