DELETE FROM system_settings WHERE setting_key IN (
    'site_name',
    'site_logo',
    'site_copyright',
    'site_icp',
    'site_seo_keywords',
    'site_description',
    'resource_api_key'
);
