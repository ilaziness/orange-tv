package captcha

import (
	"github.com/golang/freetype/truetype"
	base64Captcha "github.com/mojocn/base64Captcha"
	"golang.org/x/image/font/gofont/goregular"
)

// goRegularFont 预解析的 Go Regular 无衬线字体（golang.org/x/image 内置）。
//
// base64Captcha 内嵌的 9 种字体（3Dumb、ApothecaryFont、Flim-Flam、
// RitaSmith 等）全部为装饰性/手写风格艺术字体，人眼难以辨认，导致
// 验证码需多次刷新才能出现可读实例。改用 Go 官方提供的无衬线字体，
// 字形规整清晰，显著提升人眼识别率。
var goRegularFont = func() *truetype.Font {
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		// goregular.TTF 是 Go 官方内嵌字体，解析失败只可能发生在编译期
		// 资源损坏的极端情况，此时无法继续提供验证码服务。
		panic("captcha: parse go regular font: " + err.Error())
	}
	return f
}()

// goFontStorage 基于 Go Regular 字体的 FontsStorage 实现，
// 替换 base64Captcha 默认的装饰性字体集合，供 DriverString 使用。
type goFontStorage struct{}

// LoadFontByName 返回 Go Regular 字体，忽略 name 参数。
func (goFontStorage) LoadFontByName(_ string) *truetype.Font {
	return goRegularFont
}

// LoadFontsByNames 返回仅含 Go Regular 的字体切片，忽略 names 参数。
func (goFontStorage) LoadFontsByNames(_ []string) []*truetype.Font {
	return []*truetype.Font{goRegularFont}
}

// newGoFontStorage 创建基于 Go Regular 字体的存储。
func newGoFontStorage() base64Captcha.FontsStorage {
	return goFontStorage{}
}
