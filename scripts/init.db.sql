-- Create the greenlight role
CREATE ROLE greenlight WITH LOGIN PASSWORD 'pa55word';

-- Create the greenlight database owned by the greenlight role
CREATE DATABASE greenlight OWNER greenlight;

-- Grant all privileges to greenlight
GRANT ALL PRIVILEGES ON DATABASE greenlight to greenlight;

-- Connect to the greenlight database and enable citext
\connect greenlight

-- Create the citext extension
CREATE EXTENSION IF NOT EXISTS citext;
