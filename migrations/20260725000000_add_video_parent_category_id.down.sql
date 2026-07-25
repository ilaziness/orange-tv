ALTER TABLE videos DROP INDEX idx_parent_category;
--bun:split
ALTER TABLE videos DROP COLUMN parent_category_id;
