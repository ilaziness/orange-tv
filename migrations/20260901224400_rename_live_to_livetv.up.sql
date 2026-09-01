RENAME TABLE `live_channels` TO `livetv_channels`;

--bun:split

UPDATE `system_settings`
SET `setting_key` = 'livetv_enabled'
WHERE `setting_key` = 'live_enabled';

--bun:split

UPDATE `system_settings`
SET `setting_key` = 'livetv_sync_source_url',
    `setting_group` = 'livetv'
WHERE `setting_key` = 'live_sync_source_url';

--bun:split

UPDATE `system_settings`
SET `setting_group` = 'livetv'
WHERE `setting_group` = 'live';
