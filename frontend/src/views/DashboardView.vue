<script setup lang="ts">
import { onMounted } from 'vue'
import DashboardLayout from '../components/layouts/DashboardLayout.vue'
import DocumentList from '../components/DocumentList.vue'
import { useDocuments } from '../composables/useDocuments'

const TEST_TENANT_ID = '00000000-0000-0000-0000-000000000001'

const { documents, loading, error, fetchDocuments } = useDocuments()

onMounted(() => {
  fetchDocuments(TEST_TENANT_ID)
})
</script>

<template>
  <DashboardLayout>
    <div class="p-8">
      <h2 class="text-2xl font-semibold text-gray-900 dark:text-gray-100 mb-6">
        Documents
      </h2>
      <DocumentList 
        :documents="documents" 
        :loading="loading" 
        :error="error" 
      />
    </div>
  </DashboardLayout>
</template>