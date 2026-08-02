CREATE TABLE IF NOT EXISTS user_comment_votes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    comment_id BIGINT UNSIGNED NOT NULL,
    direction TINYINT NOT NULL COMMENT '1顶 -1踩',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_comment (user_id, comment_id),
    KEY idx_comment_direction (comment_id, direction)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

