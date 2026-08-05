-- Revert to the original single-column ascending year index.

DROP INDEX `idx_year_desc` ON `videos`;

--bun:split

CREATE INDEX `idx_year` ON `videos` (`year`);
