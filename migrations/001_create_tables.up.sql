-- Up Migration

-- Create the pictures table first without foreign keys
CREATE TABLE pictures (
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
CREATE TABLE revisions (
    id UUID PRIMARY KEY,
    picture_id UUID,
    version INT,
    file_path VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    comment TEXT,
    art_id VARCHAR(255)
);

-- Create the remaining tables
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL
);

CREATE TABLE galleries (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published BOOLEAN DEFAULT FALSE
);

CREATE TABLE gallery_images (
    gallery_id UUID NOT NULL,
    revision_id UUID NOT NULL,
    PRIMARY KEY (gallery_id, revision_id),
    FOREIGN KEY (gallery_id) REFERENCES galleries(id) ON DELETE CASCADE,
    FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE CASCADE
);

-- Add foreign keys after both tables are created
ALTER TABLE pictures
    ADD CONSTRAINT fk_latest_revision_id
    FOREIGN KEY (latest_revision_id) REFERENCES revisions(id);

ALTER TABLE pictures
    ADD CONSTRAINT fk_published_revision_id
    FOREIGN KEY (published_revision_id) REFERENCES revisions(id);

ALTER TABLE revisions
    ADD CONSTRAINT fk_picture_id
    FOREIGN KEY (picture_id) REFERENCES pictures(id) ON DELETE CASCADE;
