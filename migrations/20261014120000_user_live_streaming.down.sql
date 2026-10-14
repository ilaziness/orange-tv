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
    DROP INDEX `idx_users_is_streamer`,
    DROP COLUMN `streamer_at`,
    DROP COLUMN `is_streamer`;
