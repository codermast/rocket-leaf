<script setup lang="ts">
import { computed } from 'vue'
import { NCard, NGrid, NGi, NStatistic, NTag, NProgress, NButton } from 'naive-ui'

interface MetricItem {
  key: string
  label: string
  value: number
  suffix?: string
  trend: number
}

interface BrokerItem {
  name: string
  status: 'online' | 'offline'
}

interface AlertItem {
  title: string
  level: '低' | '中' | '高'
  time: string
}

const metrics: MetricItem[] = [
  { key: 'topics', label: 'Topic 总数', value: 128, trend: 6.2 },
  { key: 'consumers', label: '消费者组', value: 47, trend: 2.8 },
  { key: 'tps-in', label: '生产 TPS', value: 1860, suffix: '/s', trend: 9.1 },
  { key: 'tps-out', label: '消费 TPS', value: 1742, suffix: '/s', trend: 7.6 }
]

const brokers: BrokerItem[] = [
  { name: 'broker-a', status: 'online' },
  { name: 'broker-b', status: 'online' },
  { name: 'broker-c', status: 'offline' }
]

const alerts: AlertItem[] = [
  { title: 'Topic order_event 堆积超过阈值', level: '高', time: '2 分钟前' },
  { title: '消费者组 payment_group 延迟偏高', level: '中', time: '8 分钟前' },
  { title: '测试环境连接波动已恢复', level: '低', time: '25 分钟前' }
]

const onlineBrokerCount = computed(() => brokers.filter(item => item.status === 'online').length)
const brokerHealth = computed(() => {
  if (brokers.length === 0) return 0
  return Math.round((onlineBrokerCount.value / brokers.length) * 100)
})

const getAlertTagType = (level: AlertItem['level']) => {
  if (level === '高') return 'error'
  if (level === '中') return 'warning'
  return 'info'
}
</script>

<template>
  <div class="dashboard-content">
    <n-grid responsive="screen" cols="1 s:2 m:4" :x-gap="16" :y-gap="16">
      <n-gi v-for="item in metrics" :key="item.key">
        <n-card size="small" :bordered="false" class="metric-card">
          <n-statistic :label="item.label" :value="item.value">
            <template v-if="item.suffix" #suffix>{{ item.suffix }}</template>
          </n-statistic>
          <div class="metric-foot">
            <span class="metric-trend">+{{ item.trend }}%</span>
            <span class="metric-desc">较昨日</span>
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

          <div class="broker-list">
            <div v-for="broker in brokers" :key="broker.name" class="broker-item">
              <span>{{ broker.name }}</span>
              <n-tag :type="broker.status === 'online' ? 'success' : 'error'" size="small" round>
                {{ broker.status === 'online' ? '在线' : '离线' }}
              </n-tag>
            </div>
          </div>
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

          <div class="panel-action">
            <n-button size="small" quaternary>查看全部告警</n-button>
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
}

.metric-card,
.panel-card {
  background: var(--surface-2, #fff);
  border-radius: 12px;
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
  color: #18a058;
  font-weight: 600;
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

.panel-action {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>
