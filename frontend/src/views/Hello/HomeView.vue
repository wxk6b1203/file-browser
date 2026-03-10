<template>
  <div style="width: 100%;">
    <el-input-number v-model="num" :min="1" :max="10" />
    <SplitPane style="height: 100vh;" :gap="24">
      <template v-for="i in num" :key="i">
        <SplitPanePanel :borderRadius="'24px'">
          <div style="background-color: #eee; height: 100%; display: flex; align-items: center; justify-content: center;">
            {{ t('hello') }} {{ i }}
          </div>
        </SplitPanePanel>

      </template>
    </SplitPane>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import { SplitPane, SplitPanePanel } from '@/components/SplitPane';
import { ref } from 'vue';
const num = ref(1)
</script>

<style scoped></style>