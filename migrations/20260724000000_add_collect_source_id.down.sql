ALTER TABLE videos DROP INDEX idx_collect_source;
--bun:split
ALTER TABLE videos DROP COLUMN collect_source_id;
