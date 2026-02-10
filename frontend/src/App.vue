<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch, watchEffect } from 'vue'
import {
  NLayout,
  NLayoutContent,
  NConfigProvider,
  NGlobalStyle,
  NModal,
  NCard,
  NButton,
  NSpace,
  NMessageProvider,
  NForm,
  NFormItem,
  NRadioGroup,
  NRadio,
  darkTheme
} from 'naive-ui'
import TitleBar from './components/TitleBar.vue'
import Sidebar from './components/Sidebar.vue'
import ContentTopBar from './components/ContentTopBar.vue'
import DashboardContent from './components/DashboardContent.vue'
import ConnectionManagement from './components/ConnectionManagement.vue'
import TopicManagement from './components/TopicManagement.vue'
import ConsumerGroupManagement from './components/ConsumerGroupManagement.vue'
import MessageQuery from './components/MessageQuery.vue'
import ClusterStatus from './components/ClusterStatus.vue'
import ConnectionGate from './components/ConnectionGate.vue'
import * as ConnectionService from '../bindings/rocket-leaf/internal/service/connectionservice'
import type { Connection } from '../bindings/rocket-leaf/internal/model/models'

type ThemeMode = 'light' | 'dark' | 'system'

interface AppSettings {
  themeMode: ThemeMode
}

type ConnectionGateStatus = 'empty' | 'no-default' | 'default-offline' | 'ready'

interface ConnectionViewItem {
  id: number
  name: string
  isDefault: boolean
  status: string
}

const SETTINGS_STORAGE_KEY = 'rocket-leaf.app.settings'

const DEFAULT_SETTINGS: AppSettings = {
  themeMode: 'light',
}

const currentPage = ref('dashboard')
const showSettings = ref(false)
const systemPrefersDark = ref(false)
const isLoadingConnectionState = ref(false)
const isTestingDefaultConnection = ref(false)
const connectionStateError = ref('')
const connections = ref<ConnectionViewItem[]>([])

const settings = reactive<AppSettings>({ ...DEFAULT_SETTINGS })

let mediaQuery: MediaQueryList | undefined
let mediaQueryListener: ((event: MediaQueryListEvent) => void) | undefined
let themeTransitionTimer: number | undefined
const themeTransitionClass = 'theme-switching'

const isDark = computed(() => {
  if (settings.themeMode === 'dark') return true
  if (settings.themeMode === 'light') return false
  return systemPrefersDark.value
})

const currentPageLabelMap: Record<string, string> = {
  dashboard: '仪表盘',
  connections: '连接管理',
  topics: 'Topic 管理',
  'consumer-groups': '消费者组',
  messages: '消息查询',
  cluster: '集群状态'
}

const currentPageLabel = computed(() => currentPageLabelMap[currentPage.value] || '当前页面')

const shouldBypassConnectionGate = computed(() => currentPage.value === 'connections')

const defaultConnection = computed(() => {
  return connections.value.find(item => item.isDefault) ?? null
})

const hasConnections = computed(() => connections.value.length > 0)

const connectionGateStatus = computed<ConnectionGateStatus>(() => {
  if (!hasConnections.value) return 'empty'
  if (!defaultConnection.value) return 'no-default'
  if (defaultConnection.value.status !== 'online') return 'default-offline'
  return 'ready'
})

const shouldShowConnectionGate = computed(() => {
  if (shouldBypassConnectionGate.value) return false
  return connectionGateStatus.value !== 'ready'
})

const mapConnection = (conn: Connection | null): ConnectionViewItem | null => {
  if (!conn) return null
  return {
    id: conn.id,
    name: conn.name,
    isDefault: conn.isDefault || false,
    status: conn.status || 'offline'
  }
}

const loadConnectionState = async () => {
  isLoadingConnectionState.value = true
  connectionStateError.value = ''
  try {
    const result = await ConnectionService.GetConnections()
    connections.value = result
      .map(mapConnection)
      .filter((item): item is ConnectionViewItem => item !== null)
  } catch (err) {
    console.error('加载连接状态失败:', err)
    connectionStateError.value = '加载连接状态失败，请重试'
    connections.value = []
  } finally {
    isLoadingConnectionState.value = false
  }
}

const gotoConnectionManagement = () => {
  showSettings.value = false
  currentPage.value = 'connections'
}

const testDefaultConnection = async () => {
  if (!defaultConnection.value) return
  isTestingDefaultConnection.value = true
  try {
    await ConnectionService.TestConnection(defaultConnection.value.id)
  } catch (err) {
    console.error('测试默认连接失败:', err)
  } finally {
    isTestingDefaultConnection.value = false
    await loadConnectionState()
  }
}

const applyThemeTransition = () => {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  root.classList.add(themeTransitionClass)
  if (themeTransitionTimer) {
    window.clearTimeout(themeTransitionTimer)
  }
  themeTransitionTimer = window.setTimeout(() => {
    root.classList.remove(themeTransitionClass)
  }, 180)
}

const toggleTheme = () => {
  applyThemeTransition()
  settings.themeMode = isDark.value ? 'light' : 'dark'
}

const handlePageChange = (key: string) => {
  if (key === 'settings') {
    showSettings.value = true
    return
  }
  showSettings.value = false
  currentPage.value = key
}

const openSettings = () => {
  showSettings.value = true
}

const resetSettings = () => {
  Object.assign(settings, DEFAULT_SETTINGS)
}

const loadSettings = () => {
  if (typeof window === 'undefined') return

  const cached = window.localStorage.getItem(SETTINGS_STORAGE_KEY)
  if (cached) {
    try {
      const parsed = JSON.parse(cached) as Partial<AppSettings>
      if (parsed.themeMode === 'light' || parsed.themeMode === 'dark' || parsed.themeMode === 'system') {
        settings.themeMode = parsed.themeMode
      }
    } catch {
      Object.assign(settings, DEFAULT_SETTINGS)
    }
  }
}

onMounted(() => {
  loadSettings()
  loadConnectionState()

  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    systemPrefersDark.value = mediaQuery.matches
    mediaQueryListener = (event: MediaQueryListEvent) => {
      systemPrefersDark.value = event.matches
    }
    mediaQuery.addEventListener('change', mediaQueryListener)
  }
})

onUnmounted(() => {
  if (themeTransitionTimer) {
    window.clearTimeout(themeTransitionTimer)
  }
  if (mediaQuery && mediaQueryListener) {
    mediaQuery.removeEventListener('change', mediaQueryListener)
  }
})

watch(settings, (value) => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(value))
}, { deep: true })

watch(currentPage, () => {
  if (!shouldBypassConnectionGate.value) {
    loadConnectionState()
  }
})

watchEffect(() => {
  if (typeof document !== 'undefined') {
    document.documentElement.dataset.theme = isDark.value ? 'dark' : 'light'
  }
})
</script>

<template>
  <n-config-provider :theme="isDark ? darkTheme : null">
    <n-global-style />
    <n-message-provider>
      <div class="app-container">
        <TitleBar />
        <n-layout has-sider class="main-layout">
          <Sidebar :current-page="currentPage" @update:currentPage="handlePageChange" @open:settings="openSettings" />
          <n-layout class="content-wrapper">
            <ContentTopBar :currentPage="currentPage" :isDark="isDark" @toggle-theme="toggleTheme" />
            <n-layout-content class="content">
              <ConnectionManagement v-if="currentPage === 'connections'" />
              <div
                v-else-if="shouldShowConnectionGate"
                class="content-center-stage"
              >
                <ConnectionGate
                  :status="connectionGateStatus as 'empty' | 'no-default' | 'default-offline'"
                  :current-page-label="currentPageLabel"
                  :connection-count="connections.length"
                  :default-connection-name="defaultConnection?.name || ''"
                  :loading="isLoadingConnectionState"
                  :testing-default="isTestingDefaultConnection"
                  @open-connections="gotoConnectionManagement"
                  @refresh="loadConnectionState"
                  @test-default="testDefaultConnection"
                />
              </div>
              <DashboardContent v-else-if="currentPage === 'dashboard'" />
              <TopicManagement v-else-if="currentPage === 'topics'" />
              <ConsumerGroupManagement v-else-if="currentPage === 'consumer-groups'" />
              <MessageQuery v-else-if="currentPage === 'messages'" />
              <ClusterStatus v-else-if="currentPage === 'cluster'" />
              <div v-else class="placeholder">
                <h1>功能开发中</h1>
                <p>当前页面：{{ currentPage }}</p>
                <p>请继续从左侧菜单切换</p>
              </div>

              <div v-if="connectionStateError" class="connection-error-tip">{{ connectionStateError }}</div>
            </n-layout-content>
          </n-layout>
        </n-layout>
      </div>

      <n-modal v-model:show="showSettings" :mask-closable="true" :auto-focus="false">
        <n-card title="系统设置" size="small" class="settings-card" :bordered="false">
          <div class="settings-body">
            <n-form label-placement="left" label-width="96" :show-feedback="false">
              <n-form-item label="主题模式">
                <n-radio-group v-model:value="settings.themeMode">
                  <n-space>
                    <n-radio value="light">浅色</n-radio>
                    <n-radio value="dark">深色</n-radio>
                    <n-radio value="system">跟随系统</n-radio>
                  </n-space>
                </n-radio-group>
              </n-form-item>
            </n-form>
          </div>

          <template #footer>
            <n-space justify="space-between">
              <n-button quaternary @click="resetSettings">恢复默认</n-button>
              <n-space>
                <n-button @click="showSettings = false">关闭</n-button>
              </n-space>
            </n-space>
          </template>
        </n-card>
      </n-modal>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-color);
  overflow: hidden;
}

.main-layout {
  flex: 1;
  overflow: hidden;
}

.content-wrapper {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px;
  overflow: auto;
  background: var(--bg-color, #ffffff);
}

.content :deep(.n-layout-scroll-container) {
  flex: 1;
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

.content :deep(.n-layout-scroll-content) {
  flex: 1;
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

.content-center-stage {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.content-center-stage :deep(.connection-gate) {
  flex: 1;
}

.placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #888;
}

.placeholder h1 {
  margin-bottom: 8px;
}

.placeholder p {
  margin: 4px 0;
}

.settings-card {
  width: 560px;
  max-width: calc(100vw - 24px);
  border-radius: 12px;
}

.settings-body {
  padding: 4px 0 8px;
  color: var(--text-color, #333);
}

.connection-error-tip {
  position: fixed;
  bottom: 16px;
  right: 18px;
  background: rgba(208, 48, 80, 0.9);
  color: #fff;
  font-size: 12px;
  padding: 8px 10px;
  border-radius: 8px;
  z-index: 20;
}
</style>
