<script setup lang="ts">
// Sidebar.vue
// 使用 Naive UI 和 Ionicons5 图标构建的左侧边栏
import { ref, h, computed } from 'vue'
import type { Component } from 'vue'
import { NMenu, NIcon, NLayoutSider, NTooltip, NDropdown, NAvatar, NText, NBadge } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import {
    SpeedometerOutline,
    LinkOutline,
    FolderOutline,
    PeopleOutline,
    ChatbubblesOutline,
    ServerOutline,
    SettingsOutline,
    ChevronDownOutline,
    AddOutline,
    LogoGithub,
    OpenOutline,
} from '@vicons/ionicons5'
import { render } from 'naive-ui/es/_utils'

// 渲染图标的辅助函数
const renderIcon = (icon: Component) => {
    return () => h(NIcon, null, { default: () => h(icon) })
}

// RocketMQ 图标（来自 Iconify 的 simple-icons:apacherocketmq）
const RocketMQIcon = {
    render() {
        return h('span', { class: 'rocketmq-icon' })
    }
}

// 侧边栏是否折叠
const collapsed = ref(false)

// 当前选中的菜单项
const activeKey = ref<string>('dashboard')

// 连接下拉显示状态
const showConnectionDropdown = ref(false)

// RocketMQ 实例列表（模拟数据）
interface Instance {
    label: string
    value: string
    ip: string
    status: 'online' | 'offline'  // 连接状态
    icon: Component
}

const fallbackInstance: Instance = {
    label: '未连接实例',
    value: 'fallback',
    ip: '-',
    status: 'offline',
    icon: PeopleOutline
}

const instances = ref<Instance[]>([
    { label: '生产环境', value: 'prod', ip: '192.168.1.100:9876', status: 'online', icon: PeopleOutline },
    { label: '测试环境', value: 'test', ip: '192.168.1.101:9876', status: 'online', icon: LogoGithub },
    { label: '开发环境', value: 'dev', ip: 'localhost:9876', status: 'offline', icon: SpeedometerOutline }
])

// 当前选中的实例
const selectedInstance = ref<string>('prod')

// 获取当前选中实例的信息
const currentInstance = computed<Instance>(() => {
    return instances.value.find(i => i.value === selectedInstance.value) ?? instances.value[0] ?? fallbackInstance
})

// 选择实例
const selectInstance = (value: string) => {
    selectedInstance.value = value
}

// 添加新连接
const addNewConnection = () => {
    // TODO: 打开添加连接对话框
    console.log('Add new connection')
}

const addConnectionKey = '__add_connection__'

const handleConnectionItemClick = (value: string) => {
    selectInstance(value)
    showConnectionDropdown.value = false
}

const renderConnectionDropdownHeader = () => {
    return h(
        'div',
        {
            style: 'display: flex; align-items: center; padding: 8px 12px;'
        },
        [
            h(NAvatar, {
                round: true,
                style: 'margin-right: 12px;',
                src: 'https://07akioni.oss-cn-beijing.aliyuncs.com/demo1.JPG'
            }),
            h('div', null, [
                h('div', null, [h(NText, { depth: 2 }, { default: () => 'RocketMQ 集群列表' })]),
                h('div', { style: 'font-size: 12px;' }, [
                    h(
                        NText,
                        { depth: 3 },
                        { default: () => '选择 RocketMQ 实例集群' }
                    )
                ])
            ])
        ]
    )
}

// const renderConnectionOptionCard = (instance: Instance) => {
//     const isSelected = selectedInstance.value === instance.value
//     return h('div', { style: 'color: blue;' }, '特殊蓝字选项')
// }

const connectionDropdownOptions = computed(() => {
    const instanceOptions = instances.value.map((instance) => ({
        key: instance.value,
        label: () =>
            h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
                // 实例名称
                h('span', null, instance.label),
                // 状态标签
                h(NBadge, {
                    value: instance.status === 'online' ? '在线' : '离线',
                    type: instance.status === 'online' ? 'success' : 'error',
                    // 如果你只想要个小圆点，不想要文字，可以用 dot 属性
                    // dot: true 
                })]),
        icon: renderIcon(instance.icon)
    }))

    return [
        {
            key: 'connection-header',
            type: 'render' as const,
            render: renderConnectionDropdownHeader
        },
        {
            type: 'divider',
            key: 'connection-header-divider'
        },
        ...instanceOptions,
        {
            type: 'divider',
            key: 'connection-action-divider'
        },
        {
            key: addConnectionKey,
            label: '添加新连接',
            icon: renderIcon(AddOutline)
        }
    ]
})

const handleConnectionSelect = (key: string | number) => {
    const selectedKey = String(key)
    if (selectedKey === addConnectionKey) {
        addNewConnection()
        showConnectionDropdown.value = false
        return
    }
    handleConnectionItemClick(selectedKey)
}

// 菜单选项 - 只保留一级导航
const menuOptions: MenuOption[] = [
    {
        label: '仪表盘',
        key: 'dashboard',
        icon: renderIcon(SpeedometerOutline)
    },
    {
        label: '连接管理',
        key: 'connections',
        icon: renderIcon(LinkOutline)
    },
    {
        label: 'Topic 管理',
        key: 'topics',
        icon: renderIcon(FolderOutline)
    },
    {
        label: '消费者组',
        key: 'consumer-groups',
        icon: renderIcon(PeopleOutline)
    },
    {
        label: '消息查询',
        key: 'messages',
        icon: renderIcon(ChatbubblesOutline)
    },
    {
        label: '集群状态',
        key: 'cluster',
        icon: renderIcon(ServerOutline)
    }
]

// 打开 GitHub 仓库
const openGithub = () => {
    window.open('https://github.com/codermast/rocket-leaf', '_blank')
}

// 定义事件
const emit = defineEmits<{
    (e: 'update:currentPage', key: string): void
    (e: 'open:settings'): void
}>()

// 处理菜单选择
const handleSelect = (key: string) => {
    activeKey.value = key
    emit('update:currentPage', key)
    console.log('Selected:', key)
}

const openSettings = () => {
    emit('open:settings')
}
</script>

<template>
    <n-layout-sider bordered collapse-mode="width" :collapsed-width="64" :width="220" :collapsed="collapsed"
        show-trigger @collapse="collapsed = true" @expand="collapsed = false" class="sidebar">
        <!-- 实例选择器区域 -->
        <div class="instance-selector-wrapper">
            <n-dropdown trigger="click" placement="bottom-start" :show-arrow="true" :options="connectionDropdownOptions"
                v-model:show="showConnectionDropdown" :value="selectedInstance" @select="handleConnectionSelect">
                <div class="instance-card" :class="{ collapsed }">
                    <!-- 折叠状态：只显示图标 -->
                    <template v-if="collapsed">
                        <n-tooltip placement="right" :show-arrow="true">
                            <template #trigger>
                                <div class="instance-card-icon-only">
                                    <n-icon :size="20">
                                        <RocketMQIcon />
                                    </n-icon>
                                    <span class="status-dot collapsed-dot" :class="currentInstance.status"></span>
                                </div>
                            </template>
                            <div>{{ currentInstance.label }}</div>
                            <div>{{ currentInstance.ip }}</div>
                        </n-tooltip>
                    </template>

                    <!-- 展开状态：显示卡片 -->
                    <template v-else>
                        <div class="instance-card-content">
                            <div class="instance-card-icon">
                                <n-icon :size="20">
                                    <RocketMQIcon />
                                </n-icon>
                            </div>
                            <div class="instance-card-info">
                                <div class="instance-card-name">
                                    {{ currentInstance.label }}
                                    <span class="status-dot inline" :class="currentInstance.status"></span>
                                </div>
                            </div>
                            <n-icon :size="16" class="instance-card-arrow">
                                <ChevronDownOutline />
                            </n-icon>
                        </div>
                    </template>
                </div>
            </n-dropdown>
        </div>

        <!-- 菜单 -->
        <n-menu :value="activeKey" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="20"
            :options="menuOptions" @update:value="handleSelect" class="main-menu" />

        <!-- 底部区域 -->
        <div class="sidebar-footer" :class="{ collapsed }">
            <!-- 设置 -->
            <n-tooltip v-if="collapsed" placement="right">
                <template #trigger>
                    <div class="footer-item" @click="openSettings">
                        <div class="footer-icon">
                            <n-icon :size="20">
                                <SettingsOutline />
                            </n-icon>
                        </div>
                    </div>
                </template>
                设置
            </n-tooltip>
            <div v-else class="footer-item" @click="openSettings">
                <div class="footer-icon">
                    <n-icon :size="20">
                        <SettingsOutline />
                    </n-icon>
                </div>
                <span class="footer-label">设置</span>
            </div>

            <!-- GitHub -->
            <n-tooltip v-if="collapsed" placement="right">
                <template #trigger>
                    <div class="footer-item github-item" @click="openGithub">
                        <div class="footer-icon">
                            <n-icon :size="20">
                                <LogoGithub />
                            </n-icon>
                        </div>
                    </div>
                </template>
                <div>
                    <div style="font-weight: 500;">Rocket Leaf</div>
                    <div style="font-size: 12px; color: #999;">开源 RocketMQ 客户端</div>
                </div>
            </n-tooltip>
            <div v-else class="footer-item github-item" @click="openGithub">
                <div class="footer-icon">
                    <n-icon :size="20">
                        <LogoGithub />
                    </n-icon>
                </div>
                <div class="github-info">
                    <div class="github-name">
                        Rocket Leaf
                        <n-icon :size="12" style="margin-left: 4px; opacity: 0.6;">
                            <OpenOutline />
                        </n-icon>
                    </div>
                    <div class="github-desc">开源 RocketMQ 客户端</div>
                </div>
            </div>
        </div>

    </n-layout-sider>
</template>


<style scoped>
.sidebar {
    height: 100%;
    background: var(--bg-color, #ffffff);
    display: flex;
    flex-direction: column;
}

/* 确保 NLayoutSider 内部滚动容器也使用 flex */
.sidebar :deep(.n-layout-sider-scroll-container) {
    display: flex;
    flex-direction: column;
}

/* 实例选择器区域 */
.instance-selector-wrapper {
    border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.06));
}

/* 实例卡片 */
.instance-card {
    margin: 12px;
    border-radius: 10px;
    background: var(--surface-1, #f8f8f8);
    cursor: pointer;
    transition: all 0.2s ease;
}

.instance-card:hover {
    background: var(--surface-1-hover, #f0f0f0);
}

.instance-card.collapsed {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 0;
    height: var(--menu-item-height, 42px);
    margin: 6px 0;
    background: transparent;
    position: relative;
}

.instance-card.collapsed:hover {
    background: transparent;
}

.instance-card.collapsed::before {
    content: '';
    position: absolute;
    left: 8px;
    right: 8px;
    top: 0;
    bottom: 0;
    border-radius: var(--menu-item-radius, 3px);
    background: transparent;
    transition: background-color 0.3s var(--menu-item-bezier, cubic-bezier(.4, 0, .2, 1));
    pointer-events: none;
}

.instance-card.collapsed:hover::before {
    background: var(--menu-item-hover-bg, #f3f3f5);
}

.instance-card-icon-only {
    position: relative;
    width: 36px;
    height: 36px;
    min-width: 36px;
    min-height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    background: linear-gradient(135deg, #22c372 0%, #18a058 100%);
    color: white;
    flex-shrink: 0;
    box-shadow: 0 4px 10px rgba(24, 160, 88, 0.24);
    transition: transform 0.2s cubic-bezier(.4, 0, .2, 1), box-shadow 0.2s cubic-bezier(.4, 0, .2, 1), filter 0.2s ease;
    z-index: 1;
}

.instance-card.collapsed:hover .instance-card-icon-only {
    transform: translateY(-1px);
    box-shadow: 0 8px 18px rgba(24, 160, 88, 0.3);
    filter: saturate(1.06) brightness(1.03);
}

.rocketmq-icon {
    display: inline-block;
    width: 1em;
    height: 1em;
    background-color: currentColor;
    mask: url('https://api.iconify.design/simple-icons/apacherocketmq.svg') no-repeat center / contain;
    -webkit-mask: url('https://api.iconify.design/simple-icons/apacherocketmq.svg') no-repeat center / contain;
}

.instance-card-content {
    display: flex;
    align-items: center;
    padding: 12px;
    gap: 10px;
}

.instance-card-icon {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    background: linear-gradient(135deg, #22c372 0%, #18a058 100%);
    color: white;
    flex-shrink: 0;
    box-shadow: 0 4px 10px rgba(24, 160, 88, 0.2);
    transition: transform 0.2s cubic-bezier(.4, 0, .2, 1), box-shadow 0.2s cubic-bezier(.4, 0, .2, 1), filter 0.2s ease;
}

.instance-card:hover .instance-card-icon {
    transform: translateY(-1px);
    box-shadow: 0 8px 18px rgba(24, 160, 88, 0.26);
    filter: saturate(1.06) brightness(1.03);
}

.instance-card:active .instance-card-icon,
.instance-card:active .instance-card-icon-only {
    transform: translateY(0);
    box-shadow: 0 3px 8px rgba(24, 160, 88, 0.2);
    filter: none;
}

.instance-card-info {
    flex: 1;
    min-width: 0;
    overflow: hidden;
}

.instance-card-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-color, #333);
    display: flex;
    align-items: center;
    gap: 6px;
}

.instance-card-ip {
    font-size: 12px;
    color: var(--text-secondary, #666);
    margin-top: 4px;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
    letter-spacing: 0.2px;
    background: var(--chip-bg, rgba(0, 0, 0, 0.04));
    padding: 2px 6px;
    border-radius: 6px;
    display: block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.instance-card-arrow {
    color: #999;
    flex-shrink: 0;
}

/* 状态指示点 */
.status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
}

.status-dot.online {
    background: #18a058;
}

.status-dot.offline {
    background: #e88080;
}

.status-dot.inline {
    display: inline-block;
}

.instance-card-icon-only .status-dot {
    position: absolute;
    bottom: -2px;
    right: -2px;
    width: 10px;
    height: 10px;
    border: 2px solid white;
}

/* 主菜单占据剩余空间 */
.main-menu {
    flex: 1;
    overflow-y: auto;
}

/* 底部区域 - 固定在底部 */
.sidebar-footer {
    margin-top: auto;
    border-top: 1px solid var(--border-color, rgba(0, 0, 0, 0.06));
    padding: 8px;
}

.footer-item {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    gap: 10px;
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.15s;
    color: var(--text-secondary, #666);
}

.footer-item:hover {
    background: var(--surface-2-hover, #f5f5f5);
    color: var(--text-color, #333);
}

.footer-item.active {
    background: var(--surface-2-hover, #f5f5f5);
    color: var(--text-color, #333);
}

/* 折叠状态下居中对齐 */
.sidebar-footer.collapsed {
    padding: 6px 0 8px;
}

.sidebar-footer.collapsed .footer-item {
    position: relative;
    justify-content: center;
    padding: 0;
    height: var(--menu-item-height, 42px);
    margin: 0;
    border-radius: 0;
    background: transparent;
}

.sidebar-footer.collapsed .footer-item:hover {
    background: transparent;
}

.sidebar-footer.collapsed .footer-item::before {
    content: '';
    position: absolute;
    left: 8px;
    right: 8px;
    top: 0;
    bottom: 0;
    border-radius: var(--menu-item-radius, 3px);
    background: transparent;
    transition: background-color 0.3s var(--menu-item-bezier, cubic-bezier(.4, 0, .2, 1));
    pointer-events: none;
}

.sidebar-footer.collapsed .footer-item:hover::before,
.sidebar-footer.collapsed .footer-item.active::before {
    background: var(--menu-item-hover-bg, #f3f3f5);
}

.sidebar-footer.collapsed .footer-icon {
    z-index: 1;
}

.footer-icon {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    border-radius: 6px;
    background: transparent;
    color: inherit;
    transition: background 0.15s, color 0.15s;
}

.footer-label {
    font-size: 14px;
    font-weight: 500;
}

.github-item {
    margin-top: 4px;
}

.sidebar-footer.collapsed .github-item {
    margin-top: 4px;
}

.github-info {
    flex: 1;
    min-width: 0;
}

.github-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-color, #333);
    display: flex;
    align-items: center;
}

.github-desc {
    font-size: 12px;
    color: var(--text-muted, #888);
    margin-top: 1px;
}

.connection-dropdown-header {
    padding: 6px 4px;
}

.connection-dropdown-title {
    font-size: 13px;
    font-weight: 600;
    line-height: 1.2;
    color: var(--text-color, #333);
}

.connection-dropdown-subtitle {
    margin-top: 2px;
    font-size: 12px;
    line-height: 1.25;
    color: var(--text-muted, #888);
}

.connection-option-card {
    max-width: 300px;
    min-width: 236px;
    cursor: pointer;
}

.connection-option-card :deep(.n-card__content) {
    padding: 8px 10px;
    border-radius: 10px;
    border: 1px solid transparent;
    transition: background-color 0.2s ease, border-color 0.2s ease;
}

.connection-option-card.is-selected :deep(.n-card__content) {
    background: rgba(24, 160, 88, 0.1);
    border-color: rgba(24, 160, 88, 0.24);
}

.connection-option-content {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
}

.connection-option-icon {
    width: 30px;
    height: 30px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: #fff;
    background: linear-gradient(135deg, #22c372 0%, #18a058 100%);
    box-shadow: 0 2px 8px rgba(24, 160, 88, 0.22);
}

.connection-option-main {
    min-width: 0;
    flex: 1;
}

.connection-option-title-row {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
}

.connection-option-name {
    min-width: 0;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.3;
    color: var(--n-option-text-color, #303133);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.connection-option-status {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    flex-shrink: 0;
}

.connection-option-status.is-online {
    background: #18a058;
}

.connection-option-status.is-offline {
    background: #e88080;
}

.connection-option-ip {
    margin-top: 2px;
    font-size: 12px;
    line-height: 1.25;
    color: var(--n-option-text-color, #909399);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
    letter-spacing: 0.2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
