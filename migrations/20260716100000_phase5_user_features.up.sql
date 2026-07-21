-- Phase 5: 用户体系 / 收藏 / 历史 / 评论 / Banner / 访问统计
-- 使用 --bun:split 拆分语句

-- 用户登录日志（与 admin 登录日志分离，便于权限隔离）
CREATE TABLE IF NOT EXISTS user_login_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    username VARCHAR(64) NOT NULL DEFAULT '',
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1成功 0失败',
    message VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_user_login_logs_user_id (user_id),
    KEY idx_user_login_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
--bun:split

-- 用户收藏
CREATE TABLE IF NOT EXISTS user_favorites (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    video_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_video (user_id, video_id),
    KEY idx_user_favorites_user_id (user_id),
    KEY idx_user_favorites_video_id (video_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
--bun:split

-- 用户播放历史
CREATE TABLE IF NOT EXISTS user_play_history (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    video_id BIGINT UNSIGNED NOT NULL,
    play_source_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    episode_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    progress INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '播放进度（秒）',
    duration INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总时长（秒）',
    last_played_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_video (user_id, video_id),
    KEY idx_user_play_history_user_id (user_id, last_played_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
--bun:split

-- 影视评论
CREATE TABLE IF NOT EXISTS video_comments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    video_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父评论ID，0为顶级',
    content VARCHAR(1000) NOT NULL,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1正常 0隐藏',
    like_count INT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_video_comments_video_id (video_id, status, created_at),
    KEY idx_video_comments_user_id (user_id),
    KEY idx_video_comments_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
--bun:split

-- 首页 Banner（轮播横幅）
CREATE TABLE IF NOT EXISTS banners (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title VARCHAR(128) NOT NULL DEFAULT '',
    cover VARCHAR(500) NOT NULL DEFAULT '',
    link VARCHAR(500) NOT NULL DEFAULT '',
    video_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    sort INT UNSIGNED NOT NULL DEFAULT 0,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_banners_status_sort (status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
--bun:split

-- 站点访问统计（按天聚合，简化版在线人数/访问量）
CREATE TABLE IF NOT EXISTS site_stats_daily (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    stat_date DATE NOT NULL,
    pv BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '页面浏览量',
    uv BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '独立访客（按IP近似）',
    online_peak INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '当日在线峰值（近似）',
    PRIMARY KEY (id),
    UNIQUE KEY uk_site_stats_date (stat_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
--bun:split

-- 在线会话心跳表（用于近似在线人数，定期清理）
CREATE TABLE IF NOT EXISTS online_sessions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    session_key VARCHAR(64) NOT NULL,
    ip VARCHAR(64) NOT NULL DEFAULT '',
    last_active_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_online_sessions_key (session_key),
    KEY idx_online_sessions_last_active (last_active_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
--bun:split

-- 默认超级管理员用户组（若未存在）
INSERT IGNORE INTO user_groups (id, name, permissions, description, created_at, updated_at)
VALUES (1, 'super_admin', '["*"]', '超级管理员', NOW(), NOW());
--bun:split

-- 默认普通用户组
INSERT IGNORE INTO user_groups (id, name, permissions, description, created_at, updated_at)
VALUES (2, 'member', '[]', '普通用户', NOW(), NOW());
