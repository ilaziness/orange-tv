-- Phase 1 full schema (MySQL). Soft-delete via deleted_at WITHOUT dedicated index.

CREATE TABLE IF NOT EXISTS categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '分类名称',
    parent_id BIGINT NOT NULL DEFAULT 0 COMMENT '父分类ID',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS videos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL COMMENT '标题',
    subtitle VARCHAR(255) NOT NULL DEFAULT '' COMMENT '副标题',
    description TEXT COMMENT '描述',
    category_id BIGINT NOT NULL DEFAULT 0 COMMENT '分类ID',
    publish_status TINYINT NOT NULL DEFAULT 0 COMMENT '上下架状态：1上架 0下架',
    serial_status TINYINT NOT NULL DEFAULT 1 COMMENT '连载状态：1连载中 2已完结 3即将上线',
    cover_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '封面图',
    poster_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '海报图',
    year INT NOT NULL DEFAULT 0 COMMENT '年份',
    region VARCHAR(50) NOT NULL DEFAULT '' COMMENT '地区',
    rating DECIMAL(3,1) NOT NULL DEFAULT 0.0 COMMENT '评分',
    view_count INT NOT NULL DEFAULT 0 COMMENT '播放次数',
    duration INT NOT NULL DEFAULT 0 COMMENT '时长（分钟）',
    language VARCHAR(50) NOT NULL DEFAULT '' COMMENT '语言',
    release_date DATE NULL COMMENT '上映日期',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_category (category_id),
    INDEX idx_year (year),
    FULLTEXT idx_search (title, subtitle, description)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS directors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '导演名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS video_directors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_id BIGINT NOT NULL COMMENT '影视ID',
    director_id BIGINT NOT NULL COMMENT '导演ID',
    INDEX idx_video (video_id),
    INDEX idx_director (director_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS actors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '演员名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS video_actors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_id BIGINT NOT NULL COMMENT '影视ID',
    actor_id BIGINT NOT NULL COMMENT '演员ID',
    role VARCHAR(100) NOT NULL DEFAULT '' COMMENT '角色名',
    INDEX idx_video (video_id),
    INDEX idx_actor (actor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '标签名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS video_tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_id BIGINT NOT NULL COMMENT '影视ID',
    tag_id BIGINT NOT NULL COMMENT '标签ID',
    INDEX idx_video (video_id),
    INDEX idx_tag (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS play_sources (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '源名称（如"播放源1"、"采集源A"）',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS play_episodes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    source_id BIGINT NOT NULL COMMENT '播放源ID',
    video_id BIGINT NOT NULL COMMENT '影视ID',
    episode_number INT NOT NULL DEFAULT 1 COMMENT '集数编号',
    title VARCHAR(255) NOT NULL DEFAULT '' COMMENT '集标题（如"第1集"）',
    play_url VARCHAR(1000) NOT NULL COMMENT '播放地址',
    quality VARCHAR(50) NOT NULL DEFAULT '' COMMENT '清晰度',
    format VARCHAR(20) NOT NULL DEFAULT '' COMMENT '格式（hls/mp4/dash/flv）',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    UNIQUE uk_source_video_episode (source_id, video_id, episode_number),
    INDEX idx_video (video_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS collect_sources (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '采集源名称',
    type TINYINT NOT NULL DEFAULT 1 COMMENT '采集源格式：1默认(系统格式) 2苹果CMS格式',
    collect_url VARCHAR(500) NOT NULL COMMENT '采集地址',
    api_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'API密钥',
    config JSON COMMENT '采集配置',
    cron_expr VARCHAR(100) NOT NULL DEFAULT '' COMMENT '定时采集cron表达式，空表示未开启定时采集',
    play_source_id BIGINT NOT NULL DEFAULT 0 COMMENT '绑定播放源ID，采集到的播放链接存入该播放源',
    last_collect_at TIMESTAMP NULL COMMENT '最后采集时间',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS collect_source_categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    source_id BIGINT NOT NULL COMMENT '采集源ID',
    external_category VARCHAR(100) NOT NULL COMMENT '外部分类名称（采集源返回的分类）',
    category_id BIGINT NOT NULL COMMENT '系统内分类ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE uk_source_external (source_id, external_category),
    INDEX idx_category (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS collect_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    source_id BIGINT NOT NULL COMMENT '采集源ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '采集状态：1成功 2失败 3部分成功',
    total_count INT NOT NULL DEFAULT 0 COMMENT '采集总数',
    success_count INT NOT NULL DEFAULT 0 COMMENT '成功数',
    failed_count INT NOT NULL DEFAULT 0 COMMENT '失败数',
    error_message TEXT COMMENT '错误信息',
    duration_ms INT NOT NULL DEFAULT 0 COMMENT '耗时(毫秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_source (source_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS live_channels (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '频道名称',
    category VARCHAR(50) NOT NULL DEFAULT '' COMMENT '频道分类',
    stream_url VARCHAR(1000) NOT NULL COMMENT '直播流地址',
    logo VARCHAR(500) NOT NULL DEFAULT '' COMMENT '频道Logo',
    description TEXT COMMENT '频道描述',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS themes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '主题名称',
    identifier VARCHAR(100) NOT NULL UNIQUE COMMENT '主题标识',
    version VARCHAR(20) NOT NULL DEFAULT '1.0.0' COMMENT '版本',
    author VARCHAR(100) NOT NULL DEFAULT '' COMMENT '作者',
    description TEXT COMMENT '描述',
    preview_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '预览图',
    config JSON COMMENT '主题配置（管理员覆盖后的最终配置，合并自theme.json默认值）',
    custom_css TEXT COMMENT '自定义CSS',
    custom_js TEXT COMMENT '自定义JS',
    is_default TINYINT NOT NULL DEFAULT 0 COMMENT '是否默认',
    is_active TINYINT NOT NULL DEFAULT 0 COMMENT '是否当前启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS system_settings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    setting_key VARCHAR(100) NOT NULL UNIQUE COMMENT '设置键',
    setting_value TEXT COMMENT '设置值',
    setting_type TINYINT NOT NULL DEFAULT 1 COMMENT '设置类型：1string 2number 3boolean 4json',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

INSERT INTO system_settings (setting_key, setting_value, setting_type, description) VALUES
('site_mode', 'video_site', 1, '站点模式：video_site(影视站) resource_site(资源站)'),
('api_output_format', 'custom', 1, 'API输出格式：custom(自定义) apple_cms(苹果CMS)'),
('enable_third_party_collect', '1', 3, '是否允许第三方采集'),
('active_theme_id', '1', 2, '当前激活主题ID（与themes.is_active互为冗余，以themes表为准）');

--bun:split

CREATE TABLE IF NOT EXISTS user_groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '用户组名称',
    permissions JSON COMMENT '权限列表',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS admins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码（加密存储）',
    email VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱',
    avatar VARCHAR(500) NOT NULL DEFAULT '' COMMENT '头像',
    group_id BIGINT NOT NULL DEFAULT 0 COMMENT '用户组ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_group (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码（加密存储）',
    email VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱',
    avatar VARCHAR(500) NOT NULL DEFAULT '' COMMENT '头像',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS login_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_type TINYINT NOT NULL DEFAULT 1 COMMENT '用户类型：1管理员 2普通用户',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    ip_address VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'User-Agent',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '登录状态：1成功 2失败',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_type, user_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

CREATE TABLE IF NOT EXISTS system_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    level TINYINT NOT NULL DEFAULT 1 COMMENT '日志级别：1info 2warning 3error 4critical',
    module VARCHAR(50) NOT NULL DEFAULT '' COMMENT '模块',
    action VARCHAR(100) NOT NULL DEFAULT '' COMMENT '操作',
    admin_id BIGINT NOT NULL DEFAULT 0 COMMENT '操作管理员ID',
    content TEXT COMMENT '日志内容',
    ip_address VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_module_created (module, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
