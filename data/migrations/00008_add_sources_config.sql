-- Add config JSONB column to sources table
-- Run after 00003_sources_orchestration.sql

ALTER TABLE sources 
ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::jsonb;
