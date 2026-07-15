-- Phase 4: site / API / resource station settings seeds
INSERT INTO system_settings (setting_key, setting_value, setting_type, description)
SELECT 'site_name', 'Orange TV', 1, '站点名称'
WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE setting_key = 'site_name');
--bun:split
INSERT INTO system_settings (setting_key, setting_value, setting_type, description)
SELECT 'site_logo', '', 1, '站点 Logo URL'
WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE setting_key = 'site_logo');
--bun:split
INSERT INTO system_settings (setting_key, setting_value, setting_type, description)
SELECT 'site_copyright', '', 1, '站点版权信息'
WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE setting_key = 'site_copyright');
--bun:split
INSERT INTO system_settings (setting_key, setting_value, setting_type, description)
SELECT 'site_icp', '', 1, '备案号'
WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE setting_key = 'site_icp');
--bun:split
INSERT INTO system_settings (setting_key, setting_value, setting_type, description)
SELECT 'site_seo_keywords', '', 1, 'SEO 关键词'
WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE setting_key = 'site_seo_keywords');
--bun:split
INSERT INTO system_settings (setting_key, setting_value, setting_type, description)
SELECT 'site_description', '', 1, '站点描述'
WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE setting_key = 'site_description');
--bun:split
INSERT INTO system_settings (setting_key, setting_value, setting_type, description)
SELECT 'resource_api_key', '', 1, '资源站 API 访问密钥'
WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE setting_key = 'resource_api_key');
