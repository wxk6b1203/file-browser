<template>
  <FrontPage @new-connection="workspace.openNewConnection" @open-settings="workspace.openSettings" />
</template>

<script setup lang="ts">
import FrontPage from '@/views/Blank'
import { useWorkspaceStore } from '@/stores/workspace'

const workspace = useWorkspaceStore()
</script>
