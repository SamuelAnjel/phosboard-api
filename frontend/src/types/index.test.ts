import { describe, it, expect } from 'vitest'
import type {
  Tenant,
  Role,
  Source,
  GlobalDocument,
  TenantDocument,
  SentimentScore,
  DocumentAnalysis,
} from '../types'

describe('Types', () => {
  describe('Tenant', () => {
    it('has required properties', () => {
      const tenant: Tenant = {
        id: '123e4567-e89b-12d3-a456-426614174000',
        name: 'Acme Corp',
        created_at: new Date(),
      }
      expect(tenant.id).toBeDefined()
      expect(tenant.name).toBeDefined()
      expect(tenant.created_at).toBeDefined()
    })

    it('allows optional properties', () => {
      const tenant: Tenant = {
        id: '123e4567-e89b-12d3-a456-426614174000',
        name: 'Acme Corp',
        slug: 'acme',
        settings: { theme: 'dark' },
        created_at: '2024-01-01',
        updated_at: new Date(),
      }
      expect(tenant.slug).toBe('acme')
      expect(tenant.settings).toEqual({ theme: 'dark' })
    })
  })

  describe('Role', () => {
    it('has required properties', () => {
      const role: Role = {
        id: '123e4567-e89b-12d3-a456-426614174000',
        tenant_id: '123e4567-e89b-12d3-a456-426614174001',
        name: 'Admin',
        created_at: new Date(),
      }
      expect(role.id).toBeDefined()
      expect(role.tenant_id).toBeDefined()
      expect(role.name).toBeDefined()
    })
  })

  describe('Source', () => {
    it('has required properties', () => {
      const source: Source = {
        id: '123e4567-e89b-12d3-a456-426614174000',
        tenant_id: '123e4567-e89b-12d3-a456-426614174001',
        name: 'My Blog',
        type: 'wordpress',
        created_at: new Date(),
      }
      expect(source.id).toBeDefined()
      expect(source.name).toBeDefined()
      expect(source.type).toBeDefined()
    })
  })

  describe('GlobalDocument', () => {
    it('has required properties', () => {
      const doc: GlobalDocument = {
        id: '123e4567-e89b-12d3-a456-426614174000',
        source_id: '123e4567-e89b-12d3-a456-426614174001',
        title: 'Test Article',
        url: 'https://example.com/article',
        content_text: 'Article content here',
        created_at: new Date(),
      }
      expect(doc.id).toBeDefined()
      expect(doc.title).toBeDefined()
      expect(doc.url).toBeDefined()
    })
  })

  describe('TenantDocument', () => {
    it('has composite key properties', () => {
      const td: TenantDocument = {
        tenant_id: '123e4567-e89b-12d3-a456-426614174000',
        document_id: '123e4567-e89b-12d3-a456-426614174001',
        matched_keywords: ['ai', 'machine learning'],
        created_at: new Date(),
      }
      expect(td.tenant_id).toBeDefined()
      expect(td.document_id).toBeDefined()
      expect(td.matched_keywords).toHaveLength(2)
    })
  })

  describe('DocumentAnalysis', () => {
    it('has correct risk level values', () => {
      const analysis: DocumentAnalysis = {
        sentiment: 0.5 as SentimentScore,
        entities: ['Go', 'Vue'],
        risk_level: 'low',
      }
      expect(analysis.risk_level).toBe('low')

      const mediumAnalysis: DocumentAnalysis = {
        sentiment: -0.3 as SentimentScore,
        entities: [],
        risk_level: 'medium',
      }
      expect(mediumAnalysis.risk_level).toBe('medium')

      const highAnalysis: DocumentAnalysis = {
        sentiment: -0.9 as SentimentScore,
        entities: ['threat'],
        risk_level: 'high',
      }
      expect(highAnalysis.risk_level).toBe('high')
    })

    it('accepts readonly entities', () => {
      const analysis: DocumentAnalysis = {
        sentiment: 0.0 as SentimentScore,
        entities: ['Entity1', 'Entity2'] as const,
        risk_level: 'low',
      }
      expect(analysis.entities).toBeDefined()
    })
  })
})