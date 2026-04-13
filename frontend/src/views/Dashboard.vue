<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
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

const totalDocuments = computed(() => documents.value.length)
const avgTemperature = computed(() => {
  const temps = documents.value.filter(d => d.social_temperature !== null).map(d => d.social_temperature!)
  if (temps.length === 0) return 0
  return (temps.reduce((a, b) => a + b, 0) / temps.length).toFixed(1)
})
const topConcepts = computed(() => {
  const conceptCounts: Record<string, number> = {}
  documents.value.forEach(doc => {
    if (doc.semantic_analysis?.matched_concepts) {
      doc.semantic_analysis.matched_concepts.forEach((concept: string) => {
        conceptCounts[concept] = (conceptCounts[concept] || 0) + 1
      })
    }
  })
  return Object.entries(conceptCounts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
    .map(([concept, count]) => ({ concept, count }))
})

async function fetchDocuments() {
  loading.value = true
  error.value = ''

  try {
    const response = await api.get<DocumentsResponse>('/v1/documents')
    documents.value = response.data.data
  } catch (e) {
    error.value = 'Failed to load documents'
  } finally {
    loading.value = false
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
        <h1 class="text-h4 mb-6">Dashboard</h1>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="4">
        <v-card>
          <v-card-text class="d-flex align-center">
            <v-avatar color="primary" size="56" class="mr-4">
              <v-icon size="28" color="white">mdi-file-document-multiple</v-icon>
            </v-avatar>
            <div>
              <div class="text-h4">{{ totalDocuments }}</div>
              <div class="text-body-2 text-grey">Total Documentos</div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card>
          <v-card-text class="d-flex align-center">
            <v-avatar color="warning" size="56" class="mr-4">
              <v-icon size="28" color="white">mdi-thermometer</v-icon>
            </v-avatar>
            <div>
              <div class="text-h4">{{ avgTemperature }}°</div>
              <div class="text-body-2 text-grey">Temperatura Promedio</div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card>
          <v-card-text class="d-flex align-center">
            <v-avatar color="success" size="56" class="mr-4">
              <v-icon size="28" color="white">mdi-tag-multiple</v-icon>
            </v-avatar>
            <div>
              <div class="text-h4">{{ topConcepts.length }}</div>
              <div class="text-body-2 text-grey">Conceptos Matcheados</div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row v-if="topConcepts.length > 0" class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>Top Conceptos</v-card-title>
          <v-card-text>
            <v-chip
              v-for="item in topConcepts"
              :key="item.concept"
              class="mr-2 mb-2"
              color="primary"
            >
              {{ item.concept }} ({{ item.count }})
            </v-chip>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <v-col cols="12">
        <h2 class="text-h5 mb-4">Documentos Recientes</h2>
      </v-col>
    </v-row>

    <v-row v-if="loading">
      <v-col v-for="i in 3" :key="i" cols="12">
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
          <v-card-title>
            <a :href="doc.url" target="_blank" class="text-decoration-none">
              {{ doc.title || 'Sin título' }}
            </a>
          </v-card-title>
          <v-card-subtitle>
            <v-chip size="small" class="mr-2">{{ doc.source_name }}</v-chip>
            <v-chip
              v-if="doc.semantic_analysis?.official_polarity"
              size="small"
              :color="doc.semantic_analysis.official_polarity === 'positive' ? 'success' : doc.semantic_analysis.official_polarity === 'negative' ? 'error' : 'grey'"
            >
              {{ doc.semantic_analysis.official_polarity }}
            </v-chip>
          </v-card-subtitle>
          <v-card-text>
            <div class="d-flex align-center">
              <span class="text-body-2 mr-4">Temperatura:</span>
              <v-progress-circular
                :model-value="doc.social_temperature || 0"
                :color="doc.social_temperature && doc.social_temperature > 50 ? 'error' : 'success'"
                size="32"
                width="4"
              >
                {{ doc.social_temperature?.toFixed(0) || 0 }}
              </v-progress-circular>
            </div>
            <v-expansion-panels v-if="doc.semantic_analysis" class="mt-3">
              <v-expansion-panel>
                <v-expansion-panel-title>Ver Detalles</v-expansion-panel-title>
                <v-expansion-panel-text>
                  <div v-if="doc.semantic_analysis.matched_concepts?.length">
                    <strong>Conceptos:</strong>
                    <v-chip
                      v-for="concept in doc.semantic_analysis.matched_concepts"
                      :key="concept"
                      size="small"
                      class="mr-1 mt-1"
                    >
                      {{ concept }}
                    </v-chip>
                  </div>
                  <div v-if="doc.semantic_analysis.entities?.length" class="mt-2">
                    <strong>Entidades:</strong>
                    <v-chip
                      v-for="entity in doc.semantic_analysis.entities"
                      :key="entity"
                      size="small"
                      variant="outlined"
                      class="mr-1 mt-1"
                    >
                      {{ entity }}
                    </v-chip>
                  </div>
                </v-expansion-panel-text>
              </v-expansion-panel>
            </v-expansion-panels>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </DefaultLayout>
</template>
