<script setup lang="ts">
interface Props {
  sidebarItems?: Array<{ label: string; path: string; icon?: string }>
}

withDefaults(defineProps<Props>(), {
  sidebarItems: () => [
    { label: 'Dashboard', path: '/' },
    { label: 'Documents', path: '/documents' },
    { label: 'Sources', path: '/sources' },
    { label: 'Users', path: '/users' },
    { label: 'Settings', path: '/settings' },
  ],
})
</script>

<template>
  <div class="flex h-screen bg-gray-50">
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <h1 class="text-xl font-semibold text-gray-900">Phosboard</h1>
      </div>
      <nav class="flex-1 p-4">
        <ul class="space-y-1">
          <li v-for="item in sidebarItems" :key="item.path">
            <a
              :href="item.path"
              class="block px-3 py-2 rounded-md text-gray-700 hover:bg-gray-100 transition-colors"
            >
              {{ item.label }}
            </a>
          </li>
        </ul>
      </nav>
    </aside>
    <main class="flex-1 overflow-auto">
      <slot />
    </main>
  </div>
</template>