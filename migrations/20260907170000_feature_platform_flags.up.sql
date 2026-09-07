-- Convert feature toggles from global boolean to per-platform JSON flags.
UPDATE `system_settings`
SET `setting_value` = CASE
    WHEN LOWER(TRIM(`setting_value`)) IN ('1', 'true', 'yes', 'on') THEN '{"web":true,"desktop":true,"app":true,"tv":true}'
    WHEN LOWER(TRIM(`setting_value`)) IN ('0', 'false', 'no', 'off') THEN '{"web":false,"desktop":false,"app":false,"tv":false}'
    WHEN `setting_key` = 'livetv_enabled' THEN '{"web":false,"desktop":false,"app":false,"tv":false}'
    ELSE '{"web":true,"desktop":true,"app":true,"tv":true}'
END,
`setting_type` = 4
WHERE `setting_group` = 'feature'
  AND `setting_key` IN ('livetv_enabled', 'comment_enabled', 'comment_review', 'rating_enabled');
