-- Rebuild legacy login_logs table (with user_type, ip_address).
CREATE TABLE IF NOT EXISTS login_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '用户类型：1管理员 2普通用户',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    ip_address VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'User-Agent',
    status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '登录状态：1成功 2失败',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_type, user_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

-- Restore user_login_logs message column and original status default.
ALTER TABLE user_login_logs ADD COLUMN message VARCHAR(255) NOT NULL DEFAULT '' AFTER status;

--bun:split

ALTER TABLE user_login_logs MODIFY COLUMN status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1成功 0失败';

--bun:split

DROP TABLE IF EXISTS admin_login_logs;
