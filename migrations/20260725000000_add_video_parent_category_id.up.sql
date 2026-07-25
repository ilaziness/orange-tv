ALTER TABLE videos ADD COLUMN parent_category_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父分类ID，方便大分类筛选，0表示无父分类' AFTER category_id;
--bun:split
ALTER TABLE videos ADD INDEX idx_parent_category (parent_category_id);
