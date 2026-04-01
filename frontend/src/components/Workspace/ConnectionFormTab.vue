<template>
  <div class="connection-form">
    <div class="connection-form__header">
      <div>
        <h2 class="connection-form__title">{{ title }}</h2>
        <p class="connection-form__subtitle">{{ subtitle }}</p>
      </div>
      <div class="connection-form__actions">
        <el-button @click="resetForm">{{ t('workspace.connectionForm.reset') }}</el-button>
        <el-button type="primary" :loading="saving" @click="submit">
          {{ t('workspace.connectionForm.save') }}
        </el-button>
      </div>
    </div>

    <el-form label-position="top" class="connection-form__grid">
      <section class="connection-form__card">
        <h3 class="connection-form__section-title">{{ t('workspace.connectionForm.basic') }}</h3>

        <el-form-item :label="t('workspace.connectionForm.name')">
          <el-input v-model="form.name" :placeholder="t('workspace.connectionForm.namePlaceholder')" />
        </el-form-item>

        <el-form-item :label="t('workspace.connectionForm.driver')">
          <el-select v-model="form.driver" :disabled="props.mode === 'edit'">
            <el-option
              v-for="driver in connections.drivers"
              :key="driver.name"
              :label="driver.name"
              :value="driver.name"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('workspace.connectionForm.root')">
          <el-input v-model="form.root" :placeholder="t('workspace.connectionForm.rootPlaceholder')" />
        </el-form-item>

        <el-form-item :label="t('workspace.connectionForm.description')">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            :placeholder="t('workspace.connectionForm.descriptionPlaceholder')"
          />
        </el-form-item>

        <div class="connection-form__toggles">
          <label class="connection-form__toggle">
            <span>{{ t('workspace.connectionForm.enabled') }}</span>
            <el-switch v-model="form.enabled" />
          </label>
          <label class="connection-form__toggle">
            <span>{{ t('workspace.connectionForm.readOnly') }}</span>
            <el-switch v-model="form.readOnly" />
          </label>
        </div>
      </section>

      <section class="connection-form__card">
        <h3 class="connection-form__section-title">{{ t('workspace.connectionForm.driverConfig') }}</h3>

        <el-form-item
          v-for="field in currentDriverFields"
          :key="field.key"
          :label="field.label"
        >
          <el-input
            v-if="field.kind === 'text' || field.kind === 'password'"
            v-model="form.config[field.key]"
            :type="field.kind"
            :show-password="field.kind === 'password'"
            :placeholder="field.placeholder"
          />
          <el-input
            v-else-if="field.kind === 'textarea'"
            v-model="form.config[field.key]"
            type="textarea"
            :rows="4"
            :placeholder="field.placeholder"
          />
          <el-input-number
            v-else-if="field.kind === 'number'"
            v-model="form.config[field.key]"
            :min="field.min ?? 0"
            controls-position="right"
          />
          <el-switch v-else v-model="form.config[field.key]" />
        </el-form-item>
      </section>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
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
const hydratedEditModel = ref(false)

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

function applyDefinition(def?: ReturnType<typeof connections.getDefinition> | null) {
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
    return
  }

  form.id = def.id
  form.name = def.name
  form.driver = def.driver
  form.description = def.description ?? ''
  form.enabled = def.enabled
  form.readOnly = def.readOnly ?? false
  form.root = def.root ?? ''
  form.tags = def.tags ?? []
  form.metadata = def.metadata ?? {}
  form.config = {
    ...createDriverConfig(def.driver),
    ...(def.config ?? {}),
  }
}

function resetForm() {
  applyDefinition(props.mode === 'edit' ? connections.getDefinition(props.connectionId ?? '') : null)
}

async function submit() {
  saving.value = true
  try {
    const saved = await connections.saveConnection({
      ...form,
      config: { ...form.config },
      metadata: { ...form.metadata },
      tags: [...form.tags],
    })
    form.id = saved.id
    await connections.refreshConnections()
    workspace.openConnection(saved.id, saved.name)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    saving.value = false
  }
}

watch(
  () => form.driver,
  (driver, prev) => {
    if (driver === prev) return
    if (props.mode === 'edit' && !hydratedEditModel.value) return
    form.config = createDriverConfig(driver)
  },
)

connections.hydrate().then(() => {
  if (props.mode === 'edit' && props.connectionId) {
    hydratedEditModel.value = true
    const def = connections.getDefinition(props.connectionId)
    if (def) {
      applyDefinition(def)
    }
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

@media (max-width: 900px) {
  .connection-form {
    padding: 16px;
  }

  .connection-form__header {
    flex-direction: column;
  }
}
</style>
