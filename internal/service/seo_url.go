package service

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizePublicBaseURL trims trailing slashes and validates http(s) absolute URLs.
// Empty input returns ("", nil). Invalid input returns an error suitable for ParamError.
func NormalizePublicBaseURL(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	u, err := url.Parse(v)
	if err != nil {
		return "", fmt.Errorf("站点根地址格式无效")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("站点根地址仅支持 http 或 https")
	}
	if u.Host == "" {
		return "", fmt.Errorf("站点根地址缺少主机名")
	}
	if u.User != nil {
		return "", fmt.Errorf("站点根地址不允许包含用户名或密码")
	}
	u.Fragment = ""
	u.RawQuery = ""
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "/" {
		u.Path = ""
	}
	// Client/nginx SEO routes are site-root; path prefixes would generate broken sitemap/canonical URLs.
	if u.Path != "" {
		return "", fmt.Errorf("站点根地址不能包含路径，请填写到域名（含端口）为止")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

// NormalizeOptionalHTTPURL validates an optional absolute http(s) URL (e.g. OG image).
// Empty input returns ("", nil).
func NormalizeOptionalHTTPURL(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	u, err := url.Parse(v)
	if err != nil {
		return "", fmt.Errorf("默认分享图地址格式无效")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("默认分享图仅支持 http 或 https 地址")
	}
	if u.Host == "" {
		return "", fmt.Errorf("默认分享图地址缺少主机名")
	}
	if u.User != nil {
		return "", fmt.Errorf("默认分享图地址不允许包含用户名或密码")
	}
	return u.String(), nil
}

// CollapseWhitespace replaces all unicode whitespace runs with a single space and trims.
func CollapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
