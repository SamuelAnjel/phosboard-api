-- Seed data for default sources
-- Run this after 00001_initial_schema.sql

INSERT INTO sources (name, type, config, created_at, updated_at)
SELECT 'Generic Web Scraper', 'scraper', '{"default": true}', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sources WHERE name = 'Generic Web Scraper');

INSERT INTO sources (name, type, config, created_at, updated_at)
SELECT 'RSS Feed Reader', 'rss', '{"default": true}', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sources WHERE name = 'RSS Feed Reader');

INSERT INTO sources (name, type, config, created_at, updated_at)
SELECT 'API Connector', 'api', '{"default": true}', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sources WHERE name = 'API Connector');