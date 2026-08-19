-- Revert system_settings.setting_value to its original size.
-- NOTE: this will fail if any value exceeds 512 bytes; only use before storing long codes.
ALTER TABLE `system_settings` MODIFY COLUMN `setting_value` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '设置值';
