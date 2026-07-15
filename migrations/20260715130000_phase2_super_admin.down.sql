-- Phase 2 rollback: remove seeded super_admin group only when unused.

DELETE ug FROM user_groups ug
LEFT JOIN admins a ON a.group_id = ug.id AND a.deleted_at IS NULL
WHERE ug.name = 'super_admin'
  AND ug.deleted_at IS NULL
  AND a.id IS NULL;
