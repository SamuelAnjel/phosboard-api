-- Migration 00010: Social Tracks (Crisis Tracking System)
-- Creates tables for event-driven social media tracking by topic/crisis

CREATE TABLE IF NOT EXISTS social_tracks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    terms TEXT[] NOT NULL,  -- Array of search terms
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'completed')),
    source_document_id UUID REFERENCES global_documents(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_social_tracks_tenant_id ON social_tracks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_social_tracks_status ON social_tracks(status);
CREATE INDEX IF NOT EXISTS idx_social_tracks_created_at ON social_tracks(created_at DESC);

-- Optional pivot table for N:N relationship between documents and tracks
CREATE TABLE IF NOT EXISTS document_social_tracks (
    document_id UUID NOT NULL REFERENCES global_documents(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES social_tracks(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) DEFAULT 'source' CHECK (relationship_type IN ('source', 'related', 'monitored')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (document_id, track_id)
);

-- Index for querying tracks by document
CREATE INDEX IF NOT EXISTS idx_document_social_tracks_document_id ON document_social_tracks(document_id);
CREATE INDEX IF NOT EXISTS idx_document_social_tracks_track_id ON document_social_tracks(track_id);

-- Add comment explaining the purpose
COMMENT ON TABLE social_tracks IS 'Tracks for crisis/social media monitoring campaigns';
COMMENT ON COLUMN social_tracks.terms IS 'Array of search terms for social media monitoring';
COMMENT ON COLUMN social_tracks.status IS 'Track status: active, paused, completed';
COMMENT ON TABLE document_social_tracks IS 'Many-to-many relationship between documents and social tracks';