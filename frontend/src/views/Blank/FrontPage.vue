<template>
  <div class="front-page">
    <div class="content">
      <!-- Logo/Icon 区域 -->
      <div class="logo-section">
        <div class="logo-bg">
          <i-mdi-folder-open-outline class="logo-icon" />
        </div>
      </div>

      <!-- 欢迎文字 -->
      <h1 class="welcome-text">{{ t('welcome.title') }}</h1>
      <p class="subtitle">{{ t('welcome.subtitle') }}</p>

      <!-- 快捷键提示区域 -->
      <div class="shortcuts-section">
        <div
          v-for="(shortcut, index) in shortcuts"
          :key="index"
          class="shortcut-item"
          :class="{ 'with-divider': index > 0 }"
        >
          <div class="shortcut-left">
            <i-mdi-keyboard class="shortcut-icon" />
            <span class="shortcut-desc">{{ t(shortcut.desc) }}</span>
          </div>
          <div class="shortcut-keys">
            <template v-for="(key, kIndex) in shortcut.keys" :key="kIndex">
              <kbd v-if="key !== '+'" class="key">{{ formatKey(key) }}</kbd>
              <span v-else class="key-separator">+</span>
            </template>
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="actions-section">
        <el-button type="primary" size="large" class="cta-button" @click="onNewConnection">
          <template #icon>
            <i-ep-plus />
          </template>
          {{ t('actions.newConnection') }}
        </el-button>
        <el-button size="large" text @click="onOpenSettings">
          <template #icon>
            <i-ep-setting />
          </template>
          {{ t('actions.settings') }}
        </el-button>
      </div>
    </div>

    <!-- 底部提示 -->
    <div class="footer-hint">
      <i-ep-info-filled class="hint-icon" />
      <span>{{ t('welcome.footerHint') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const emit = defineEmits<{
  newConnection: []
  openSettings: []
}>()

// 检测是否为 Mac 系统
const isMac = computed(() => {
  if (typeof navigator === 'undefined') return false
  return navigator.platform.toLowerCase().includes('mac')
})

// 快捷键列表
const shortcuts = [
  { keys: ['CtrlOrCmd', 'Shift', 'N'], desc: 'shortcuts.newConnection' },
  { keys: ['CtrlOrCmd', '.'], desc: 'shortcuts.settings' },
]

// 格式化按键显示
function formatKey(key: string): string {
  if (key === 'CtrlOrCmd') return isMac.value ? '⌘' : 'Ctrl'
  return key
}

// 新建连接
function onNewConnection() {
  emit('newConnection')
}

// 打开设置
function onOpenSettings() {
  emit('openSettings')
}
</script>

<style lang="scss" scoped>
.front-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  min-height: 100%;
  padding: 40px 20px;
  background-color: var(--el-bg-color);
  color: var(--el-text-color-primary);
  position: relative;
  box-sizing: border-box;
}

.content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 520px;
  position: relative;
  z-index: 1;
}

// Logo 区域
.logo-section {
  margin-bottom: 28px;

  .logo-bg {
    width: 96px;
    height: 96px;
    border-radius: 24px;
    background: linear-gradient(135deg, var(--el-color-primary-light-8), var(--el-color-primary-light-9));
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

    .logo-icon {
      font-size: 48px;
      color: var(--el-color-primary);
    }
  }
}

// 欢迎文字
.welcome-text {
  font-size: 26px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: var(--el-text-color-primary);
}

.subtitle {
  font-size: 15px;
  color: var(--el-text-color-secondary);
  margin: 0 0 40px 0;
}

// 快捷键区域
.shortcuts-section {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin-bottom: 40px;
  width: 100%;
  background: var(--el-fill-color-light);
  border-radius: 12px;
  padding: 8px 0;
  border: 1px solid var(--el-border-color-lighter);
}

.shortcut-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  font-size: 14px;
  padding: 10px 20px;
  transition: background-color 150ms ease;

  &:hover {
    background-color: var(--el-fill-color);
  }

  &.with-divider {
    border-top: 1px solid var(--el-border-color-lighter);
  }
}

.shortcut-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.shortcut-icon {
  font-size: 16px;
  color: var(--el-text-color-secondary);
  opacity: 0.7;
}

.shortcut-desc {
  color: var(--el-text-color-regular);
  font-size: 14px;
}

.shortcut-keys {
  display: flex;
  align-items: center;
  gap: 4px;
}

.key {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 26px;
  height: 26px;
  padding: 0 7px;
  font-family: var(--el-font-family);
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  box-shadow: 0 1px 0 var(--el-border-color-darker);
}

.key-separator {
  color: var(--el-text-color-placeholder);
  font-size: 11px;
  padding: 0 2px;
}

// 操作按钮区域
.actions-section {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;

  .cta-button {
    min-width: 140px;
  }
}

// 底部提示
.footer-hint {
  position: absolute;
  bottom: 24px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--el-text-color-secondary);

  .hint-icon {
    font-size: 14px;
    color: var(--el-color-info);
  }
}

// 响应式适配
@media (max-width: 480px) {
  .front-page {
    padding: 32px 16px;
  }

  .welcome-text {
    font-size: 22px;
  }

  .subtitle {
    font-size: 14px;
    margin-bottom: 32px;
  }

  .shortcuts-section {
    margin-bottom: 32px;
  }

  .shortcut-item {
    padding: 10px 16px;
  }

  .actions-section {
    flex-direction: column;
    width: 100%;

    :deep(.el-button) {
      width: 100%;
      justify-content: center;
    }
  }
}

// 深色模式适配
:global(.dark) {
  .key {
    background: var(--el-fill-color-dark);
    box-shadow: 0 1px 0 rgba(0, 0, 0, 0.2);
  }
}
</style>
