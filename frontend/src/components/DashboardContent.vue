<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NButton, NCard, NEmpty, NGi, NGrid, NProgress, NStatistic, NTag, useMessage } from 'naive-ui'
import * as ClusterService from '../../bindings/rocket-leaf/internal/service/clusterservice'
import * as ConsumerService from '../../bindings/rocket-leaf/internal/service/consumerservice'
import * as TopicService from '../../bindings/rocket-leaf/internal/service/topicservice'
import type {
  BrokerNode as BackendBrokerNode,
  ConsumerGroupItem as BackendConsumerGroupItem,
  TopicItem as BackendTopicItem
} from '../../bindings/rocket-leaf/internal/model/models'

type BrokerStatus = 'online' | 'warning' | 'offline'

interface DashboardSnapshot {
  topics: number
  consumers: number
  tpsIn: number
  tpsOut: number
}

interface MetricItem {
  key: string
  label: string
  value: number
  suffix?: string
  trend: number | null
}

interface BrokerItem {
  name: string
  address: string
  status: BrokerStatus
  tpsIn: number
  tpsOut: number
}

interface AlertItem {
  title: string
  level: '低' | '中' | '高'
  time: string
}

const message = useMessage()
const isLoading = ref(false)
const lastRefreshAt = ref('-')

const currentSnapshot = ref<DashboardSnapshot>({
  topics: 0,
  consumers: 0,
  tpsIn: 0,
  tpsOut: 0
})
const previousSnapshot = ref<DashboardSnapshot | null>(null)
const brokers = ref<BrokerItem[]>([])

const pad2 = (value: number) => String(value).padStart(2, '0')

const formatDateTime = (date: Date) => {
  const y = date.getFullYear()
  const m = pad2(date.getMonth() + 1)
  const d = pad2(date.getDate())
  const hh = pad2(date.getHours())
  const mm = pad2(date.getMinutes())
  const ss = pad2(date.getSeconds())
  return `${y}-${m}-${d} ${hh}:${mm}:${ss}`
}

const calcTrend = (current: number, previous: number | null) => {
  if (previous === null) return null
  if (previous === 0) return current === 0 ? 0 : 100
  return Number((((current - previous) / previous) * 100).toFixed(1))
}

const toBrokerStatus = (status: string): BrokerStatus => {
  if (status === 'online' || status === 'warning' || status === 'offline') {
    return status
  }
  return 'offline'
}

const toBrokerItem = (broker: BackendBrokerNode): BrokerItem => ({
  name: broker.brokerName || broker.address || '未知 Broker',
  address: broker.address || '-',
  status: toBrokerStatus(broker.status || 'offline'),
  tpsIn: broker.tpsIn || 0,
  tpsOut: broker.tpsOut || 0
})

const metrics = computed<MetricItem[]>(() => {
  const current = currentSnapshot.value
  const previous = previousSnapshot.value
  return [
    {
      key: 'topics',
      label: 'Topic 总数',
      value: current.topics,
      trend: calcTrend(current.topics, previous?.topics ?? null)
    },
    {
      key: 'consumers',
      label: '消费者组',
      value: current.consumers,
      trend: calcTrend(current.consumers, previous?.consumers ?? null)
    },
    {
      key: 'tps-in',
      label: '生产 TPS',
      value: current.tpsIn,
      suffix: '/s',
      trend: calcTrend(current.tpsIn, previous?.tpsIn ?? null)
    },
    {
      key: 'tps-out',
      label: '消费 TPS',
      value: current.tpsOut,
      suffix: '/s',
      trend: calcTrend(current.tpsOut, previous?.tpsOut ?? null)
    }
  ]
})

const onlineBrokerCount = computed(() => brokers.value.filter(item => item.status === 'online').length)

const brokerHealth = computed(() => {
  if (brokers.value.length === 0) return 0
  return Math.round((onlineBrokerCount.value / brokers.value.length) * 100)
})

const alerts = computed<AlertItem[]>(() => {
  const now = lastRefreshAt.value === '-' ? formatDateTime(new Date()) : lastRefreshAt.value
  const result: AlertItem[] = []

  const offlineBrokerCount = brokers.value.filter(item => item.status === 'offline').length
  const warningBrokerCount = brokers.value.filter(item => item.status === 'warning').length

  if (brokers.value.length === 0) {
    result.push({ title: '未检测到 Broker 数据，请先连接并设置默认 NameServer', level: '高', time: now })
  }
  if (offlineBrokerCount > 0) {
    result.push({ title: `检测到 ${offlineBrokerCount} 个离线 Broker，请尽快排查`, level: '高', time: now })
  }
  if (warningBrokerCount > 0) {
    result.push({ title: `检测到 ${warningBrokerCount} 个告警 Broker，建议关注资源使用情况`, level: '中', time: now })
  }
  if (currentSnapshot.value.topics === 0) {
    result.push({ title: '当前暂无业务 Topic 数据', level: '中', time: now })
  }
  if (currentSnapshot.value.consumers === 0) {
    result.push({ title: '当前暂无消费者组数据', level: '低', time: now })
  }
  if (currentSnapshot.value.tpsIn === 0 && currentSnapshot.value.tpsOut === 0 && currentSnapshot.value.topics > 0) {
    result.push({ title: 'Topic 已存在但当前流量为 0，请确认生产和消费是否正常', level: '中', time: now })
  }

  if (result.length === 0) {
    result.push({ title: '集群运行稳定，当前未发现异常指标', level: '低', time: now })
  }

  return result.slice(0, 5)
})

const getAlertTagType = (level: AlertItem['level']) => {
  if (level === '高') return 'error'
  if (level === '中') return 'warning'
  return 'info'
}

const getTrendClass = (trend: number | null) => {
  if (trend === null) return 'metric-trend metric-trend-neutral'
  return trend >= 0 ? 'metric-trend metric-trend-up' : 'metric-trend metric-trend-down'
}

const formatTrend = (trend: number | null) => {
  if (trend === null) return '--'
  if (trend > 0) return `+${trend}%`
  return `${trend}%`
}

const loadDashboardData = async () => {
  isLoading.value = true
  const failedParts: string[] = []

  const loadWithFallback = async <T>(loader: () => Promise<T>, fallback: T, part: string): Promise<T> => {
    try {
      return await loader()
    } catch (err) {
      console.error(`加载${part}失败:`, err)
      failedParts.push(part)
      return fallback
    }
  }

  try {
    const [topicTotal, topics, groups, rawBrokers] = await Promise.all([
      loadWithFallback(TopicService.GetTopicTotal, 0, 'Topic 总数'),
      loadWithFallback(TopicService.GetTopics, [] as (BackendTopicItem | null)[], 'Topic 数据'),
      loadWithFallback(ConsumerService.GetConsumerGroups, [] as (BackendConsumerGroupItem | null)[], '消费者组数据'),
      loadWithFallback(ClusterService.GetBrokers, [] as (BackendBrokerNode | null)[], 'Broker 数据')
    ])

    const topicList = topics.filter((item): item is BackendTopicItem => item !== null)
    const groupList = groups.filter((item): item is BackendConsumerGroupItem => item !== null)
    const brokerList = rawBrokers.filter((item): item is BackendBrokerNode => item !== null)

    const detailResults = await Promise.allSettled(
      brokerList.map(async (broker) => {
        if (!broker.address) return broker
        const detail = await ClusterService.GetBrokerDetail(broker.address)
        return detail ?? broker
      })
    )

    const detailedBrokers = detailResults
      .map((result, index) => {
        if (result.status === 'fulfilled' && result.value) {
          return result.value
        }
        if (result.status === 'rejected') {
          console.error(`加载 Broker 详情失败(${brokerList[index]?.address || '-'})：`, result.reason)
        }
        return brokerList[index]
      })
      .filter((item): item is BackendBrokerNode => item !== undefined)

    const nextBrokers = detailedBrokers.map(toBrokerItem)
    const totalTpsIn = nextBrokers.reduce((sum, broker) => sum + broker.tpsIn, 0)
    const totalTpsOut = nextBrokers.reduce((sum, broker) => sum + broker.tpsOut, 0)

    previousSnapshot.value = { ...currentSnapshot.value }
    currentSnapshot.value = {
      topics: topicTotal || topicList.length,
      consumers: groupList.length,
      tpsIn: totalTpsIn,
      tpsOut: totalTpsOut
    }
    brokers.value = nextBrokers
    lastRefreshAt.value = formatDateTime(new Date())

    if (failedParts.length > 0) {
      const uniqFailedParts = Array.from(new Set(failedParts))
      message.warning(`部分数据加载失败：${uniqFailedParts.join('、')}`)
    }
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadDashboardData()
})
</script>

<template>
  <div class="dashboard-content">
    <div class="dashboard-toolbar">
      <span class="dashboard-refresh">最近刷新：{{ lastRefreshAt }}</span>
      <n-button style="margin:0 8px;" size="small" :loading="isLoading" @click="loadDashboardData">刷新数据</n-button>
    </div>

    <n-grid responsive="screen" cols="4" :x-gap="16" :y-gap="16">
      <n-gi v-for="item in metrics" :key="item.key">
        <n-card hoverable size="small" class="metric-card">
          <n-statistic :label="item.label" :value="item.value">
            <template v-if="item.suffix" #suffix>{{ item.suffix }}</template>
          </n-statistic>
          <div class="metric-foot">
            <span :class="getTrendClass(item.trend)">{{ formatTrend(item.trend) }}</span>
            <span class="metric-desc">较上次刷新</span>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <n-grid responsive="screen" cols="1 m:2" :x-gap="16" :y-gap="16" class="section-grid">
      <n-gi>
        <n-card title="集群健康度" :bordered="false" class="panel-card">
          <div class="health-summary">
            <span>在线 Broker</span>
            <n-tag type="success" size="small">{{ onlineBrokerCount }} / {{ brokers.length }}</n-tag>
          </div>
          <n-progress type="line" :percentage="brokerHealth" :show-indicator="false" status="success" />

          <div v-if="brokers.length > 0" class="broker-list">
            <div v-for="broker in brokers" :key="broker.name + broker.address" class="broker-item">
              <div class="broker-info">
                <span class="broker-name">{{ broker.name }}</span>
                <span class="broker-address">{{ broker.address }}</span>
              </div>
              <div class="broker-meta">
                <span class="broker-tps">IN {{ broker.tpsIn }}/s · OUT {{ broker.tpsOut }}/s</span>
                <n-tag
                  :type="broker.status === 'online' ? 'success' : broker.status === 'warning' ? 'warning' : 'error'"
                  size="small" round>
                  {{ broker.status === 'online' ? '在线' : broker.status === 'warning' ? '告警' : '离线' }}
                </n-tag>
              </div>
            </div>
          </div>
          <n-empty v-else size="small" description="暂无 Broker 数据" class="panel-empty" />
        </n-card>
      </n-gi>

      <n-gi>
        <n-card title="最近告警" :bordered="false" class="panel-card">
          <div class="alert-list">
            <div v-for="alert in alerts" :key="alert.title" class="alert-item">
              <div class="alert-title">{{ alert.title }}</div>
              <div class="alert-meta">
                <n-tag size="small" :type="getAlertTagType(alert.level)" round>{{ alert.level }}</n-tag>
                <span>{{ alert.time }}</span>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<style scoped>
.dashboard-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.dashboard-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dashboard-refresh {
  color: var(--text-muted, #888);
  font-size: 12px;
}

.metric-card,
.panel-card {
  background: var(--surface-2, #fff);
  border-radius: 5px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.03);
}

.metric-foot {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.metric-trend {
  font-weight: 600;
}

.metric-trend-up {
  color: #18a058;
}

.metric-trend-down {
  color: #d03050;
}

.metric-trend-neutral {
  color: var(--text-muted, #888);
}

.metric-desc {
  color: var(--text-muted, #888);
}

.section-grid {
  margin-top: 2px;
}

.health-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  font-size: 13px;
  color: var(--text-secondary, #666);
}

.broker-list,
.alert-list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.broker-item,
.alert-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 10px;
  background: var(--surface-1, #f8f8f8);
}

.broker-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.broker-name {
  font-size: 13px;
  color: var(--text-color, #333);
}

.broker-address {
  color: var(--text-muted, #888);
  font-size: 12px;
}

.broker-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.broker-tps {
  color: var(--text-muted, #888);
  font-size: 12px;
}

.alert-item {
  align-items: flex-start;
  flex-direction: column;
  gap: 6px;
}

.alert-title {
  font-size: 13px;
  color: var(--text-color, #333);
}

.alert-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted, #888);
}

.panel-empty {
  margin-top: 12px;
}
</style>
