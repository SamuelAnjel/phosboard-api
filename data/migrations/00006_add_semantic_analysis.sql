-- PHOS-27: Add semantic_analysis column to global_documents
ALTER TABLE global_documents ADD COLUMN IF NOT EXISTS semantic_analysis JSONB DEFAULT '{}'::jsonb;
