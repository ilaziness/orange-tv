ALTER TABLE videos ADD COLUMN collect_source_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采集源ID，标识最初由哪个采集源采集' AFTER language;
--bun:split
ALTER TABLE videos ADD INDEX idx_collect_source (collect_source_id);
