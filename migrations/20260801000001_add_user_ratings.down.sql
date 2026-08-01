DELETE FROM system_settings WHERE setting_key = 'rating_enabled' AND setting_group = 'feature';

--bun:split

ALTER TABLE videos DROP COLUMN rating_count;

--bun:split

DROP TABLE IF EXISTS user_ratings;
