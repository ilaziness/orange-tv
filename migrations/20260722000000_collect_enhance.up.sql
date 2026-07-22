-- collect_sources: add schedule fields, drop config

ALTER TABLE collect_sources ADD COLUMN schedule_enabled TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否开启定时采集：1是 0否';
--bun:split
ALTER TABLE collect_sources ADD COLUMN data_range VARCHAR(20) NOT NULL DEFAULT '' COMMENT '定时采集数据范围：today/last1d/last3d/last1w/last1m/all';
--bun:split
ALTER TABLE collect_sources DROP COLUMN config;
--bun:split

-- collect_logs: simplify status, replace count fields, change duration to seconds

ALTER TABLE collect_logs DROP COLUMN total_count;
--bun:split
ALTER TABLE collect_logs DROP COLUMN success_count;
--bun:split
ALTER TABLE collect_logs DROP COLUMN failed_count;
--bun:split
ALTER TABLE collect_logs DROP COLUMN error_message;
--bun:split
ALTER TABLE collect_logs DROP COLUMN duration_ms;
--bun:split
ALTER TABLE collect_logs ADD COLUMN collect_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采集数量（累加）';
--bun:split
ALTER TABLE collect_logs ADD COLUMN duration_sec INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '耗时(秒)';
--bun:split
ALTER TABLE collect_logs MODIFY COLUMN status TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '采集状态：1完成 2采集中';
