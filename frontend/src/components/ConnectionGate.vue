<script setup lang="ts">
import { computed } from 'vue'
import { NAlert, NButton, NCard, NEmpty, NSpace, NTag } from 'naive-ui'

type GateStatus = 'empty' | 'no-default' | 'default-offline'

const props = defineProps<{
  status: GateStatus
  currentPageLabel: string
  connectionCount: number
  defaultConnectionName: string
  loading: boolean
  testingDefault: boolean
}>()

const emit = defineEmits<{
  (e: 'open-connections'): void
  (e: 'refresh'): void
  (e: 'test-default'): void
}>()

const title = computed(() => {
  if (props.status === 'empty') return '欢迎使用，先创建连接'
  if (props.status === 'no-default') return '请先设置默认连接'
  return '默认连接当前不可用'
})

const description = computed(() => {
  if (props.status === 'empty') {
    return '当前没有任何连接配置，先创建一个 NameServer 连接后再使用仪表盘和管理功能。'
  }
  if (props.status === 'no-default') {
    return '检测到已有连接配置，但尚未设置默认连接。请先在连接管理中指定一个默认连接。'
  }
  return '检测到默认连接离线。建议先测试连接，确认可用后再进入业务页面。'
})
</script>

<template>
  <div class="connection-gate">
    <n-card class="connection-gate-card" :bordered="false">
      <n-empty :description="title" size="large" />

      <p class="connection-gate-desc">{{ description }}</p>

      <div class="connection-gate-meta">
        <n-tag size="small" type="info">连接数：{{ connectionCount }}</n-tag>
        <n-tag v-if="defaultConnectionName" size="small" type="warning">默认连接：{{ defaultConnectionName }}</n-tag>
      </div>

      <n-alert v-if="status === 'default-offline'" type="warning" :show-icon="false" class="connection-gate-alert">
        无法进入「{{ currentPageLabel }}」页面，请先恢复默认连接状态。
      </n-alert>

      <n-space justify="center" class="connection-gate-actions">
        <n-button type="primary" @click="emit('open-connections')">去连接管理</n-button>
        <n-button v-if="status === 'default-offline'" :loading="testingDefault" @click="emit('test-default')">
          测试默认连接
        </n-button>
        <n-button :loading="loading" @click="emit('refresh')">刷新状态</n-button>
      </n-space>
    </n-card>
  </div>
</template>

<style scoped>
.connection-gate {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.connection-gate-card {
  width: min(720px, calc(100vw - 120px));
  border-radius: 12px;
  text-align: center;
  padding: 8px 6px;
}

.connection-gate-desc {
  margin: 16px 0 10px;
  color: var(--text-secondary, #666);
  line-height: 1.7;
}

.connection-gate-meta {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.connection-gate-alert {
  margin: 10px auto 0;
  max-width: 520px;
  text-align: left;
}

.connection-gate-actions {
  margin-top: 16px;
}
</style>
