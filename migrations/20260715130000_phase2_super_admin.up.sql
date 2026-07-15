-- Phase 2: seed super_admin user group (no default admin credentials).

INSERT INTO user_groups (name, permissions, description)
SELECT 'super_admin', '["*"]', '超级管理员，拥有全部后台权限'
WHERE NOT EXISTS (
    SELECT 1 FROM user_groups WHERE name = 'super_admin' AND deleted_at IS NULL
);
