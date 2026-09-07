-- Reverse feature platform flags + cloud storage + media library (reverse order).

DROP TABLE IF EXISTS `media_assets`;

--bun:split

DELETE FROM `system_settings` WHERE `setting_key` IN (
  'storage_provider',
  'storage_aliyun',
  'storage_tencent',
  'storage_qiniu'
);

--bun:split

-- Revert per-platform JSON feature flags back to global boolean strings.
UPDATE `system_settings`
SET `setting_value` = CASE
    WHEN JSON_VALID(`setting_value`) AND JSON_EXTRACT(`setting_value`, '$.web') IS NOT NULL THEN
        IF(JSON_EXTRACT(`setting_value`, '$.web') = TRUE
           OR JSON_UNQUOTE(JSON_EXTRACT(`setting_value`, '$.web')) IN ('1', 'true', 'TRUE'), '1', '0')
    WHEN `setting_key` = 'livetv_enabled' THEN '0'
    ELSE '1'
END,
`setting_type` = 3
WHERE `setting_group` = 'feature'
  AND `setting_key` IN ('livetv_enabled', 'comment_enabled', 'comment_review', 'rating_enabled');
