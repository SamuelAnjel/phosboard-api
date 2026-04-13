<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

interface NavItem {
  title: string
  icon: string
  to: string
}

const router = useRouter()
const authStore = useAuthStore()
const drawer = ref(true)

const navItems: NavItem[] = [
  { title: 'Dashboard', icon: 'mdi-view-dashboard', to: '/' },
  { title: 'Documentos', icon: 'mdi-file-document-multiple', to: '/documents' },
  { title: 'Conceptos Semilla', icon: 'mdi-tag-multiple', to: '/concepts' },
  { title: 'Fuentes RSS', icon: 'mdi-rss', to: '/sources' },
  { title: 'Ajustes', icon: 'mdi-cog', to: '/settings' },
]

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <v-navigation-drawer v-model="drawer" app>
    <v-list-item class="pa-4">
      <v-list-item-title class="text-h6">Phosboard</v-list-item-title>
      <v-list-item-subtitle>Document Intelligence</v-list-item-subtitle>
    </v-list-item>

    <v-divider></v-divider>

    <v-list density="compact" nav>
      <v-list-item
        v-for="item in navItems"
        :key="item.title"
        :to="item.to"
        :prepend-icon="item.icon"
        :title="item.title"
        link
      ></v-list-item>
    </v-list>
  </v-navigation-drawer>

  <v-app-bar app elevation="1">
    <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>

    <v-toolbar-title>Phosboard</v-toolbar-title>

    <v-spacer></v-spacer>

    <v-btn icon @click="handleLogout">
      <v-icon>mdi-logout</v-icon>
    </v-btn>
  </v-app-bar>

  <v-main>
    <v-container fluid>
      <slot></slot>
    </v-container>
  </v-main>
</template>
