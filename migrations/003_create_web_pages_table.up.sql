CREATE TABLE web_pages (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    html TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add index on id column
CREATE INDEX idx_web_pages_id ON web_pages(id);

-- Add index on title column
CREATE INDEX idx_web_pages_title ON web_pages(title);