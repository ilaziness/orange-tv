UPDATE `system_settings`
SET `setting_group` = 'live'
WHERE `setting_group` = 'livetv';

--bun:split

UPDATE `system_settings`
SET `setting_key` = 'live_sync_source_url',
    `setting_group` = 'live'
WHERE `setting_key` = 'livetv_sync_source_url';

--bun:split

UPDATE `system_settings`
SET `setting_key` = 'live_enabled'
WHERE `setting_key` = 'livetv_enabled';

--bun:split

RENAME TABLE `livetv_channels` TO `live_channels`;
