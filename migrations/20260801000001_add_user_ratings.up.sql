CREATE TABLE IF NOT EXISTS user_ratings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    video_id BIGINT UNSIGNED NOT NULL,
    score DECIMAL(3,1) UNSIGNED NOT NULL COMMENT '评分 0.5-10.0',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_video (user_id, video_id),
    KEY idx_user_ratings_video_id (video_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

ALTER TABLE videos ADD COLUMN rating_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户评分人数' AFTER rating;

--bun:split

-- Seed default feature settings
INSERT IGNORE INTO system_settings (setting_key, setting_group, setting_value, setting_type, description) VALUES
('rating_enabled', 'feature', '1', 3, '视频评分开关');
