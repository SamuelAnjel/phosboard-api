<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../plugins/axios'
import DefaultLayout from '../components/layouts/DefaultLayout.vue'

interface Document {
  id: string
  title: string
  url: string
  source_name: string
  semantic_analysis: {
    matched_concepts: string[]
    entities: string[]
    official_polarity: string
    search_queries: string[]
  } | null
  social_temperature: number | null
  created_at: string
}

interface DocumentsResponse {
  data: Document[]
  meta: {
    total: number
    limit: number
    offset: number
  }
}

const documents = ref<Document[]>([])
const loading = ref(true)
const error = ref('')
const pagination = ref({
  page: 1,
  itemsPerPage: 10,
  total: 0,
})

async function fetchDocuments() {
  loading.value = true
  error.value = ''

  try {
    const offset = (pagination.value.page - 1) * pagination.value.itemsPerPage
    const response = await api.get<DocumentsResponse>('/v1/documents', {
      params: {
        limit: pagination.value.itemsPerPage,
        offset: offset,
      },
    })
    documents.value = response.data.data
    pagination.value.total = response.data.meta.total
  } catch (e) {
    error.value = 'Failed to load documents'
  } finally {
    loading.value = false
  }
}

function getPolarityColor(polarity: string | undefined): string {
  if (!polarity) return 'grey'
  switch (polarity) {
    case 'positive': return 'success'
    case 'negative': return 'error'
    default: return 'grey'
  }
}

onMounted(() => {
  fetchDocuments()
})
</script>

<template>
  <DefaultLayout>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-6">Documentos</h1>
      </v-col>
    </v-row>

    <v-row v-if="loading">
      <v-col v-for="i in 5" :key="i" cols="12">
        <v-skeleton-loader type="card"></v-skeleton-loader>
      </v-col>
    </v-row>

    <v-row v-else-if="error">
      <v-col cols="12">
        <v-alert type="error">{{ error }}</v-alert>
      </v-col>
    </v-row>

    <v-row v-else-if="documents.length === 0">
      <v-col cols="12">
        <v-alert type="info">No se encontraron documentos</v-alert>
      </v-col>
    </v-row>

    <v-row v-else>
      <v-col v-for="doc in documents" :key="doc.id" cols="12">
        <v-card>
          <v-card-title class="d-flex align-center">
            <a :href="doc.url" target="_blank" class="text-decoration-none text-primary mr-2">
              {{ doc.title || 'Sin título' }}
            </a>
          </v-card-title>
          <v-card-subtitle>
            <v-chip size="small" class="mr-2">{{ doc.source_name }}</v-chip>
            <v-chip
              v-if="doc.semantic_analysis?.official_polarity"
              size="small"
              :color="getPolarityColor(doc.semantic_analysis.official_polarity)"
            >
              {{ doc.semantic_analysis.official_polarity }}
            </v-chip>
          </v-card-subtitle>
          <v-card-text>
            <p class="text-body-2 text-truncate mb-3">{{ doc.url }}</p>
            
            <div class="d-flex align-center mb-3">
              <span class="text-body-2 mr-2">Temperatura Social:</span>
              <v-progress-circular
                :model-value="doc.social_temperature || 0"
                :color="doc.social_temperature && doc.social_temperature > 50 ? 'error' : 'success'"
                size="40"
                width="4"
              >
                {{ doc.social_temperature?.toFixed(0) || 'N/A' }}
              </v-progress-circular>
            </div>

            <v-expansion-panels v-if="doc.semantic_analysis">
              <v-expansion-panel>
                <v-expansion-panel-title>
                  <v-icon size="small" class="mr-2">mdi-information</v-icon>
                  Análisis Semántico
                </v-expansion-panel-title>
                <v-expansion-panel-text>
                  <div v-if="doc.semantic_analysis.matched_concepts?.length">
                    <div class="text-subtitle-2 mb-1">Conceptos Matcheados:</div>
                    <v-chip
                      v-for="concept in doc.semantic_analysis.matched_concepts"
                      :key="concept"
                      size="small"
                      color="primary"
                      class="mr-1 mb-1"
                    >
                      {{ concept }}
                    </v-chip>
                  </div>
                  <div v-if="doc.semantic_analysis.entities?.length" class="mt-3">
                    <div class="text-subtitle-2 mb-1">Entidades:</div>
                    <v-chip
                      v-for="entity in doc.semantic_analysis.entities"
                      :key="entity"
                      size="small"
                      variant="outlined"
                      class="mr-1 mb-1"
                    >
                      {{ entity }}
                    </v-chip>
                  </div>
                  <div v-if="!doc.semantic_analysis.matched_concepts?.length && !doc.semantic_analysis.entities?.length" class="text-grey">
                    Sin análisis disponible
                  </div>
                </v-expansion-panel-text>
              </v-expansion-panel>
            </v-expansion-panels>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row v-if="!loading && !error && documents.length > 0">
      <v-col cols="12">
        <v-pagination
          v-model="pagination.page"
          :length="Math.ceil(pagination.total / pagination.itemsPerPage)"
          @update:model-value="fetchDocuments"
        ></v-pagination>
      </v-col>
    </v-row>
  </DefaultLayout>
</template>
