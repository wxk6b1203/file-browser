<template>
  <el-config-provider>
  <RouterView />
  </el-config-provider>
</template>
<script setup lang="ts">
import { RouterView } from 'vue-router'
import { ElConfigProvider } from 'element-plus'
import { defineShortcuts, DEFAULT_SHORTCUTS } from '@/composables/useShortcut'

// Initialize global keyboard shortcut listener once at app startup.
defineShortcuts(DEFAULT_SHORTCUTS)
</script>
<style scoped></style>
