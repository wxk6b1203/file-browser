import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { ElMessage } from 'element-plus'
import { CloseConnection, DeleteConnection, ListConnections, ListConnectionStates, ListDrivers, OpenConnection, SaveConnection, TestConnection } from '../../wailsjs/go/main/App'
import { connection, folder } from '../../wailsjs/go/models'
import { useNotificationsStore } from './notifications'

function sortConnections(items: connection.Definition[]): connection.Definition[] {
  return [...items].sort((a, b) => a.name.localeCompare(b.name))
}

function cloneDefinition(item: connection.Definition): connection.Definition {
  return new connection.Definition(JSON.parse(JSON.stringify(item)))
}

export const useConnectionsStore = defineStore('connections', () => {
  const notifications = useNotificationsStore()
  const ready = ref(false)
  const loading = ref(false)
  const drivers = ref<folder.DriverInfo[]>([])
  const definitions = ref<connection.Definition[]>([])
  const states = ref<connection.State[]>([])
  let pendingHydration: Promise<void> | null = null

  const definitionMap = computed(() => {
    const map = new Map<string, connection.Definition>()
    for (const item of definitions.value) {
      map.set(item.id, item)
    }
    return map
  })

  const stateMap = computed(() => {
    const map = new Map<string, connection.State>()
    for (const item of states.value) {
      map.set(item.id, item)
    }
    return map
  })

  async function hydrate() {
    if (pendingHydration) {
      return pendingHydration
    }

    loading.value = true
    pendingHydration = (async () => {
      try {
        const [driverItems, definitionItems, stateItems] = await Promise.all([
          ListDrivers(),
          ListConnections(),
          ListConnectionStates(),
        ])
        drivers.value = driverItems
        definitions.value = sortConnections(definitionItems)
        states.value = stateItems
        ready.value = true
      } finally {
        loading.value = false
        pendingHydration = null
      }
    })()

    return pendingHydration
  }

  async function refreshConnections() {
    definitions.value = sortConnections(await ListConnections())
  }

  async function refreshStates() {
    states.value = await ListConnectionStates()
  }

  async function saveConnection(def: connection.Definition) {
    const saved = await SaveConnection(def)
    await Promise.all([refreshConnections(), refreshStates()])
    ElMessage.success(saved.id === def.id && def.id ? '连接已保存' : '连接已创建')
    return saved
  }

  async function testConnection(def: connection.Definition) {
    return TestConnection(def)
  }

  async function deleteConnection(id: string) {
    await DeleteConnection(id)
    definitions.value = definitions.value.filter((item) => item.id !== id)
    states.value = states.value.filter((item) => item.id !== id)
    ElMessage.success('连接已删除')
  }

  async function openConnection(id: string) {
    try {
      const state = await OpenConnection(id)
      await refreshStates()
      return state
    } catch (error) {
      const definition = definitionMap.value.get(id)
      notifications.push({
        level: 'error',
        source: definition?.name || 'Connection',
        title: 'Connection Open Failed',
        message: error instanceof Error ? error.message : String(error),
      })
      throw error
    }
  }

  async function closeConnection(id: string) {
    try {
      await CloseConnection(id)
      await refreshStates()
    } catch (error) {
      const definition = definitionMap.value.get(id)
      notifications.push({
        level: 'error',
        source: definition?.name || 'Connection',
        title: 'Connection Close Failed',
        message: error instanceof Error ? error.message : String(error),
      })
      throw error
    }
  }

  function getDefinition(id: string) {
    const item = definitionMap.value.get(id)
    return item ? cloneDefinition(item) : null
  }

  function getState(id: string) {
    const item = stateMap.value.get(id)
    return item ? new connection.State(JSON.parse(JSON.stringify(item))) : null
  }

  return {
    ready,
    loading,
    drivers,
    definitions,
    states,
    definitionMap,
    stateMap,
    hydrate,
    refreshConnections,
    refreshStates,
    saveConnection,
    testConnection,
    deleteConnection,
    openConnection,
    closeConnection,
    getDefinition,
    getState,
  }
})
