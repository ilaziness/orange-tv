-- Feature platform flags + cloud storage settings + media library.

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
