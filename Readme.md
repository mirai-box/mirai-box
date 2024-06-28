# Mirai Box Service

Add admin user:

```
docker exec -it postgres
DROP DATABASE IF EXISTS picture_db;

-- Create the user
CREATE USER picture_db WITH PASSWORD 'picture_db';

-- Create the database
CREATE DATABASE picture_db OWNER picture_db;

-- Grant privileges on the database
GRANT ALL PRIVILEGES ON DATABASE picture_db TO picture_db;

-- Connect to the newly created database
\c picture_db

-- Grant privileges on the public schema
GRANT ALL PRIVILEGES ON SCHEMA public TO picture_db;

-- Set default privileges in the public schema
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO picture_db;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO picture_db;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO picture_db;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
INSERT INTO users (id, username, password, role) 
VALUES (uuid_generate_v4(), 'igor', '$2a$10$8ZRj1Wd/euezkJRHfS.TaesICxuOCA.26Sx9V57WQz5VRTA6w40TW', 'admin');
```
