<script setup lang="ts">
import { RouterView } from 'vue-router'
import { ElConfigProvider } from 'element-plus'

</script>

<template>
  <el-config-provider>

  <RouterView />
  </el-config-provider>
</template>

<style scoped></style>
