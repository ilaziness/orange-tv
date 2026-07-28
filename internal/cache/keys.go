package cache

import (
	"fmt"
	"time"
)

// 缓存键和 TTL（成对定义）
const (
	// Category
	KeyCategoryTreeClient = "category:tree:client"
	TTLCategoryTree       = 5 * time.Minute

	// Live
	KeyLiveListClient = "live:list:client"
	TTLLiveList       = 5 * time.Minute

	// Settings (per-group cache)
	KeyTplSettingsGroup = "settings:group:%s"
	KeySettingsPublic   = "settings:public"
	TTLSettings         = 5 * time.Minute

	// Open - 分类
	KeyOpenCategories = "open:categories"
	TTLOpenCategories = 5 * time.Minute

	// Open - 视频列表（动态键）
	KeyTplOpenVideoList = "open:videos:list:%s:%d:%d"
	TTLOpenVideoList    = 2 * time.Minute

	// Open - 视频详情（动态键）
	KeyTplOpenVideoDetail = "open:videos:detail:%s:%d"
	TTLOpenVideoDetail    = 2 * time.Minute

	// Video (client) - 视频列表（动态键）
	KeyTplVideoList = "video:list:%d:%s:%d:%d"
	TTLVideoList    = 2 * time.Minute
)

// SettingsGroupKey generates a per-group settings cache key.
func SettingsGroupKey(group string) string {
	return fmt.Sprintf(KeyTplSettingsGroup, group)
}

// VideoListKey 生成客户端视频列表缓存键。
func VideoListKey(categoryID uint64, sort string, page, limit int) string {
	return fmt.Sprintf(KeyTplVideoList, categoryID, sort, page, limit)
}

// OpenVideoListKey 生成 Open API 视频列表缓存键。
func OpenVideoListKey(format string, page, pageSize int) string {
	return fmt.Sprintf(KeyTplOpenVideoList, format, page, pageSize)
}

// OpenVideoDetailKey 生成 Open API 视频详情缓存键。
func OpenVideoDetailKey(format string, id int64) string {
	return fmt.Sprintf(KeyTplOpenVideoDetail, format, id)
}
