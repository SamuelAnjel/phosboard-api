<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../plugins/axios'
import DefaultLayout from '../components/layouts/DefaultLayout.vue'

interface Concept {
  id: string
  concept_term: string
  is_active: boolean
  created_at: string
}

const concepts = ref<Concept[]>([])
const loading = ref(true)
const error = ref('')
const newConcept = ref('')

async function fetchConcepts() {
  loading.value = true
  error.value = ''

  try {
    const response = await api.get<{ data: Concept[] }>('/v1/tenants/85c5f582-86b1-4217-bd4a-e1b1d0aac195/concepts')
    concepts.value = response.data.data
  } catch (e) {
    error.value = 'Failed to load concepts'
  } finally {
    loading.value = false
  }
}

async function addConcept() {
  if (!newConcept.value.trim()) return

  try {
    await api.post('/v1/tenants/85c5f582-86b1-4217-bd4a-e1b1d0aac195/concepts', {
      concept_term: newConcept.value,
    })
    newConcept.value = ''
    fetchConcepts()
  } catch (e) {
    error.value = 'Failed to add concept'
  }
}

async function deleteConcept(id: string) {
  try {
    await api.delete(`/v1/tenants/85c5f582-86b1-4217-bd4a-e1b1d0aac195/concepts/${id}`)
    fetchConcepts()
  } catch (e) {
    error.value = 'Failed to delete concept'
  }
}

onMounted(() => {
  fetchConcepts()
})
</script>

<template>
  <DefaultLayout>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-6">Conceptos Semilla</h1>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>Agregar Concepto</v-card-title>
          <v-card-text>
            <v-text-field
              v-model="newConcept"
              label="Nuevo concepto"
              placeholder="Ej: Seguridad, Tránsito, Eventos"
              variant="outlined"
              @keyup.enter="addConcept"
            ></v-text-field>
            <v-btn color="primary" @click="addConcept" :disabled="!newConcept.trim()">
              Agregar
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>Conceptos Activos</v-card-title>
          <v-card-text>
            <v-row v-if="loading">
              <v-col v-for="i in 3" :key="i" cols="12" md="4">
                <v-skeleton-loader type="chip"></v-skeleton-loader>
              </v-col>
            </v-row>
            <div v-else-if="error" class="text-error">{{ error }}</div>
            <div v-else-if="concepts.length === 0" class="text-grey">
              No hay conceptos configurados
            </div>
            <v-chip
              v-else
              v-for="concept in concepts"
              :key="concept.id"
              closable
              color="primary"
              class="mr-2 mb-2"
              @click:close="deleteConcept(concept.id)"
            >
              {{ concept.concept_term }}
            </v-chip>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </DefaultLayout>
</template>
