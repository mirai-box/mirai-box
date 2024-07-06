-- Up Migration

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
INSERT INTO users (id, username, password, role) 
VALUES (
    uuid_generate_v4(), 
    'igor', 
    '$2a$10$8ZRj1Wd/euezkJRHfS.TaesICxuOCA.26Sx9V57WQz5VRTA6w40TW', 
    'admin'
)
ON CONFLICT (username) DO NOTHING;  -- This prevents error if the user already exists

-- Insert a gallery with the title "Main"
INSERT INTO galleries (id, title, published)
VALUES (
    uuid_generate_v4(),
    'Main',
    true
)
ON CONFLICT (id) DO NOTHING;  -- This prevents error if a gallery with this ID already exists
