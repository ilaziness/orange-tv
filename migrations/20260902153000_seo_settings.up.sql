INSERT IGNORE INTO `system_settings` (`setting_key`, `setting_group`, `setting_value`, `setting_type`, `description`, `created_at`, `updated_at`) VALUES
('seo_public_base_url', 'seo', '', 1, '公开站点根地址（无尾斜杠）', NOW(), NOW()),
('seo_default_og_image', 'seo', '', 1, '默认 Open Graph 图片地址', NOW(), NOW()),
('seo_sitemap_enabled', 'seo', '1', 3, '是否输出 sitemap', NOW(), NOW()),
('seo_llms_enabled', 'seo', '1', 3, '是否输出 llms.txt', NOW(), NOW()),
('seo_llms_intro', 'seo', '', 1, 'llms.txt 站点简介', NOW(), NOW()),
('seo_allow_ai_search', 'seo', '1', 3, '是否允许 AI 检索类爬虫', NOW(), NOW()),
('seo_allow_ai_training', 'seo', '0', 3, '是否允许 AI 训练类爬虫', NOW(), NOW()),
('seo_google_site_verification', 'seo', '', 1, 'Google 站点验证码', NOW(), NOW()),
('seo_baidu_site_verification', 'seo', '', 1, '百度站点验证码', NOW(), NOW()),
('seo_bing_site_verification', 'seo', '', 1, 'Bing 站点验证码', NOW(), NOW());
