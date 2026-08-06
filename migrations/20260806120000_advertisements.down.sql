-- Drop advertisements table and restore old ad settings group placeholder.
-- Note: old ad settings data is NOT restored (it was a single config row, now replaced by per-ad rows).

DROP TABLE IF EXISTS `advertisements`;
