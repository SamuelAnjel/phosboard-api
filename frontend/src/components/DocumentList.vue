<script setup lang="ts">
import { computed } from 'vue'
import type { DocumentWithSource } from '../types'
import DocumentCard from './DocumentCard.vue'

interface Props {
  documents: DocumentWithSource[]
  loading: boolean
  error: string | null
}

const props = defineProps<Props>()

const hasDocuments = computed(() => props.documents.length > 0)
</script>

<template>
  <div class="space-y-4">
    <!-- Loading State -->
    <div v-if="loading" class="space-y-4">
      <div 
        v-for="i in 3" 
        :key="i"
        class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 animate-pulse"
      >
        <div class="h-6 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-3"></div>
        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-3"></div>
        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
      </div>
    </div>

    <!-- Error State -->
    <div 
      v-else-if="error" 
      class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 text-red-600 dark:text-red-400"
    >
      <p class="font-medium">Error loading documents</p>
      <p class="text-sm mt-1">{{ error }}</p>
    </div>

    <!-- Empty State -->
    <div 
      v-else-if="!hasDocuments"
      class="text-center py-12 text-gray-500 dark:text-gray-400"
    >
      <p class="text-lg">No documents found</p>
      <p class="text-sm mt-1">Documents will appear here once they're processed</p>
    </div>

    <!-- Document List -->
    <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <DocumentCard 
        v-for="doc in documents" 
        :key="doc.id" 
        :document="doc" 
      />
    </div>
  </div>
</template>