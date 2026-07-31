// Package constant contains database-model related enum constants.
package constant

// Status constants for enable/disable fields.
const (
	StatusDisabled uint8 = 0
	StatusEnabled  uint8 = 1
)

// Publish status for videos.
const (
	PublishStatusOffline uint8 = 0
	PublishStatusOnline  uint8 = 1
)

// Serial status for videos.
const (
	SerialStatusOngoing  uint8 = 1
	SerialStatusFinished uint8 = 2
	SerialStatusUpcoming uint8 = 3
)

// Collect source formats (collect_sources.type).
const (
	CollectTypeDefault  uint8 = 1 // system JSON format
	CollectTypeAppleCMS uint8 = 2 // 苹果 CMS
)

// Collect log status (collect_logs.status).
const (
	CollectLogCompleted uint8 = 1
	CollectLogRunning   uint8 = 2
	CollectLogFailed    uint8 = 3
)

// Login log status (shared by admin_login_logs and user_login_logs).
const (
	LoginStatusSuccess uint8 = 1
	LoginStatusFailed  uint8 = 2
)

// System log levels.
const (
	SystemLogLevelInfo     uint8 = 1
	SystemLogLevelWarning  uint8 = 2
	SystemLogLevelError    uint8 = 3
	SystemLogLevelCritical uint8 = 4
)

// Setting type values (system_settings.setting_type).
const (
	SettingTypeString  uint8 = 1
	SettingTypeNumber  uint8 = 2
	SettingTypeBoolean uint8 = 3
	SettingTypeJSON    uint8 = 4
)

// Comment status values (video_comments.status).
const (
	CommentStatusHidden uint8 = 0 // 隐藏/待审核
	CommentStatusNormal uint8 = 1 // 正常
)
