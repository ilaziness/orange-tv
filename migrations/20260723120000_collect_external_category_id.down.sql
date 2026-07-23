-- reverse: external_category_id -> external_category

ALTER TABLE collect_source_categories DROP INDEX uk_source_external;
--bun:split
ALTER TABLE collect_source_categories DROP COLUMN external_category_id;
--bun:split
ALTER TABLE collect_source_categories ADD COLUMN external_category VARCHAR(100) NOT NULL DEFAULT '' COMMENT '外部分类名称（采集源返回的分类）' AFTER source_id;
--bun:split
ALTER TABLE collect_source_categories ADD UNIQUE KEY uk_source_external (source_id, external_category);
