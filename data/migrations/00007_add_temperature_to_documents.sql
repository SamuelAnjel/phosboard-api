-- PHOS-30: Add social_temperature to global_documents
ALTER TABLE global_documents ADD COLUMN IF NOT EXISTS social_temperature NUMERIC(5,2) DEFAULT 0.00;
