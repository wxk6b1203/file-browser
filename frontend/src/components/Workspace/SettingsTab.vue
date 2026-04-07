<template>
  <div class="settings-tab">
    <div class="settings-tab__hero">
      <div>
        <p class="settings-tab__eyebrow">{{ t('workspace.settings.eyebrow') }}</p>
        <h2 class="settings-tab__title">{{ t('workspace.settings.title') }}</h2>
        <p class="settings-tab__subtitle">{{ t('workspace.settings.subtitle') }}</p>
      </div>

      <div class="settings-tab__actions">
        <el-button :disabled="settings.loading || settings.saving" @click="resetForm">
          {{ t('workspace.settings.reset') }}
        </el-button>
        <el-button type="primary" :loading="settings.saving" @click="save">
          {{ t('workspace.settings.save') }}
        </el-button>
      </div>
    </div>

    <div v-if="settings.loading && !settings.ready" class="settings-tab__loading">
      {{ t('workspace.settings.loading') }}
    </div>

    <div v-else class="settings-tab__grid">
      <section class="settings-tab__card">
        <h3>{{ t('workspace.settings.generalTitle') }}</h3>
        <p>{{ t('workspace.settings.generalDesc') }}</p>

        <el-form label-position="top">
          <el-form-item :label="t('workspace.settings.localeLabel')">
            <el-select v-model="form.locale">
              <el-option value="zh" :label="t('workspace.settings.localeZh')" />
              <el-option value="en" :label="t('workspace.settings.localeEn')" />
              <el-option value="ja" :label="t('workspace.settings.localeJa')" />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('workspace.settings.themeLabel')">
            <el-select v-model="form.theme">
              <el-option value="system" :label="t('workspace.settings.themeSystem')" />
              <el-option
                v-for="item in theme.themes"
                :key="item.id"
                :value="item.id"
                :label="item.label"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('workspace.settings.explorerFontSizeLabel')">
            <div class="settings-tab__inline-field">
              <el-input-number
                v-model="form.explorerFontSize"
                :min="11"
                :max="18"
                controls-position="right"
              />
              <el-button @click="resetExplorerFontSize">{{ t('workspace.settings.restoreDefault') }}</el-button>
            </div>
          </el-form-item>

          <el-form-item :label="t('workspace.settings.fileListFontSizeLabel')">
            <div class="settings-tab__inline-field">
              <el-input-number
                v-model="form.fileListFontSize"
                :min="11"
                :max="18"
                controls-position="right"
              />
              <el-button @click="resetFileListFontSize">{{ t('workspace.settings.restoreDefault') }}</el-button>
            </div>
          </el-form-item>

          <el-form-item :label="t('workspace.settings.tempDirLabel')">
            <el-input v-model="form.appTempDir" :placeholder="t('workspace.settings.tempDirPlaceholder')" />
          </el-form-item>
        </el-form>
      </section>

      <section class="settings-tab__card">
        <h3>{{ t('workspace.settings.searchTitle') }}</h3>
        <p>{{ t('workspace.settings.searchDesc') }}</p>

        <el-form label-position="top">
          <el-form-item :label="t('workspace.settings.searchConcurrencyLabel')">
            <el-input-number v-model="form.searchMaxConcurrency" :min="1" controls-position="right" />
          </el-form-item>

          <el-form-item :label="t('workspace.settings.searchLimitLabel')">
            <el-input-number v-model="form.searchResultLimit" :min="1" controls-position="right" />
          </el-form-item>
        </el-form>
      </section>

      <section class="settings-tab__card">
        <h3>{{ t('workspace.settings.transferTitle') }}</h3>
        <p>{{ t('workspace.settings.transferDesc') }}</p>

        <el-form label-position="top">
          <el-form-item :label="t('workspace.settings.transferTempDirLabel')">
            <el-input v-model="form.transferTempDir" :placeholder="t('workspace.settings.transferTempDirPlaceholder')" />
          </el-form-item>

          <el-form-item :label="t('workspace.settings.transferDownloadDirLabel')">
            <div class="settings-tab__inline-field">
              <el-input v-model="form.transferDownloadDir" :placeholder="t('workspace.settings.transferDownloadDirPlaceholder')" />
              <el-button @click="pickDownloadDirectory">{{ t('workspace.settings.browse') }}</el-button>
            </div>
          </el-form-item>

          <el-form-item :label="t('workspace.settings.transferOverwriteLabel')">
            <el-select v-model="form.transferOverwriteStrategy">
              <el-option value="rename" :label="t('workspace.settings.transferOverwriteRename')" />
              <el-option value="overwrite" :label="t('workspace.settings.transferOverwriteOverwrite')" />
            </el-select>
          </el-form-item>
        </el-form>
      </section>

      <section class="settings-tab__card">
        <h3>{{ t('workspace.settings.logTitle') }}</h3>
        <p>{{ t('workspace.settings.logDesc') }}</p>

        <el-form label-position="top">
          <el-form-item :label="t('workspace.settings.logLevelLabel')">
            <el-select v-model="form.logLevel">
              <el-option value="debug" label="debug" />
              <el-option value="info" label="info" />
              <el-option value="warn" label="warn" />
              <el-option value="error" label="error" />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('workspace.settings.logOutputsLabel')">
            <el-input
              v-model="form.logOutputsText"
              type="textarea"
              :rows="4"
              :placeholder="t('workspace.settings.logOutputsPlaceholder')"
            />
          </el-form-item>
        </el-form>
      </section>

      <section class="settings-tab__card settings-tab__card--wide">
        <h3>{{ t('workspace.settings.pathsTitle') }}</h3>
        <p>{{ t('workspace.settings.pathsDesc') }}</p>

        <div class="settings-tab__path-list">
          <div class="settings-tab__path-item">
            <span class="settings-tab__path-label">{{ t('workspace.settings.connectionsFileLabel') }}</span>
            <code>{{ settings.appConfig?.paths?.connectionsFile || '--' }}</code>
          </div>
          <div class="settings-tab__path-item">
            <span class="settings-tab__path-label">{{ t('workspace.settings.stateFileLabel') }}</span>
            <code>{{ settings.appConfig?.paths?.stateFile || '--' }}</code>
          </div>
        </div>
      </section>

      <section class="settings-tab__card settings-tab__card--wide">
        <h3>{{ t('workspace.settings.noteTitle') }}</h3>
        <p>{{ t('workspace.settings.noteDesc') }}</p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { PickDownloadDirectory } from '../../../wailsjs/go/main/App'
import { config } from '../../../wailsjs/go/models'
import { useSettingsStore } from '@/stores/settings'
import { useThemeStore } from '@/stores/theme'

interface SettingsFormState {
  locale: string
  theme: string
  explorerFontSize: number
  fileListFontSize: number
  appTempDir: string
  searchMaxConcurrency: number
  searchResultLimit: number
  transferTempDir: string
  transferDownloadDir: string
  transferOverwriteStrategy: string
  logLevel: string
  logOutputsText: string
}

const { t } = useI18n()
const settings = useSettingsStore()
const theme = useThemeStore()
const DEFAULT_UI_FONT_SIZE = 13

const form = reactive<SettingsFormState>({
  locale: 'zh',
  theme: 'system',
  explorerFontSize: DEFAULT_UI_FONT_SIZE,
  fileListFontSize: DEFAULT_UI_FONT_SIZE,
  appTempDir: '',
  searchMaxConcurrency: 4,
  searchResultLimit: 500,
  transferTempDir: '',
  transferDownloadDir: '',
  transferOverwriteStrategy: 'rename',
  logLevel: 'info',
  logOutputsText: 'stdout',
})

function cloneAppConfig(value: config.AppConfig | null | undefined) {
  if (!value) return null
  return JSON.parse(JSON.stringify(value)) as config.AppConfig
}

function applyConfigToForm(next: config.AppConfig | null | undefined) {
  if (!next) return

  form.locale = next.ui?.locale || next.app?.locale || 'zh'
  form.theme = next.ui?.theme || next.app?.theme || 'system'
  form.explorerFontSize = next.ui?.explorerFontSize || DEFAULT_UI_FONT_SIZE
  form.fileListFontSize = next.ui?.fileListFontSize || DEFAULT_UI_FONT_SIZE
  form.appTempDir = next.app?.tempDir || ''
  form.searchMaxConcurrency = next.search?.maxConcurrency || 4
  form.searchResultLimit = next.search?.resultLimit || 500
  form.transferTempDir = next.transfer?.tempDir || ''
  form.transferDownloadDir = next.transfer?.downloadDir || ''
  form.transferOverwriteStrategy = next.transfer?.overwriteStrategy || 'rename'
  form.logLevel = next.log?.level || 'info'
  form.logOutputsText = (next.log?.outputs || []).join('\n') || 'stdout'
}

function buildConfigFromForm() {
  const base = cloneAppConfig(settings.appConfig)
  if (!base) {
    return null
  }

  base.app.locale = form.locale
  base.app.theme = form.theme
  base.app.tempDir = form.appTempDir.trim()

  base.ui.locale = form.locale
  base.ui.theme = form.theme
  base.ui.explorerFontSize = Math.max(11, Math.min(18, Number(form.explorerFontSize) || DEFAULT_UI_FONT_SIZE))
  base.ui.fileListFontSize = Math.max(11, Math.min(18, Number(form.fileListFontSize) || DEFAULT_UI_FONT_SIZE))

  base.search.maxConcurrency = Math.max(1, Number(form.searchMaxConcurrency) || 1)
  base.search.resultLimit = Math.max(1, Number(form.searchResultLimit) || 1)

  base.transfer.tempDir = form.transferTempDir.trim()
  base.transfer.downloadDir = form.transferDownloadDir.trim()
  base.transfer.overwriteStrategy = form.transferOverwriteStrategy

  base.log.level = form.logLevel
  base.log.outputs = form.logOutputsText
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean)

  return base
}

function resetForm() {
  applyConfigToForm(settings.appConfig)
}

function resetExplorerFontSize() {
  form.explorerFontSize = DEFAULT_UI_FONT_SIZE
}

function resetFileListFontSize() {
  form.fileListFontSize = DEFAULT_UI_FONT_SIZE
}

async function pickDownloadDirectory() {
  try {
    const selected = await PickDownloadDirectory()
    if (!selected) return
    form.transferDownloadDir = selected
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function save() {
  const next = buildConfigFromForm()
  if (!next) {
    ElMessage.error(t('workspace.settings.loadFailed'))
    return
  }

  try {
    const saved = await settings.saveConfig(next)
    applyConfigToForm(saved)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

watch(
  () => settings.appConfig,
  (next) => {
    if (!next) return
    applyConfigToForm(next)
  },
  { immediate: true },
)

void settings.hydrate()
</script>

<style scoped>
.settings-tab {
  height: 100%;
  padding: 24px;
  overflow: auto;
  background:
    linear-gradient(130deg, color-mix(in srgb, var(--theme-color-primary) 14%, transparent), transparent 34%),
    var(--theme-color-bg-base);
}

.settings-tab__hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.settings-tab__eyebrow {
  margin: 0 0 8px;
  color: var(--theme-color-primary);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
}

.settings-tab__title {
  margin: 0;
  color: var(--theme-color-text-base);
  font-size: 28px;
}

.settings-tab__subtitle {
  margin: 8px 0 0;
  color: var(--theme-color-text-secondary);
  max-width: 720px;
}

.settings-tab__actions {
  display: flex;
  gap: 10px;
}

.settings-tab__loading {
  padding: 24px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 16px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
  color: var(--theme-color-text-secondary);
}

.settings-tab__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 18px;
}

.settings-tab__card {
  padding: 18px;
  border-radius: 16px;
  border: 1px solid var(--theme-color-border-light);
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
  box-shadow: var(--theme-shadow-sm);
}

.settings-tab__card--wide {
  grid-column: 1 / -1;
}

.settings-tab__card h3 {
  margin: 0 0 8px;
  color: var(--theme-color-text-base);
}

.settings-tab__card p {
  margin: 0 0 14px;
  color: var(--theme-color-text-secondary);
}

.settings-tab__path-list {
  display: grid;
  gap: 12px;
}

.settings-tab__path-item {
  display: grid;
  gap: 6px;
}

.settings-tab__path-label {
  font-size: 12px;
  color: var(--theme-color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.settings-tab__path-item code {
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--theme-color-bg-overlay);
  color: var(--theme-color-text);
  word-break: break-all;
}

.settings-tab__inline-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
  width: 100%;
}

@media (max-width: 900px) {
  .settings-tab {
    padding: 16px;
  }

  .settings-tab__hero {
    flex-direction: column;
  }

  .settings-tab__actions {
    width: 100%;
  }

  .settings-tab__actions :deep(.el-button) {
    flex: 1;
  }
}
</style>
