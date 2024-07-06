ALTER TABLE galleries
ADD COLUMN gallery_type VARCHAR(255) default 'main';

CREATE INDEX gallery_type_index ON galleries (gallery_type);