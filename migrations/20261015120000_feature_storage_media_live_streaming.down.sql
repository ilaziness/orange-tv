-- Reverse user live streaming + feature platform flags + cloud storage + payment + media library (reverse order).

ALTER TABLE `user_login_logs` ADD INDEX `idx_user_login_logs_created_at` (`created_at`);

--bun:split

ALTER TABLE `advertisements`
    ADD INDEX `idx_scene_status` (`scene`, `status`),
    ADD INDEX `idx_sort` (`sort`, `id`);

--bun:split

ALTER TABLE `banners` ADD INDEX `idx_banners_status_sort` (`status`, `sort`);

--bun:split

ALTER TABLE `user_comment_votes` ADD INDEX `idx_comment_direction` (`comment_id`, `direction`);

--bun:split

ALTER TABLE `user_play_history` ADD INDEX `idx_user_play_history_user_id` (`user_id`, `last_played_at`);

--bun:split

ALTER TABLE `user_favorites`
    ADD INDEX `idx_user_favorites_user_id` (`user_id`),
    ADD INDEX `idx_user_favorites_video_id` (`video_id`);

--bun:split

ALTER TABLE `system_logs` ADD INDEX `idx_module_created` (`module`, `created_at`);

--bun:split

ALTER TABLE `admin_login_logs`
    ADD INDEX `idx_admin_login_logs_user` (`user_id`),
    ADD INDEX `idx_admin_login_logs_created` (`created_at`);

--bun:split

ALTER TABLE `admins` ADD INDEX `idx_group` (`group_id`);

--bun:split

ALTER TABLE `system_settings` ADD INDEX `idx_setting_group` (`setting_group`);

--bun:split

ALTER TABLE `collect_logs` ADD INDEX `idx_created` (`created_at`);

--bun:split

ALTER TABLE `collect_source_categories` ADD INDEX `idx_category` (`category_id`);

--bun:split

ALTER TABLE `videos`
    ADD INDEX `idx_collect_source` (`collect_source_id`),
    ADD FULLTEXT INDEX `idx_search` (`title`, `subtitle`),
    ADD INDEX `idx_year_desc` (`year` DESC, `id` DESC);

--bun:split

ALTER TABLE `categories` ADD INDEX `idx_parent` (`parent_id`);

--bun:split

DELETE FROM `system_settings` WHERE `setting_key` IN (
    'live_push_server',
    'live_play_hls_template',
    'live_play_flv_template',
    'live_enabled'
);

--bun:split

DROP TABLE IF EXISTS `live_sessions`;

--bun:split

DROP TABLE IF EXISTS `live_rooms`;

--bun:split

DROP TABLE IF EXISTS `live_categories`;

--bun:split

DROP TABLE IF EXISTS `streamer_applies`;

--bun:split

ALTER TABLE `users`
    DROP COLUMN `streamer_at`,
    DROP COLUMN `is_streamer`;

--bun:split

DROP TABLE IF EXISTS `media_assets`;

--bun:split

DELETE FROM `system_settings` WHERE `setting_key` IN (
  'payment_alipay',
  'payment_wechat',
  'storage_provider',
  'storage_aliyun',
  'storage_tencent',
  'storage_qiniu'
);

--bun:split

ALTER TABLE `system_settings`
    MODIFY `setting_value` VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '设置值';

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
