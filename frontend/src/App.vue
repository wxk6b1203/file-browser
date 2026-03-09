<template>
  <el-config-provider>
  <RouterView />
  </el-config-provider>
</template>
<script setup lang="ts">
import { RouterView } from 'vue-router'
import { ElConfigProvider } from 'element-plus'
</script>
<style scoped></style>
