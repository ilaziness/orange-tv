-- Rollback: drop all tables in reverse order of creation.

DROP TABLE IF EXISTS online_sessions;

--bun:split

DROP TABLE IF EXISTS site_stats_daily;

--bun:split

DROP TABLE IF EXISTS banners;

--bun:split

DROP TABLE IF EXISTS video_comments;

--bun:split

DROP TABLE IF EXISTS user_play_history;

--bun:split

DROP TABLE IF EXISTS user_favorites;

--bun:split

DROP TABLE IF EXISTS user_login_logs;

--bun:split

DROP TABLE IF EXISTS system_logs;

--bun:split

DROP TABLE IF EXISTS login_logs;

--bun:split

DROP TABLE IF EXISTS users;

--bun:split

DROP TABLE IF EXISTS admins;

--bun:split

DROP TABLE IF EXISTS user_groups;

--bun:split

DROP TABLE IF EXISTS system_settings;

--bun:split

DROP TABLE IF EXISTS live_channels;

--bun:split

DROP TABLE IF EXISTS collect_logs;

--bun:split

DROP TABLE IF EXISTS collect_source_categories;

--bun:split

DROP TABLE IF EXISTS collect_sources;

--bun:split

DROP TABLE IF EXISTS play_episodes;

--bun:split

DROP TABLE IF EXISTS play_sources;

--bun:split

DROP TABLE IF EXISTS video_tags;

--bun:split

DROP TABLE IF EXISTS tags;

--bun:split

DROP TABLE IF EXISTS video_actors;

--bun:split

DROP TABLE IF EXISTS actors;

--bun:split

DROP TABLE IF EXISTS video_directors;

--bun:split

DROP TABLE IF EXISTS directors;

--bun:split

DROP TABLE IF EXISTS videos;

--bun:split

DROP TABLE IF EXISTS categories;
