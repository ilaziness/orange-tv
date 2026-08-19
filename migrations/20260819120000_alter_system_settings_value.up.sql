-- Expand system_settings.setting_value to accommodate analytics/tracking code snippets.
ALTER TABLE `system_settings` MODIFY COLUMN `setting_value` VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '设置值';
