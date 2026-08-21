-- Initial schema (MySQL). Merged from all pre-release migrations.
-- Soft-delete via deleted_at WITHOUT dedicated index.

CREATE TABLE IF NOT EXISTS `categories` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '分类名称',
    `parent_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父分类ID',
    `sort_order` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX `idx_parent` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分类表';

--bun:split

CREATE TABLE IF NOT EXISTS `videos` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `title` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '标题',
    `subtitle` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '副标题',
    `description` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '描述',
    `category_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分类ID',
    `parent_category_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父分类ID，方便大分类筛选，0表示无父分类',
    `publish_status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上下架状态：1上架 0下架',
    `serial_status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '连载状态：1连载中 2已完结 3即将上线',
    `cover_image` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '封面图',
    `poster_image` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '海报图',
    `year` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '年份',
    `region` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '地区',
    `rating` DECIMAL(3,1) UNSIGNED NOT NULL DEFAULT 0.0 COMMENT '评分',
    `rating_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户评分人数',
    `view_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '播放次数',
    `duration` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '时长（分钟）',
    `language` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '语言',
    `collect_source_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采集源ID，标识最初由哪个采集源采集',
    `release_date` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '上映日期',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX `idx_category` (`category_id`),
    INDEX `idx_parent_category` (`parent_category_id`),
    INDEX `idx_year_desc` (`year` DESC, `id` DESC),
    INDEX `idx_collect_source` (`collect_source_id`),
    FULLTEXT `idx_search` (`title`, `subtitle`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='影视表';

--bun:split

CREATE TABLE IF NOT EXISTS `directors` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '导演名称',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    UNIQUE `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='导演表';

--bun:split

CREATE TABLE IF NOT EXISTS `video_directors` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '影视ID',
    `director_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '导演ID',
    INDEX `idx_video` (`video_id`),
    INDEX `idx_director` (`director_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='影视导演关联表';

--bun:split

CREATE TABLE IF NOT EXISTS `actors` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '演员名称',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    UNIQUE `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='演员表';

--bun:split

CREATE TABLE IF NOT EXISTS `video_actors` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '影视ID',
    `actor_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '演员ID',
    INDEX `idx_video` (`video_id`),
    INDEX `idx_actor` (`actor_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='影视演员关联表';

--bun:split

CREATE TABLE IF NOT EXISTS `tags` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '标签名称',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    UNIQUE `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签表';

--bun:split

CREATE TABLE IF NOT EXISTS `video_tags` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '影视ID',
    `tag_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '标签ID',
    INDEX `idx_video` (`video_id`),
    INDEX `idx_tag` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='影视标签关联表';

--bun:split

CREATE TABLE IF NOT EXISTS `play_sources` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '源名称（如"播放源1"、"采集源A"）',
    `sort_order` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='播放源表';

--bun:split

CREATE TABLE IF NOT EXISTS `play_episodes` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `source_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '播放源ID',
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '影视ID',
    `episode_number` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '集数编号',
    `title` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '集标题（如"第1集"）',
    `play_url` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '播放地址',
    `quality` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '清晰度',
    `format` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '格式（hls/mp4/dash/flv）',
    `sort_order` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    UNIQUE `uk_source_video_episode` (`source_id`, `video_id`, `episode_number`),
    INDEX `idx_video` (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='播放剧集表';

--bun:split

CREATE TABLE IF NOT EXISTS `collect_sources` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '采集源名称',
    `type` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '采集源格式：1默认(系统格式) 2苹果CMS格式',
    `collect_url` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '采集地址',
    `api_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'API密钥',
    `cron_expr` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '定时采集cron表达式，空表示未开启定时采集',
    `schedule_enabled` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否开启定时采集：1是 0否',
    `data_range` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '定时采集数据范围：today/last1d/last3d/last1w/last1m/all',
    `play_source_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定播放源ID，采集到的播放链接存入该播放源',
    `last_collect_at` DATETIME NULL COMMENT '最后采集时间',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采集源表';

--bun:split

CREATE TABLE IF NOT EXISTS `collect_source_categories` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `source_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采集源ID',
    `external_category_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '外部分类ID（采集源返回的 type_id）',
    `category_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '系统内分类ID',
    `created_at` DATETIME NOT NULL,
    UNIQUE `uk_source_external` (`source_id`, `external_category_id`),
    INDEX `idx_category` (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采集源分类映射表';

--bun:split

CREATE TABLE IF NOT EXISTS `collect_logs` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `source_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采集源ID',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '采集状态：1完成 2采集中',
    `collect_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采集数量（累加）',
    `duration_sec` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '耗时(秒)',
    `created_at` DATETIME NOT NULL,
    INDEX `idx_source` (`source_id`),
    INDEX `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采集日志表';

--bun:split

CREATE TABLE IF NOT EXISTS `live_channels` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '频道名称',
    `category` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '频道分类',
    `stream_url` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '直播流地址',
    `logo` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '频道Logo',
    `description` VARCHAR(125) NOT NULL DEFAULT '' COMMENT '频道描述',
    `sort_order` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播频道表';

--bun:split

CREATE TABLE IF NOT EXISTS `system_settings` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `setting_key` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '设置键',
    `setting_group` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '设置分组',
    `setting_value` VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '设置值',
    `setting_type` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '设置类型：1string 2number 3boolean 4json',
    `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    UNIQUE `uk_setting_key` (`setting_key`),
    INDEX `idx_setting_group` (`setting_group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统设置表';

--bun:split

INSERT IGNORE INTO `system_settings` (`setting_key`, `setting_group`, `setting_value`, `setting_type`, `description`, `created_at`, `updated_at`) VALUES
('enable_third_party_collect', 'api', '1', 3, '是否允许第三方采集', NOW(), NOW()),
('site_name', 'site', '小橘TV', 1, '站点名称', NOW(), NOW()),
('site_logo', 'site', '', 1, '站点 Logo URL', NOW(), NOW()),
('site_copyright', 'site', '', 1, '站点版权信息', NOW(), NOW()),
('site_icp', 'site', '', 1, '备案号', NOW(), NOW()),
('site_seo_keywords', 'site', '', 1, 'SEO 关键词', NOW(), NOW()),
('site_description', 'site', '', 1, '站点描述', NOW(), NOW()),
('live_enabled', 'feature', '0', 3, '电视直播开关', NOW(), NOW()),
('comment_enabled', 'feature', '1', 3, '视频评论开关', NOW(), NOW()),
('comment_review', 'feature', '1', 3, '评论是否需要审核', NOW(), NOW()),
('rating_enabled', 'feature', '1', 3, '视频评分开关', NOW(), NOW());

--bun:split

CREATE TABLE IF NOT EXISTS `user_groups` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '用户组名称',
    `permissions` JSON COMMENT '权限列表',
    `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    UNIQUE `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户组表';

--bun:split

INSERT IGNORE INTO `user_groups` (`id`, `name`, `permissions`, `description`, `created_at`, `updated_at`)
VALUES (1, 'super_admin', '["*"]', '超级管理员', NOW(), NOW());

--bun:split

INSERT IGNORE INTO `user_groups` (`id`, `name`, `permissions`, `description`, `created_at`, `updated_at`)
VALUES (2, 'member', '[]', '普通用户', NOW(), NOW());

--bun:split

CREATE TABLE IF NOT EXISTS `admins` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `username` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '用户名',
    `nickname` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '昵称',
    `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码（加密存储）',
    `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `avatar` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '头像',
    `group_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户组ID',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `last_login_at` DATETIME NULL COMMENT '最后登录时间',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX `idx_group` (`group_id`),
    UNIQUE `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员表';

--bun:split

CREATE TABLE IF NOT EXISTS `users` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码（加密存储）',
    `str_id` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '10位数字唯一展示ID',
    `nickname` VARCHAR(15) NOT NULL DEFAULT '' COMMENT '昵称',
    `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `avatar` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '头像',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    `last_login_at` DATETIME NULL COMMENT '最后登录时间',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
    UNIQUE INDEX `idx_users_str_id` (`str_id`),
    UNIQUE INDEX `uk_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

--bun:split

CREATE TABLE IF NOT EXISTS `admin_login_logs` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '管理员ID',
    `username` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '用户名',
    `ip` VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    `user_agent` VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'User-Agent',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '登录状态：1成功 2失败',
    `created_at` DATETIME NOT NULL,
    INDEX `idx_admin_login_logs_user` (`user_id`),
    INDEX `idx_admin_login_logs_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员登录日志表';

--bun:split

CREATE TABLE IF NOT EXISTS `system_logs` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `level` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '日志级别：1info 2warning 3error 4critical',
    `module` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '模块',
    `action` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '操作',
    `admin_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作管理员ID',
    `content` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '日志内容',
    `ip_address` VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    `created_at` DATETIME NOT NULL,
    INDEX `idx_module_created` (`module`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统日志表';

--bun:split

CREATE TABLE IF NOT EXISTS `user_login_logs` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '登录邮箱',
    `ip` VARCHAR(45) NOT NULL DEFAULT '',
    `user_agent` VARCHAR(500) NOT NULL DEFAULT '',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '1成功 2失败',
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_user_login_logs_user_id` (`user_id`),
    KEY `idx_user_login_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户登录日志表';

--bun:split

CREATE TABLE IF NOT EXISTS `user_favorites` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_video` (`user_id`, `video_id`),
    KEY `idx_user_favorites_user_id` (`user_id`),
    KEY `idx_user_favorites_video_id` (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户收藏表';

--bun:split

CREATE TABLE IF NOT EXISTS `user_play_history` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `play_source_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `episode_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `progress` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '播放进度（秒）',
    `duration` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总时长（秒）',
    `last_played_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_video` (`user_id`, `video_id`),
    KEY `idx_user_play_history_user_id` (`user_id`, `last_played_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户播放历史表';

--bun:split

CREATE TABLE IF NOT EXISTS `video_comments` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `parent_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父评论ID，0为顶级',
    `content` VARCHAR(1000) NOT NULL DEFAULT '',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0隐藏 1正常',
    `like_count` INT UNSIGNED NOT NULL DEFAULT 0,
    `dislike_count` INT UNSIGNED NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_video_comments_video_id` (`video_id`, `status`, `created_at`),
    KEY `idx_video_comments_user_id` (`user_id`),
    KEY `idx_video_comments_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频评论表';

--bun:split

CREATE TABLE IF NOT EXISTS `user_comment_votes` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `comment_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `direction` TINYINT NOT NULL COMMENT '1顶 -1踩',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_comment` (`user_id`, `comment_id`),
    KEY `idx_comment_direction` (`comment_id`, `direction`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户评论投票表';

--bun:split

CREATE TABLE IF NOT EXISTS `user_ratings` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `score` DECIMAL(3,1) UNSIGNED NOT NULL COMMENT '评分 0.5-10.0',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_video` (`user_id`, `video_id`),
    KEY `idx_user_ratings_video_id` (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户评分表';

--bun:split

CREATE TABLE IF NOT EXISTS `banners` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `title` VARCHAR(128) NOT NULL DEFAULT '',
    `cover` VARCHAR(500) NOT NULL DEFAULT '',
    `link` VARCHAR(500) NOT NULL DEFAULT '',
    `video_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `sort` INT UNSIGNED NOT NULL DEFAULT 0,
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_banners_status_sort` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='横幅表';

--bun:split

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

CREATE TABLE IF NOT EXISTS `site_stats_daily` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `stat_date` DATE NOT NULL,
    `pv` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '页面浏览量',
    `uv` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '独立访客（按IP近似）',
    `online_peak` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '当日在线峰值（近似）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_site_stats_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='站点每日统计表';

--bun:split

CREATE TABLE IF NOT EXISTS `online_sessions` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `session_key` VARCHAR(64) NOT NULL DEFAULT '',
    `ip` VARCHAR(45) NOT NULL DEFAULT '',
    `last_active_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_online_sessions_key` (`session_key`),
    KEY `idx_online_sessions_last_active` (`last_active_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='在线会话表';
