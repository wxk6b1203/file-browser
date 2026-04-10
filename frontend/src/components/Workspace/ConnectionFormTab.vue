<template>
  <div class="connection-form">
    <div class="connection-form__header">
      <div>
        <h2 class="connection-form__title">{{ title }}</h2>
        <p class="connection-form__subtitle">{{ subtitle }}</p>
      </div>
      <div class="connection-form__actions">
        <el-button :loading="testing" @click="testCurrentConnection">{{ t('workspace.connectionForm.test') }}</el-button>
        <el-button @click="resetForm">{{ t('workspace.connectionForm.reset') }}</el-button>
        <el-button type="primary" :loading="saving" @click="submit">
          {{ t('workspace.connectionForm.save') }}
        </el-button>
      </div>
    </div>

    <el-form label-position="top" class="connection-form__grid">
      <section class="connection-form__card">
        <h3 class="connection-form__section-title">{{ t('workspace.connectionForm.basic') }}</h3>

        <el-form-item :label="t('workspace.connectionForm.name')" :error="validationErrors.name">
          <el-input
            v-model="form.name"
            :placeholder="t('workspace.connectionForm.namePlaceholder')"
            @input="onFieldChanged('name')"
          />
        </el-form-item>

        <el-form-item :label="t('workspace.connectionForm.driver')" :error="validationErrors.driver">
          <el-select v-model="form.driver" :disabled="props.mode === 'edit'">
            <el-option
              v-for="driver in connections.drivers"
              :key="driver.name"
              :label="driver.name"
              :value="driver.name"
            />
          </el-select>
        </el-form-item>

        <el-form-item v-if="showLogicalRootField" :label="t('workspace.connectionForm.root')">
          <el-input
            v-model="form.root"
            :placeholder="t('workspace.connectionForm.rootPlaceholder')"
            @input="onFieldChanged()"
          />
        </el-form-item>

        <el-form-item :label="t('workspace.connectionForm.description')">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            :placeholder="t('workspace.connectionForm.descriptionPlaceholder')"
            @input="onFieldChanged()"
          />
        </el-form-item>

        <div class="connection-form__toggles">
          <label class="connection-form__toggle">
            <span>{{ t('workspace.connectionForm.enabled') }}</span>
            <el-switch v-model="form.enabled" @change="onFieldChanged()" />
          </label>
          <label class="connection-form__toggle">
            <span>{{ t('workspace.connectionForm.readOnly') }}</span>
            <el-switch v-model="form.readOnly" @change="onFieldChanged()" />
          </label>
        </div>
      </section>

      <section class="connection-form__card">
        <h3 class="connection-form__section-title">{{ t('workspace.connectionForm.driverConfig') }}</h3>

        <el-form-item
          v-for="field in currentDriverFields"
          :key="field.key"
          :label="field.label"
          :required="field.required"
          :error="validationErrors[`config.${field.key}`]"
        >
          <el-input
            v-if="field.kind === 'text' || field.kind === 'password'"
            v-model="form.config[field.key]"
            :type="field.kind"
            :show-password="field.kind === 'password'"
            :placeholder="field.placeholder"
            @input="onFieldChanged(`config.${field.key}`)"
          />
          <el-input
            v-else-if="field.kind === 'textarea'"
            v-model="form.config[field.key]"
            type="textarea"
            :rows="4"
            :placeholder="field.placeholder"
            @input="onFieldChanged(`config.${field.key}`)"
          />
          <el-input-number
            v-else-if="field.kind === 'number'"
            v-model="form.config[field.key]"
            :min="field.min ?? 0"
            controls-position="right"
            @change="onFieldChanged(`config.${field.key}`)"
          />
          <el-switch v-else v-model="form.config[field.key]" @change="onFieldChanged(`config.${field.key}`)" />
        </el-form-item>
      </section>

      <section v-if="testResult || testError" class="connection-form__card connection-form__card--wide">
        <div class="connection-form__result-head">
          <div>
            <h3 class="connection-form__section-title">{{ t('workspace.connectionForm.testResultTitle') }}</h3>
            <p class="connection-form__result-subtitle">
              {{ testResult ? t('workspace.connectionForm.testResultSuccess') : t('workspace.connectionForm.testResultFailed') }}
            </p>
          </div>
          <span
            class="connection-form__result-badge"
            :class="testResult ? 'connection-form__result-badge--success' : 'connection-form__result-badge--error'"
          >
            {{ testResult ? t('workspace.connectionForm.resultConnected') : t('workspace.connectionForm.resultError') }}
          </span>
        </div>

        <div v-if="testResult" class="connection-form__result-grid">
          <div class="connection-form__result-item">
            <span class="connection-form__result-label">{{ t('workspace.connectionForm.resultDriver') }}</span>
            <strong>{{ testResult.driver }}</strong>
          </div>
          <div class="connection-form__result-item">
            <span class="connection-form__result-label">{{ t('workspace.connectionForm.resultCapabilities') }}</span>
            <strong>{{ capabilityItems.length }}</strong>
          </div>
        </div>

        <div v-if="testResult && capabilityItems.length > 0" class="connection-form__capabilities">
          <el-tag
            v-for="item in capabilityItems"
            :key="item.key"
            size="small"
            type="success"
            effect="plain"
          >
            {{ item.label }}
          </el-tag>
        </div>

        <div v-if="testError" class="connection-form__error-box">
          {{ testError }}
        </div>
      </section>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { connection } from '../../../wailsjs/go/models'
import { emitConnectionConfigRefresh } from '@/composables/useConnectionConfigRefresh'
import { useConnectionsStore } from '@/stores/connections'
import { useWorkspaceStore } from '@/stores/workspace'
import { createDriverConfig, DRIVER_FIELD_SCHEMAS } from './connectionSchemas'

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  connectionId?: string
}>(), {
  mode: 'create',
  connectionId: undefined,
})

const { t } = useI18n()
const connections = useConnectionsStore()
const workspace = useWorkspaceStore()

const saving = ref(false)
const testing = ref(false)
const hydratedEditModel = ref(false)
const suppressDriverWatch = ref(false)
const validationErrors = ref<Record<string, string>>({})
const testResult = ref<connection.State | null>(null)
const testError = ref('')

const form = reactive({
  id: '',
  name: '',
  driver: 'Local',
  description: '',
  enabled: true,
  readOnly: false,
  root: '',
  tags: [] as string[],
  metadata: {} as Record<string, string>,
  config: createDriverConfig('Local') as Record<string, any>,
})

const title = computed(() => props.mode === 'edit'
  ? t('workspace.connectionForm.editTitle')
  : t('workspace.connectionForm.createTitle'))
const subtitle = computed(() => props.mode === 'edit'
  ? t('workspace.connectionForm.editSubtitle')
  : t('workspace.connectionForm.createSubtitle'))
const currentDriverFields = computed(() => DRIVER_FIELD_SCHEMAS[form.driver] ?? [])
const showLogicalRootField = computed(() => scopedPathConfigKeyForDriver(form.driver) === null)
const capabilityItems = computed(() => {
  const caps = testResult.value?.capabilities
  if (!caps) return []

  const items = [
    { key: 'CanList', label: t('workspace.connectionForm.capabilityList'), enabled: caps.CanList },
    { key: 'CanRead', label: t('workspace.connectionForm.capabilityRead'), enabled: caps.CanRead },
    { key: 'CanWrite', label: t('workspace.connectionForm.capabilityWrite'), enabled: caps.CanWrite },
    { key: 'CanDelete', label: t('workspace.connectionForm.capabilityDelete'), enabled: caps.CanDelete },
    { key: 'CanCopy', label: t('workspace.connectionForm.capabilityCopy'), enabled: caps.CanCopy },
    { key: 'CanMove', label: t('workspace.connectionForm.capabilityMove'), enabled: caps.CanMove },
    { key: 'CanRename', label: t('workspace.connectionForm.capabilityRename'), enabled: caps.CanRename },
    { key: 'CanMkdir', label: t('workspace.connectionForm.capabilityMkdir'), enabled: caps.CanMkdir },
    { key: 'CanPresign', label: t('workspace.connectionForm.capabilityPresign'), enabled: caps.CanPresign },
    { key: 'CanTransfer', label: t('workspace.connectionForm.capabilityTransfer'), enabled: caps.CanTransfer },
    { key: 'AtomicMove', label: t('workspace.connectionForm.capabilityAtomicMove'), enabled: caps.AtomicMove },
    { key: 'SupportsVersion', label: t('workspace.connectionForm.capabilityVersion'), enabled: caps.SupportsVersion },
  ]

  return items.filter((item) => item.enabled)
})

function scopedPathConfigKeyForDriver(driver: string) {
  if (driver === 'Local' || driver === 'SFTP' || driver === 'WebDAV') {
    return 'rootPath'
  }
  if (driver === 'S3' || driver === 'OSS') {
    return 'prefix'
  }
  return null
}

function normalizeScopedPathFields(driver: string, root: string, config: Record<string, any>) {
  const configKey = scopedPathConfigKeyForDriver(driver)
  const nextRoot = root.trim()
  if (!configKey) {
    return {
      root: nextRoot,
      config,
    }
  }

  const configValue = String(config[configKey] ?? '').trim()
  const nextValue = configValue || nextRoot

  return {
    root: '',
    config: {
      ...config,
      [configKey]: nextValue,
    },
  }
}

function applyDefinition(def?: ReturnType<typeof connections.getDefinition> | null) {
  suppressDriverWatch.value = true
  if (!def) {
    form.id = ''
    form.name = ''
    form.driver = 'Local'
    form.description = ''
    form.enabled = true
    form.readOnly = false
    form.root = ''
    form.tags = []
    form.metadata = {}
    form.config = createDriverConfig('Local')
    clearValidation()
    clearTestState()
    nextTick(() => {
      suppressDriverWatch.value = false
    })
    return
  }

  form.id = def.id
  form.name = def.name
  form.driver = def.driver
  form.description = def.description ?? ''
  form.enabled = def.enabled
  form.readOnly = def.readOnly ?? false
  form.tags = def.tags ?? []
  form.metadata = def.metadata ?? {}
  const next = normalizeScopedPathFields(def.driver, def.root ?? '', {
    ...createDriverConfig(def.driver),
    ...(def.config ?? {}),
  })
  form.root = next.root
  form.config = next.config
  clearValidation()
  clearTestState()
  nextTick(() => {
    suppressDriverWatch.value = false
  })
}

function resetForm() {
  applyDefinition(props.mode === 'edit' ? connections.getDefinition(props.connectionId ?? '') : null)
}

async function submit() {
  const errors = validateForm()
  if (Object.keys(errors).length > 0) {
    ElMessage.error(t('workspace.connectionForm.validationFailed'))
    return
  }

  saving.value = true
  try {
    const saved = await connections.saveConnection(buildDefinition())
    if (props.mode === 'edit') {
      form.id = saved.id
    }
    workspace.resetConnectionBrowserState(saved.id, '')
    emitConnectionConfigRefresh({
      connectionId: saved.id,
      resetToRoot: true,
    })
    workspace.openConnection(saved.id, saved.name, '')
    if (props.mode === 'create') {
      applyDefinition(null)
      workspace.closeTabById('connection-form:new')
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    saving.value = false
  }
}

function buildDefinition() {
  const next = normalizeScopedPathFields(form.driver, form.root, { ...form.config })
  return {
    ...form,
    root: next.root,
    config: next.config,
    metadata: { ...form.metadata },
    tags: [...form.tags],
  }
}

function clearValidation() {
  validationErrors.value = {}
}

function clearTestState() {
  testResult.value = null
  testError.value = ''
}

function onFieldChanged(errorKey?: string) {
  if (errorKey && validationErrors.value[errorKey]) {
    const next = { ...validationErrors.value }
    delete next[errorKey]
    validationErrors.value = next
  }
  clearTestState()
}

function onDriverChanged() {
  onFieldChanged('driver')
  const next = normalizeScopedPathFields(form.driver, form.root, createDriverConfig(form.driver))
  form.root = next.root
  form.config = next.config
  clearValidation()
}

function validateForm() {
  const nextErrors: Record<string, string> = {}

  if (!String(form.name ?? '').trim()) {
    nextErrors.name = t('workspace.connectionForm.validationNameRequired')
  }
  if (!String(form.driver ?? '').trim()) {
    nextErrors.driver = t('workspace.connectionForm.validationDriverRequired')
  }

  for (const field of currentDriverFields.value) {
    if (!field.required) continue
    const value = form.config[field.key]
    if (typeof value === 'number') {
      if (!Number.isFinite(value)) {
        nextErrors[`config.${field.key}`] = t('workspace.connectionForm.validationFieldRequired', { field: field.label })
      }
      continue
    }
    if (!String(value ?? '').trim()) {
      nextErrors[`config.${field.key}`] = t('workspace.connectionForm.validationFieldRequired', { field: field.label })
    }
  }

  if (form.driver === 'SFTP') {
    const password = String(form.config.password ?? '').trim()
    const privateKey = String(form.config.privateKey ?? '').trim()
    const privateKeyPath = String(form.config.privateKeyPath ?? '').trim()
    if (!password && !privateKey && !privateKeyPath) {
      nextErrors['config.password'] = t('workspace.connectionForm.validationSftpAuthRequired')
      nextErrors['config.privateKey'] = t('workspace.connectionForm.validationSftpAuthRequired')
      nextErrors['config.privateKeyPath'] = t('workspace.connectionForm.validationSftpAuthRequired')
    }
  }

  validationErrors.value = nextErrors
  return nextErrors
}

async function testCurrentConnection() {
  const errors = validateForm()
  if (Object.keys(errors).length > 0) {
    ElMessage.error(t('workspace.connectionForm.validationFailed'))
    return
  }

  testing.value = true
  clearTestState()
  try {
    const state = await connections.testConnection(buildDefinition())
    testResult.value = state

    ElMessage.success(t('workspace.connectionForm.testSuccess', {
      driver: state?.driver || form.driver,
      count: capabilityItems.value.length,
    }))
  } catch (error) {
    testError.value = error instanceof Error ? error.message : String(error)
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    testing.value = false
  }
}

watch(
  () => form.driver,
  (driver, prev) => {
    if (driver === prev) return
    if (suppressDriverWatch.value) return
    if (props.mode === 'edit' && !hydratedEditModel.value) return
    onDriverChanged()
  },
)

connections.hydrate().then(() => {
  if (props.mode === 'edit' && props.connectionId) {
    const def = connections.getDefinition(props.connectionId)
    if (def) {
      applyDefinition(def)
    }
    hydratedEditModel.value = true
  }
})
</script>

<style scoped>
.connection-form {
  height: 100%;
  padding: 24px;
  overflow: auto;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--theme-color-primary) 14%, transparent), transparent 28%),
    linear-gradient(180deg, color-mix(in srgb, var(--theme-color-bg-overlay) 70%, transparent), transparent 45%);
}

.connection-form__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.connection-form__title {
  margin: 0;
  font-size: 24px;
  line-height: 1.1;
  color: var(--theme-color-text-base);
}

.connection-form__subtitle {
  margin: 8px 0 0;
  color: var(--theme-color-text-secondary);
}

.connection-form__actions {
  display: flex;
  gap: 10px;
}

.connection-form__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 18px;
}

.connection-form__card {
  padding: 18px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 16px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
  box-shadow: var(--theme-shadow-sm);
}

.connection-form__card--wide {
  grid-column: 1 / -1;
}

.connection-form__section-title {
  margin: 0 0 16px;
  font-size: 15px;
  font-weight: 700;
  color: var(--theme-color-text-base);
  letter-spacing: 0.02em;
}

.connection-form__toggles {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.connection-form__toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 180px;
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--theme-color-bg-overlay);
  color: var(--theme-color-text);
}

.connection-form__result-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.connection-form__result-subtitle {
  margin: -6px 0 0;
  color: var(--theme-color-text-secondary);
}

.connection-form__result-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 88px;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.connection-form__result-badge--success {
  background: color-mix(in srgb, var(--theme-color-success) 14%, transparent);
  color: var(--theme-color-success);
}

.connection-form__result-badge--error {
  background: color-mix(in srgb, var(--theme-color-danger) 14%, transparent);
  color: var(--theme-color-danger);
}

.connection-form__result-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
  margin-top: 14px;
}

.connection-form__result-item {
  display: grid;
  gap: 6px;
  padding: 12px;
  border-radius: 12px;
  background: var(--theme-color-bg-overlay);
}

.connection-form__result-label {
  font-size: 12px;
  color: var(--theme-color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.connection-form__capabilities {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}

.connection-form__error-box {
  margin-top: 14px;
  padding: 12px 14px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--theme-color-danger) 10%, transparent);
  color: var(--theme-color-text-base);
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 900px) {
  .connection-form {
    padding: 16px;
  }

  .connection-form__header {
    flex-direction: column;
  }
}
</style>
