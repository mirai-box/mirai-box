-- Up Migration

-- Create the pictures table first without foreign keys
CREATE TABLE IF NOT EXISTS pictures (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    content_type VARCHAR(255) DEFAULT NULL,
    filename VARCHAR(255) DEFAULT NULL,
    latest_revision_id UUID NULL,
    published_revision_id UUID NULL,
    tags JSON
);

-- Create the revisions table next without foreign keys
CREATE TABLE IF NOT EXISTS revisions (
    id UUID PRIMARY KEY,
    picture_id UUID,
    version INT,
    file_path VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    comment TEXT,
    art_id VARCHAR(255)
);

-- Create the remaining tables
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS galleries (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS gallery_images (
    gallery_id UUID NOT NULL,
    revision_id UUID NOT NULL,
    PRIMARY KEY (gallery_id, revision_id),
    FOREIGN KEY (gallery_id) REFERENCES galleries(id) ON DELETE CASCADE,
    FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE CASCADE
);

-- Add foreign keys after both tables are created
ALTER TABLE pictures
    ADD CONSTRAINT fk_latest_revision_id
    FOREIGN KEY (latest_revision_id) REFERENCES revisions(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE pictures
    ADD CONSTRAINT fk_published_revision_id
    FOREIGN KEY (published_revision_id) REFERENCES revisions(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE revisions
    ADD CONSTRAINT fk_picture_id
    FOREIGN KEY (picture_id) REFERENCES pictures(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;

-- Add an index on revisions(picture_id) for better performance
CREATE INDEX idx_revisions_picture_id ON revisions(picture_id);

-- Add an index on pictures(latest_revision_id) for better performance
CREATE INDEX idx_pictures_latest_revision_id ON pictures(latest_revision_id);

-- Add an index on pictures(published_revision_id) for better performance
CREATE INDEX idx_pictures_published_revision_id ON pictures(published_revision_id);