--bun:split

ALTER TABLE live_channels ADD COLUMN deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间';
