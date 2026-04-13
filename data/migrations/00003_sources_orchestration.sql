-- Add orchestration fields to sources table
-- Run after 00001_initial_schema.sql

ALTER TABLE sources 
ADD COLUMN IF NOT EXISTS url TEXT,
ADD COLUMN IF NOT EXISTS fetch_strategy VARCHAR(20) DEFAULT 'rss',
ADD COLUMN IF NOT EXISTS interval_minutes INTEGER DEFAULT 60,
ADD COLUMN IF NOT EXISTS last_run_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_sources_last_run_at ON sources(last_run_at);