<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NGrid,
  NGi,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NSwitch,
  NSelect,
  NSpace,
  NTag,
  useMessage
} from 'naive-ui'
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui'
import * as ConnectionService from '../../bindings/rocket-leaf/internal/service/connectionservice'
import type { Connection } from '../../bindings/rocket-leaf/internal/model/models'
import { addConnectionCompat, updateConnectionCompat } from '../utils/connectionServiceCompat'
import { emitConnectionsChanged } from '../utils/connectionEvents'

type ConnectionStatus = 'online' | 'offline' | 'unknown'
type ConnectionEnv = '生产' | '测试' | '开发'
type SummaryTone = 'neutral' | 'success' | 'warning'

interface ConnectionItem {
  id: number
  name: string
  env: ConnectionEnv
  nameServer: string
  timeoutSec: number
  enableACL: boolean
  accessKey: string
  secretKey: string
  status: ConnectionStatus
  lastCheck: string
  isDefault: boolean
  remark: string
}

interface ConnectionFormModel {
  name: string
  env: ConnectionEnv
  nameServer: string
  timeoutSec: number
  enableACL: boolean
  accessKey: string
  secretKey: string
  remark: string
}

interface SummaryItem {
  key: string
  label: string
  value: number
  hint: string
  tone: SummaryTone
}

const envOptions = [
  { label: '生产', value: '生产' },
  { label: '测试', value: '测试' },
  { label: '开发', value: '开发' }
]

// 连接列表
const connectionList = ref<ConnectionItem[]>([])
const isLoading = ref(false)
const searchKeyword = ref('')
const showEditor = ref(false)
const isSubmitting = ref(false)
const editingId = ref<number | null>(null)
const testingIds = ref<number[]>([])
const formRef = ref<FormInst | null>(null)
const message = useMessage()

// 将后端 Connection 转换为前端 ConnectionItem
const mapConnection = (conn: Connection | null): ConnectionItem | null => {
  if (!conn) return null
  return {
    id: conn.id,
    name: conn.name,
    env: (conn.env as ConnectionEnv) || '开发',
    nameServer: conn.nameServer,
    timeoutSec: conn.timeoutSec || 5,
    enableACL: conn.enableACL || false,
    accessKey: conn.accessKey || '',
    secretKey: conn.secretKey || '',
    status: conn.status as ConnectionStatus || 'unknown',
    lastCheck: conn.lastCheck || '-',
    isDefault: conn.isDefault || false,
    remark: conn.remark || ''
  }
}

// 加载连接列表
const loadConnections = async () => {
  isLoading.value = true
  try {
    const connections = await ConnectionService.GetConnections()
    connectionList.value = connections
      .map(mapConnection)
      .filter((c): c is ConnectionItem => c !== null)
  } catch (err) {
    console.error('加载连接列表失败:', err)
    message.error('加载连接列表失败')
  } finally {
    isLoading.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadConnections()
})

const createDefaultFormModel = (): ConnectionFormModel => ({
  name: '',
  env: '开发',
  nameServer: '',
  timeoutSec: 5,
  enableACL: false,
  accessKey: '',
  secretKey: '',
  remark: ''
})

const formModel = ref<ConnectionFormModel>(createDefaultFormModel())

const formRules: FormRules = {
  name: [{ required: true, message: '请输入连接名称', trigger: ['blur', 'input'] }],
  env: [{ required: true, message: '请选择环境', trigger: ['change'] }],
  nameServer: [{ required: true, message: '请输入 NameServer 地址', trigger: ['blur', 'input'] }],
  timeoutSec: [{ required: true, type: 'number', message: '请设置超时时间', trigger: ['blur', 'change'] }],
  accessKey: [{
    validator: (_rule, value: string) => {
      if (formModel.value.enableACL && !value.trim()) {
        return new Error('启用 ACL 时 AccessKey 不能为空')
      }
      return true
    },
    trigger: ['blur', 'input']
  }],
  secretKey: [{
    validator: (_rule, value: string) => {
      if (formModel.value.enableACL && !value.trim()) {
        return new Error('启用 ACL 时 SecretKey 不能为空')
      }
      return true
    },
    trigger: ['blur', 'input']
  }]
}

const filteredConnections = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return connectionList.value
  return connectionList.value.filter(item =>
    item.name.toLowerCase().includes(keyword)
    || item.nameServer.toLowerCase().includes(keyword)
    || item.env.toLowerCase().includes(keyword)
  )
})

const onlineCount = computed(() => connectionList.value.filter(item => item.status === 'online').length)

const summaryItems = computed<SummaryItem[]>(() => {
  const total = connectionList.value.length
  const online = onlineCount.value
  const offline = total - online
  const health = total > 0 ? Math.round((online / total) * 100) : 0

  return [
    { key: 'total', label: '连接总数', value: total, hint: '已配置连接', tone: 'neutral' },
    { key: 'online', label: '在线连接', value: online, hint: `健康度 ${health}%`, tone: 'success' },
    { key: 'offline', label: '离线连接', value: offline, hint: offline > 0 ? '建议排查' : '状态稳定', tone: offline > 0 ? 'warning' : 'success' }
  ]
})

const openCreate = () => {
  editingId.value = null
  formModel.value = createDefaultFormModel()
  showEditor.value = true
}

const openEdit = (row: ConnectionItem) => {
  editingId.value = row.id
  formModel.value = {
    name: row.name,
    env: row.env,
    nameServer: row.nameServer,
    timeoutSec: row.timeoutSec,
    enableACL: row.enableACL,
    accessKey: row.accessKey,
    secretKey: row.secretKey,
    remark: row.remark
  }
  showEditor.value = true
}

const handleACLToggle = (enabled: boolean) => {
  formModel.value.enableACL = enabled
  if (!enabled) {
    formModel.value.accessKey = ''
    formModel.value.secretKey = ''
  }
}

// 设置默认连接
const setAsDefault = async (id: number) => {
  try {
    await ConnectionService.SetDefaultConnection(id)
    await loadConnections()
    emitConnectionsChanged()
    message.success('已设置为默认连接')
  } catch (err) {
    console.error('设置默认连接失败:', err)
    message.error('设置默认连接失败')
  }
}

// 删除连接
const deleteConnection = async (id: number) => {
  try {
    await ConnectionService.DeleteConnection(id)
    await loadConnections()
    emitConnectionsChanged()
    message.success('连接已删除')
  } catch (err) {
    console.error('删除连接失败:', err)
    message.error('删除连接失败')
  }
}

// 测试连接
const testConnection = async (id: number) => {
  if (testingIds.value.includes(id)) return
  testingIds.value.push(id)

  try {
    const result = await ConnectionService.TestConnection(id)
    await loadConnections()
    emitConnectionsChanged()
    if (result === 'online') {
      message.success('连接测试成功')
    } else {
      message.warning('连接测试失败')
    }
  } catch (err) {
    console.error('测试连接失败:', err)
    message.error('测试连接失败')
  } finally {
    testingIds.value = testingIds.value.filter(itemId => itemId !== id)
  }
}

// 保存连接
const saveConnection = async () => {
  if (!formRef.value) return
  await formRef.value.validate()
  isSubmitting.value = true

  try {
    const payload = formModel.value
    if (editingId.value === null) {
      // 新建连接
      await addConnectionCompat({
        ...payload,
        accessKey: payload.accessKey.trim(),
        secretKey: payload.secretKey.trim()
      })
      message.success('连接创建成功')
    } else {
      // 更新连接
      await updateConnectionCompat(editingId.value, {
        ...payload,
        accessKey: payload.accessKey.trim(),
        secretKey: payload.secretKey.trim()
      })
      message.success('连接更新成功')
    }
    await loadConnections()
    emitConnectionsChanged()
    showEditor.value = false
  } catch (err) {
    console.error('保存连接失败:', err)
    const errorMessage = err instanceof Error ? err.message : '保存连接失败'
    message.error(errorMessage)
  } finally {
    isSubmitting.value = false
  }
}

const tableScrollX = 900

const columns: DataTableColumns<ConnectionItem> = [
  {
    title: '连接名称',
    key: 'name',
    width: 120,
    render: (row) => h('div', { class: 'name-cell' }, [
      row.name,
      row.isDefault ? h(NTag, { size: 'small', type: 'info', style: 'margin-left: 8px' }, { default: () => '默认' }) : null
    ])
  },
  { title: '环境', key: 'env', width: 60 },
  { title: 'NameServer', key: 'nameServer', minWidth: 90, ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 60,
    render: (row) => h(NTag, {
      type: row.status === 'online' ? 'success' : row.status === 'offline' ? 'error' : 'warning',
      size: 'small'
    }, { default: () => row.status === 'online' ? '在线' : row.status === 'offline' ? '离线' : '未知' })
  },
  { title: '超时(秒)', key: 'timeoutSec', width: 100 },
  { title: '最近检测', key: 'lastCheck', width: 168 },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render: (row) => h(NSpace, { size: 'small' }, {
      default: () => [
        h(NButton, { text: true, type: 'primary', onClick: () => testConnection(row.id), loading: testingIds.value.includes(row.id) }, { default: () => '测试' }),
        h(NButton, { text: true, type: 'info', onClick: () => openEdit(row) }, { default: () => '编辑' }),
        h(NButton, { text: true, type: 'warning', onClick: () => setAsDefault(row.id), disabled: row.isDefault }, { default: () => '设为默认' }),
        h(NPopconfirm, { onPositiveClick: () => deleteConnection(row.id) }, {
          trigger: () => h(NButton, { text: true, type: 'error' }, { default: () => '删除' }),
          default: () => '确认删除此连接？'
        })
      ]
    })
  }
]
</script>

<template>
  <div class="connection-management">
    <!-- 概览卡片 -->
    <NGrid :cols="3" :x-gap="16" :y-gap="16" class="summary-grid">
      <NGi v-for="item in summaryItems" :key="item.key">
        <NCard size="small" hoverable :class="['summary-card', `tone-${item.tone}`]">
          <div class="summary-value">{{ item.value }}</div>
          <div class="summary-label">{{ item.label }}</div>
          <div class="summary-hint">{{ item.hint }}</div>
        </NCard>
      </NGi>
    </NGrid>

    <!-- 搜索与操作栏 -->
    <div class="toolbar">
      <NInput v-model:value="searchKeyword" placeholder="搜索连接名称 / NameServer / 环境" clearable style="width: 320px" />
      <NSpace>
        <NButton @click="loadConnections" :loading="isLoading">刷新</NButton>
        <NButton type="primary" @click="openCreate">新建连接</NButton>
      </NSpace>
    </div>

    <!-- 数据表格 -->
    <NDataTable :columns="columns" :data="filteredConnections" :scroll-x="tableScrollX" :loading="isLoading"
      :bordered="true" striped class="connection-table" />

    <!-- 编辑弹窗 -->
    <NModal v-model:show="showEditor" preset="card" :title="editingId === null ? '新建连接' : '编辑连接'" style="width: 620px"
      :mask-closable="false">
      <NForm ref="formRef" class="connection-editor-form" :model="formModel" :rules="formRules" label-placement="left"
        label-width="140px">
        <NFormItem label="连接名称" path="name">
          <NInput v-model:value="formModel.name" placeholder="如：生产集群" />
        </NFormItem>
        <NFormItem label="环境" path="env">
          <NSelect v-model:value="formModel.env" :options="envOptions" placeholder="请选择环境" />
        </NFormItem>
        <NFormItem label="NameServer" path="nameServer">
          <NInput v-model:value="formModel.nameServer" placeholder="如：192.168.1.100:9876" />
        </NFormItem>
        <NFormItem label="超时时间" path="timeoutSec">
          <NInputNumber v-model:value="formModel.timeoutSec" :min="1" :max="60" placeholder="秒" style="width: 100%" />
        </NFormItem>

        <NFormItem label="启用 ACL">
          <NSwitch :value="formModel.enableACL" @update:value="handleACLToggle">
            <template #checked>已启用</template>
            <template #unchecked>未启用</template>
          </NSwitch>
        </NFormItem>

        <NAlert v-if="formModel.enableACL" type="info" :show-icon="false" style="margin-bottom: 12px;">
          已启用 ACL 鉴权，请填写 AccessKey 与 SecretKey。
        </NAlert>

        <NFormItem v-if="formModel.enableACL" label="AccessKey" path="accessKey">
          <NInput v-model:value="formModel.accessKey" placeholder="请输入 ACL AccessKey" />
        </NFormItem>

        <NFormItem v-if="formModel.enableACL" label="SecretKey" path="secretKey">
          <NInput v-model:value="formModel.secretKey" type="password" show-password-on="click"
            placeholder="请输入 ACL SecretKey" />
        </NFormItem>

        <NFormItem label="备注" path="remark">
          <NInput v-model:value="formModel.remark" type="textarea" placeholder="连接备注信息" :rows="2" />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showEditor = false">取消</NButton>
          <NButton type="primary" :loading="isSubmitting" @click="saveConnection">确定</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.connection-management {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.summary-grid {
  margin-bottom: 8px;
}

.summary-card {
  text-align: center;
}

.summary-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
}

.summary-label {
  font-size: 14px;
  color: #666;
  margin-top: 4px;
}

.summary-hint {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.tone-neutral .summary-value {
  color: #333;
}

.tone-success .summary-value {
  color: #18a058;
}

.tone-warning .summary-value {
  color: #f0a020;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.connection-table {
  flex: 1;
}

.name-cell {
  display: flex;
  align-items: center;
}

.connection-editor-form :deep(.n-form-item-label) {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  white-space: nowrap;
  word-break: keep-all;
}

.connection-editor-form :deep(.n-form-item-label__text) {
  white-space: nowrap;
  word-break: normal;
  overflow-wrap: normal;
}
</style>
