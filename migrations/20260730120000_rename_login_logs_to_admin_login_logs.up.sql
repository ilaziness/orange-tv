-- Drop legacy mixed login_logs table (admin + user records with user_type).
-- Historical data is discarded per decision (clean rebuild).

--bun:split

DROP TABLE IF EXISTS login_logs;

--bun:split

-- Admin-only login logs (no user_type column).
CREATE TABLE IF NOT EXISTS admin_login_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
    username VARCHAR(50) NOT NULL DEFAULT '' COMMENT '用户名',
    ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'User-Agent',
    status TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '登录状态：1成功 2失败',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_admin_login_logs_user (user_id),
    INDEX idx_admin_login_logs_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--bun:split

-- Unify user_login_logs schema with admin_login_logs:
--   1. drop message column (no longer recorded)
--   2. unify status semantics to 1=success 2=failed
ALTER TABLE user_login_logs DROP COLUMN message;

--bun:split

ALTER TABLE user_login_logs MODIFY COLUMN status TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '1成功 2失败';
