-- Migration 00011: Add social temperature to social tracks
-- Adds temperature tracking for social media monitoring campaigns

-- Add social_temperature column to social_tracks table
ALTER TABLE social_tracks 
ADD COLUMN IF NOT EXISTS social_temperature DECIMAL(5,2) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS last_temperature_update TIMESTAMP WITH TIME ZONE DEFAULT NULL;

-- Add comment explaining the new column
COMMENT ON COLUMN social_tracks.social_temperature IS 'Aggregated social media temperature score (0-100)';
COMMENT ON COLUMN social_tracks.last_temperature_update IS 'Timestamp of last temperature calculation';