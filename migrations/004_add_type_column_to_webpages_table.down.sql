DROP INDEX IF EXISTS page_type_index;

ALTER TABLE web_pages
DROP COLUMN IF EXISTS page_type;