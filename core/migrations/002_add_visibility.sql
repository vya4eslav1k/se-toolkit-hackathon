-- migrations/002_add_visibility.sql

ALTER TABLE polls ADD COLUMN IF NOT EXISTS visibility VARCHAR(50) NOT NULL DEFAULT 'public';
