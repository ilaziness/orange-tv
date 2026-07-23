-- collect_source_categories: external_category (name string) -> external_category_id (int)

ALTER TABLE collect_source_categories DROP INDEX uk_source_external;
--bun:split
ALTER TABLE collect_source_categories DROP COLUMN external_category;
--bun:split
ALTER TABLE collect_source_categories ADD COLUMN external_category_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '外部分类ID（采集源返回的 type_id）' AFTER source_id;
--bun:split
ALTER TABLE collect_source_categories ADD UNIQUE KEY uk_source_external (source_id, external_category_id);
