-- Feature platform flags + cloud storage settings + media library + user live streaming.

-- Convert feature toggles from global boolean to per-platform JSON flags.
UPDATE `system_settings`
SET `setting_value` = CASE
    WHEN LOWER(TRIM(`setting_value`)) IN ('1', 'true', 'yes', 'on') THEN '{"web":true,"desktop":true,"app":true,"tv":true}'
    WHEN LOWER(TRIM(`setting_value`)) IN ('0', 'false', 'no', 'off') THEN '{"web":false,"desktop":false,"app":false,"tv":false}'
    WHEN `setting_key` = 'livetv_enabled' THEN '{"web":false,"desktop":false,"app":false,"tv":false}'
    ELSE '{"web":true,"desktop":true,"app":true,"tv":true}'
END,
`setting_type` = 4
WHERE `setting_group` = 'feature'
  AND `setting_key` IN ('livetv_enabled', 'comment_enabled', 'comment_review', 'rating_enabled')
  AND `setting_type` = 3;

--bun:split

INSERT IGNORE INTO `system_settings` (`setting_key`, `setting_group`, `setting_value`, `setting_type`, `description`, `created_at`, `updated_at`) VALUES
('storage_provider', 'storage', 'none', 1, '当前启用的云存储厂商（none/aliyun/tencent/qiniu）', NOW(), NOW()),
('storage_aliyun', 'storage', '{}', 4, '阿里云 OSS 配置 JSON', NOW(), NOW()),
('storage_tencent', 'storage', '{}', 4, '腾讯云 COS 配置 JSON', NOW(), NOW()),
('storage_qiniu', 'storage', '{}', 4, '七牛云 Kodo 配置 JSON', NOW(), NOW());

--bun:split

ALTER TABLE `system_settings`
    MODIFY `setting_value` TEXT NOT NULL COMMENT '设置值';

--bun:split

INSERT IGNORE INTO `system_settings` (`setting_key`, `setting_group`, `setting_value`, `setting_type`, `description`, `created_at`, `updated_at`) VALUES
('payment_alipay', 'payment', '{}', 4, '支付宝配置 JSON', NOW(), NOW()),
('payment_wechat', 'payment', '{}', 4, '微信支付配置 JSON', NOW(), NOW());

--bun:split

CREATE TABLE IF NOT EXISTS `media_assets` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `media_type` VARCHAR(32) NOT NULL COMMENT '媒体类型：image（预留 video 等）',
    `owner_kind` VARCHAR(16) NOT NULL COMMENT '上传方：admin / user',
    `owner_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员或用户 ID',
    `provider` VARCHAR(32) NOT NULL COMMENT '云厂商：aliyun/tencent/qiniu',
    `object_key` VARCHAR(512) NOT NULL COMMENT '对象存储 key',
    `url` VARCHAR(1024) NOT NULL COMMENT '加速域名公网地址',
    `mime` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'MIME 类型',
    `size` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小（字节）',
    `original_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原始文件名',
    `width` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图片宽度（可选）',
    `height` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图片高度（可选）',
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_media_type_created` (`media_type`, `created_at`),
    KEY `idx_owner` (`owner_kind`, `owner_id`),
    KEY `idx_object_key` (`object_key`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='媒体库资源';

--bun:split

-- User live streaming (主播开播) schema + settings seeds.
-- Distinct from IPTV livetv_* tables.

ALTER TABLE `users`
    ADD COLUMN `is_streamer` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否主播：1是 0否' AFTER `status`,
    ADD COLUMN `streamer_at` DATETIME NULL DEFAULT NULL COMMENT '成为主播时间' AFTER `is_streamer`,
    ADD INDEX `idx_users_is_streamer` (`is_streamer`);

--bun:split

CREATE TABLE IF NOT EXISTS `streamer_applies` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人用户ID',
    `reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '申请理由',
    `contact` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '联系方式',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态：0待审 1通过 2拒绝',
    `review_admin_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '审核管理员ID',
    `review_remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '审核备注',
    `reviewed_at` DATETIME NULL DEFAULT NULL COMMENT '审核时间',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_streamer_applies_user_status` (`user_id`, `status`),
    KEY `idx_streamer_applies_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主播申请流水';

--bun:split

CREATE TABLE IF NOT EXISTS `live_categories` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '分类名称',
    `sort` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序，越小越前',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_live_categories_status_sort` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户直播分类';

--bun:split

INSERT IGNORE INTO `live_categories` (`id`, `name`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(1, '娱乐', 10, 1, NOW(), NOW()),
(2, '游戏', 20, 1, NOW(), NOW()),
(3, '聊天', 30, 1, NOW(), NOW()),
(4, '才艺', 40, 1, NOW(), NOW()),
(5, '其他', 100, 1, NOW(), NOW());

--bun:split

CREATE TABLE IF NOT EXISTS `live_rooms` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL COMMENT '主播用户ID',
    `category_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '直播分类ID',
    `title` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '直播间标题',
    `cover` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '封面URL',
    `notice` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '公告',
    `stream_key` VARCHAR(64) NOT NULL COMMENT '推流密钥（明文）',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态：0空闲 1直播中 2禁播',
    `current_session_id` INT UNSIGNED NULL DEFAULT NULL COMMENT '当前场次ID',
    `started_at` DATETIME NULL DEFAULT NULL COMMENT '本次开播时间',
    `ended_at` DATETIME NULL DEFAULT NULL COMMENT '最近停播时间',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_live_rooms_user_id` (`user_id`),
    UNIQUE KEY `uk_live_rooms_stream_key` (`stream_key`),
    KEY `idx_live_rooms_status_category` (`status`, `category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户直播间';

--bun:split

CREATE TABLE IF NOT EXISTS `live_sessions` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `room_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '直播间ID',
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '主播用户ID',
    `title` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '开播时标题快照',
    `category_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '开播时分类快照',
    `streamer_nickname` VARCHAR(15) NOT NULL DEFAULT '' COMMENT '开播时昵称快照',
    `streamer_avatar` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '开播时头像快照',
    `started_at` DATETIME NOT NULL COMMENT '开播时间',
    `ended_at` DATETIME NULL DEFAULT NULL COMMENT '结束时间',
    `end_reason` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '结束原因：host/admin/revoke/banned',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_live_sessions_room_started` (`room_id`, `started_at`),
    KEY `idx_live_sessions_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户直播场次';

--bun:split

INSERT IGNORE INTO `system_settings` (`setting_key`, `setting_group`, `setting_value`, `setting_type`, `description`, `created_at`, `updated_at`) VALUES
('live_push_server', 'live', '', 1, 'OBS 推流服务器地址', NOW(), NOW()),
('live_play_hls_template', 'live', '', 1, 'HLS 播放地址模板（含 {stream_key}）', NOW(), NOW()),
('live_play_flv_template', 'live', '', 1, 'FLV 播放地址模板（含 {stream_key}，可空）', NOW(), NOW());

--bun:split

INSERT IGNORE INTO `system_settings` (`setting_key`, `setting_group`, `setting_value`, `setting_type`, `description`, `created_at`, `updated_at`) VALUES
('live_enabled', 'feature', '{"web":false,"desktop":false,"app":false,"tv":false}', 4, '用户直播开关（按端）', NOW(), NOW());
