-- revert collect_logs

ALTER TABLE collect_logs MODIFY COLUMN status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '采集状态：1成功 2失败 3部分成功';
--bun:split
ALTER TABLE collect_logs DROP COLUMN duration_sec;
--bun:split
ALTER TABLE collect_logs DROP COLUMN collect_count;
--bun:split
ALTER TABLE collect_logs ADD COLUMN duration_ms INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '耗时(毫秒)';
--bun:split
ALTER TABLE collect_logs ADD COLUMN error_message TEXT COMMENT '错误信息';
--bun:split
ALTER TABLE collect_logs ADD COLUMN failed_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '失败数';
--bun:split
ALTER TABLE collect_logs ADD COLUMN success_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成功数';
--bun:split
ALTER TABLE collect_logs ADD COLUMN total_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采集总数';
--bun:split

-- revert collect_sources

ALTER TABLE collect_sources ADD COLUMN config JSON COMMENT '采集配置';
--bun:split
ALTER TABLE collect_sources DROP COLUMN data_range;
--bun:split
ALTER TABLE collect_sources DROP COLUMN schedule_enabled;
