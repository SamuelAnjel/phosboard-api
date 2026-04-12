<script setup lang="ts">
import type { DocumentWithSource } from '../types'

interface Props {
  document: DocumentWithSource
}

const props = defineProps<Props>()

function truncateUrl(url: string, maxLength = 60): string {
  if (url.length <= maxLength) return url
  return url.slice(0, maxLength) + '...'
}

function handleClick() {
  window.open(props.document.url, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <article 
    class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md transition-shadow cursor-pointer"
    @click="handleClick"
  >
    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2 line-clamp-2">
      {{ document.title }}
    </h3>
    
    <div class="flex items-center gap-2 mb-3">
      <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200">
        {{ document.source_name }}
      </span>
    </div>
    
    <p class="text-sm text-gray-500 dark:text-gray-400 truncate" :title="document.url">
      {{ truncateUrl(document.url) }}
    </p>
  </article>
</template>