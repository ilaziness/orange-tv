DROP INDEX idx_setting_group ON system_settings;

--bun:split

ALTER TABLE system_settings DROP COLUMN setting_group;
