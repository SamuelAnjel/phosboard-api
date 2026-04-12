-- PHOS-17: Social Climate Data Model
-- Migration: social_mentions, document_temperatures, discovery_tasks

-- social_mentions: partitioned by publication date
CREATE TABLE IF NOT EXISTS social_mentions (
    id UUID DEFAULT uuid_generate_v4(),
    document_id UUID REFERENCES global_documents(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL,
    external_id VARCHAR(255),
    author_username VARCHAR(255),
    text_content TEXT,
    engagement_score INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    retweet_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    sentiment_score DECIMAL(5,4),
    sentiment_label VARCHAR(20),
    posted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    discovered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    raw_payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (id, posted_at)
) PARTITION BY RANGE (posted_at);

-- Partition for March 2026
CREATE TABLE IF NOT EXISTS social_mentions_2026_03 PARTITION OF social_mentions
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

-- Partition for April 2026
CREATE TABLE IF NOT EXISTS social_mentions_2026_04 PARTITION OF social_mentions
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

-- Indexes for social_mentions
CREATE INDEX IF NOT EXISTS idx_social_mentions_document_id ON social_mentions(document_id);
CREATE INDEX IF NOT EXISTS idx_social_mentions_platform ON social_mentions(platform);
CREATE INDEX IF NOT EXISTS idx_social_mentions_posted_at ON social_mentions(posted_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_mentions_sentiment ON social_mentions(sentiment_score);

-- document_temperatures: aggregate social metrics per document
CREATE TABLE IF NOT EXISTS document_temperatures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES global_documents(id) ON DELETE CASCADE UNIQUE,
    total_mentions INTEGER DEFAULT 0,
    total_engagement INTEGER DEFAULT 0,
    avg_sentiment_score DECIMAL(5,4),
    sentiment_distribution JSONB DEFAULT '{}',
    velocity_metrics JSONB DEFAULT '{}',
    temperature_score DECIMAL(5,4),
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for document_temperatures
CREATE INDEX IF NOT EXISTS idx_document_temperatures_document_id ON document_temperatures(document_id);
CREATE INDEX IF NOT EXISTS idx_document_temperatures_score ON document_temperatures(temperature_score DESC);

-- discovery_tasks: Outbox pattern for URLs to be scraped
CREATE TABLE IF NOT EXISTS discovery_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    url TEXT NOT NULL UNIQUE,
    source_type VARCHAR(50) NOT NULL,
    discovered_from_url TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    priority INTEGER DEFAULT 0,
    retry_count INTEGER DEFAULT 0,
    last_error TEXT,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for discovery_tasks
CREATE INDEX IF NOT EXISTS idx_discovery_tasks_status ON discovery_tasks(status, priority DESC);
CREATE INDEX IF NOT EXISTS idx_discovery_tasks_scheduled_at ON discovery_tasks(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_discovery_tasks_created_at ON discovery_tasks(created_at DESC);
