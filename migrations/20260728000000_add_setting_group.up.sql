-- Add setting_group column to system_settings

ALTER TABLE system_settings ADD COLUMN setting_group VARCHAR(50) NOT NULL DEFAULT '' COMMENT '设置分组' AFTER setting_key;

--bun:split

CREATE INDEX idx_setting_group ON system_settings(setting_group);

--bun:split

-- Backfill existing 10 seed records with their group values
UPDATE system_settings SET setting_group = 'site' WHERE setting_key IN ('site_name', 'site_logo', 'site_copyright', 'site_icp', 'site_seo_keywords', 'site_description');

--bun:split

UPDATE system_settings SET setting_group = 'api' WHERE setting_key IN ('site_mode', 'api_output_format', 'enable_third_party_collect', 'resource_api_key');
