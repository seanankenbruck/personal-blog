-- Database initialization script
-- This script runs when the PostgreSQL container starts for the first time

-- Create the blog_user role if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'blog_user') THEN
        CREATE USER blog_user WITH PASSWORD 'blog_password';
    END IF;
END
$$;

-- Create the blog_db database if it doesn't exist
SELECT 'CREATE DATABASE blog_db OWNER blog_user'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'blog_db')\gexec

-- Grant all privileges on the database to blog_user
GRANT ALL PRIVILEGES ON DATABASE blog_db TO blog_user;

-- Connect to the blog_db database
\c blog_db;

-- Grant schema privileges to blog_user
GRANT ALL ON SCHEMA public TO blog_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO blog_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO blog_user;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO blog_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO blog_user;

-- Configure authentication for external connections
-- This allows the blog_user to connect from any host
ALTER USER blog_user WITH LOGIN; 