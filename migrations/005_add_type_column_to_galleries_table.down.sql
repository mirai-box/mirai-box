DROP INDEX IF EXISTS gallery_type_index;

ALTER TABLE galleries
DROP COLUMN IF EXISTS gallery_type;