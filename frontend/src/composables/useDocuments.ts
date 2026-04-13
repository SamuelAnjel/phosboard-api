import { ref } from 'vue'
import type { DocumentWithSource } from '../types'

interface ApiResponse<T> {
  data?: T
  error?: string
}

export function useDocuments() {
  const documents = ref<DocumentWithSource[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchDocuments(tenantId: string): Promise<void> {
    loading.value = true
    error.value = null

    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/documents`)
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const result: ApiResponse<DocumentWithSource[]> = await response.json()

      if (result.error) {
        throw new Error(result.error)
      }

      documents.value = result.data ?? []
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch documents'
      documents.value = []
    } finally {
      loading.value = false
    }
  }

  return {
    documents,
    loading,
    error,
    fetchDocuments,
  }
}