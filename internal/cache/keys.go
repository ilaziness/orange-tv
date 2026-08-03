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
	TTLSettings         = 5 * time.Minute

	// Open - 分类
	KeyOpenCategories = "open:categories"
	TTLOpenCategories = 5 * time.Minute

	// Open - 视频列表（动态键，含 data_range + source 维度）
	KeyTplOpenVideoList = "open:videos:list:%d:%d:%s:%s"
	TTLOpenVideoList    = 2 * time.Minute

	// Video (client) - 视频列表（动态键）
	KeyTplVideoList = "video:list:%d:%d:%s:%d:%d"
	TTLVideoList    = 2 * time.Minute
)

// SettingsGroupKey generates a per-group settings cache key.
func SettingsGroupKey(group string) string {
	return fmt.Sprintf(KeyTplSettingsGroup, group)
}

// VideoListKey 生成客户端视频列表缓存键。
func VideoListKey(categoryID, parentCategoryID uint64, sort string, page, limit int) string {
	return fmt.Sprintf(KeyTplVideoList, categoryID, parentCategoryID, sort, page, limit)
}

// OpenVideoListKey 生成 Open API 视频列表缓存键。
// dataRange 和 source 为筛选维度，空字符串表示无筛选。
func OpenVideoListKey(page, pageSize int, dataRange, source string) string {
	return fmt.Sprintf(KeyTplOpenVideoList, page, pageSize, dataRange, source)
}
