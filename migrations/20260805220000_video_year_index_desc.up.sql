-- Optimize year index: year DESC + id DESC composite to match the default sort
-- "year DESC, id DESC", eliminating filesort for the most common query pattern.

DROP INDEX `idx_year` ON `videos`;

--bun:split

CREATE INDEX `idx_year_desc` ON `videos` (`year` DESC, `id` DESC);
