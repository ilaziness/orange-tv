DELETE FROM `system_settings` WHERE `setting_key` IN (
  'seo_public_base_url',
  'seo_default_og_image',
  'seo_sitemap_enabled',
  'seo_llms_enabled',
  'seo_llms_intro',
  'seo_allow_ai_search',
  'seo_allow_ai_training',
  'seo_google_site_verification',
  'seo_baidu_site_verification',
  'seo_bing_site_verification'
);
