-- PHOS-26: Tenant Concepts (Seed Concepts)
-- Tabla para almacenar conceptos/palabras clave por tenant

CREATE TABLE IF NOT EXISTS tenant_concepts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    concept_term VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (tenant_id, concept_term)
);

CREATE INDEX IF NOT EXISTS idx_tenant_concepts_tenant_id ON tenant_concepts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_concepts_is_active ON tenant_concepts(is_active) WHERE is_active = TRUE;
