// Package lock 集中管理业务分布式锁的 key 模板与生成函数。
// 调用方应通过本包的辅助函数生成 key，避免散落在各业务模块的字面量。
package lock

import "fmt"

// 业务锁 key 模板常量。
const (
	// KeyTplUserRegister 用户注册按邮箱维度的分布式锁 key 模板。
	// 用于注册接口并发场景下做邮箱去重，与 DB 唯一索引互为兜底。
	KeyTplUserRegister = "user:register:email:%s"

	// KeyTplLiveRoom 用户直播间写路径锁（开停播/强制下播/禁播/重置密钥等）。
	KeyTplLiveRoom = "live:room:%d"

	// KeyTplLiveApplyUser 主播申请写路径锁（提交申请/审核），按用户维度。
	KeyTplLiveApplyUser = "live:apply:user:%d"
)

// UserRegisterKey 生成注册接口的邮箱锁 key。
// 入参 email 应为已 trim 的小写邮箱，避免大小写差异导致 key 不一致。
func UserRegisterKey(email string) string {
	return fmt.Sprintf(KeyTplUserRegister, email)
}

// LiveRoomKey 生成直播间写操作锁 key。
func LiveRoomKey(roomID uint32) string {
	return fmt.Sprintf(KeyTplLiveRoom, roomID)
}

// LiveApplyUserKey 生成主播申请按用户维度的锁 key。
func LiveApplyUserKey(userID uint32) string {
	return fmt.Sprintf(KeyTplLiveApplyUser, userID)
}
