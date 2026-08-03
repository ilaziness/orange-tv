UPDATE videos SET release_date = '' WHERE release_date IS NULL;
--bun:split
ALTER TABLE videos MODIFY COLUMN release_date VARCHAR(64) NOT NULL DEFAULT '' COMMENT '上映日期';
