ALTER TABLE web_pages
ADD COLUMN page_type VARCHAR(255) default 'main';

CREATE INDEX page_type_index ON web_pages (page_type);