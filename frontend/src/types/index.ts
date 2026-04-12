export interface Tenant {
  readonly id: string
  name: string
  slug?: string
  settings?: Record<string, unknown>
  created_at: Date | string
  updated_at?: Date | string
}

export interface Role {
  readonly id: string
  tenant_id: string
  name: string
  description?: string
  permissions?: Record<string, unknown>
  created_at: Date | string
  updated_at?: Date | string
}

export interface TenantUser {
  readonly id: string
  tenant_id: string
  user_id: string
  role_id: string
  created_at: Date | string
  updated_at?: Date | string
}

export interface Source {
  readonly id: string
  tenant_id: string
  name: string
  type: string
  config?: Record<string, unknown>
  created_at: Date | string
  updated_at?: Date | string
}

export interface GlobalDocument {
  readonly id: string
  source_id: string
  title: string
  url: string
  content_text: string
  raw_payload?: Record<string, unknown>
  content_embedding?: number[]
  created_at: Date | string
  updated_at?: Date | string
}

export interface DocumentWithSource {
  id: string
  title: string
  url: string
  source_name: string
}

export interface TenantDocument {
  readonly tenant_id: string
  readonly document_id: string
  matched_keywords?: string[]
  routed_at?: Date | string
  created_at: Date | string
  updated_at?: Date | string
}

export type SentimentScore = number & { __brand: 'SentimentScore' }

export interface DocumentAnalysis {
  sentiment: SentimentScore
  entities: readonly string[]
  risk_level: 'low' | 'medium' | 'high'
}