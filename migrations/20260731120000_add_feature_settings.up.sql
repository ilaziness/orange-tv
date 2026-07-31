-- Seed default feature settings

INSERT IGNORE INTO system_settings (setting_key, setting_group, setting_value, setting_type, description) VALUES
('live_enabled', 'feature', '0', 3, '电视直播开关'),
('comment_enabled', 'feature', '1', 3, '视频评论开关'),
('comment_review', 'feature', '1', 3, '评论是否需要审核');
