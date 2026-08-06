-- Advertisements table: replaces the old ad settings in system_settings.
-- Supports video_loading (pre-roll) and general ad scenes.
-- type=code stores third-party ad platform code (e.g. AdSense) in content_code.

CREATE TABLE IF NOT EXISTS `advertisements` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `ad_key` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '广告标识，前端用于区分广告位置，唯一',
    `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '广告标题',
    `scene` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '广告场景：video_loading 播放前广告 / general 一般广告',
    `type` VARCHAR(20) NOT NULL DEFAULT 'image' COMMENT '广告类型：image 图片 / video 视频 / html iframe页面 / code 广告平台代码',
    `content_url` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '广告素材URL（image/video/html类型使用）',
    `content_code` TEXT COMMENT '广告平台代码（code类型使用，如AdSense的script代码片段）',
    `link_url` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '点击跳转链接（code类型不适用）',
    `duration` INT UNSIGNED NOT NULL DEFAULT 5 COMMENT '单条展示时长（秒）',
    `sort` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序，越小越靠前',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    UNIQUE `uk_ad_key` (`ad_key`),
    INDEX `idx_scene_status` (`scene`, `status`),
    INDEX `idx_sort` (`sort`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='广告表';

--bun:split

-- Remove old ad settings from system_settings (replaced by advertisements table)
DELETE FROM `system_settings` WHERE `setting_group` = 'ad';
