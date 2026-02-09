<script setup lang="ts">
import { ref, h, computed, onMounted, watch } from 'vue'
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
    OpenOutline
} from '@vicons/ionicons5'

import { Browser } from '@wailsio/runtime'

interface Instance {
    label: string
    value: string
    ip: string
    status: 'online' | 'offline'
    icon: Component
}

interface SidebarMenuItem {
    label: string
    key: string
    icon: Component
}

const emit = defineEmits<{
    (e: 'update:currentPage', key: string): void
    (e: 'open:settings'): void
}>()

const STORAGE_KEYS = {
    collapsed: 'rocket-leaf.sidebar.collapsed',
    activeKey: 'rocket-leaf.sidebar.active-key',
    selectedInstance: 'rocket-leaf.sidebar.selected-instance'
} as const

const renderIcon = (icon: Component) => {
    return () => h(NIcon, null, { default: () => h(icon) })
}

const RocketMQIcon = {
    render() {
        return h('span', { class: 'rocketmq-icon' })
    }
}

const collapsed = ref(false)
const activeKey = ref<string>('dashboard')
const showConnectionDropdown = ref(false)

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

const selectedInstance = ref<string>('prod')

const currentInstance = computed<Instance>(() => {
    return instances.value.find(item => item.value === selectedInstance.value) ?? instances.value[0] ?? fallbackInstance
})

const onlineInstanceCount = computed(() => instances.value.filter(item => item.status === 'online').length)
const offlineInstanceCount = computed(() => instances.value.length - onlineInstanceCount.value)

const menuCatalog: SidebarMenuItem[] = [
    { label: '仪表盘', key: 'dashboard', icon: SpeedometerOutline },
    { label: '连接管理', key: 'connections', icon: LinkOutline },
    { label: 'Topic 管理', key: 'topics', icon: FolderOutline },
    { label: '消费者组', key: 'consumer-groups', icon: PeopleOutline },
    { label: '消息查询', key: 'messages', icon: ChatbubblesOutline },
    { label: '集群状态', key: 'cluster', icon: ServerOutline }
]

const menuKeys = new Set(menuCatalog.map(item => item.key))

const menuOptions = computed<MenuOption[]>(() => {
    return menuCatalog.map((item) => ({
        key: item.key,
        icon: renderIcon(item.icon),
        label: () => h('div', { class: 'menu-option-label' }, [
            h('span', { class: 'menu-option-text' }, item.label)
        ])
    }))
})

const addConnectionKey = '__add_connection__'
const manageConnectionKey = '__manage_connection__'

const canUseStorage = () => typeof window !== 'undefined'

const writeStorage = (key: string, value: string) => {
    if (!canUseStorage()) return
    window.localStorage.setItem(key, value)
}

const readStorage = (key: string) => {
    if (!canUseStorage()) return null
    return window.localStorage.getItem(key)
}

const navigateToPage = (key: string) => {
    activeKey.value = key
    emit('update:currentPage', key)
}

const selectInstance = (value: string) => {
    selectedInstance.value = value
}

const addNewConnection = () => {
    navigateToPage('connections')
    showConnectionDropdown.value = false
}

const handleConnectionItemClick = (value: string) => {
    selectInstance(value)
    showConnectionDropdown.value = false
}

const renderConnectionDropdownHeader = () => {
    return h(
        'div',
        { style: 'display: flex; align-items: center; padding: 8px 12px;' },
        [
            h(
                NAvatar,
                {
                    round: true,
                    size: 34,
                    style: 'margin-right: 12px; background: linear-gradient(135deg, #22c372 0%, #18a058 100%); color: #fff;'
                },
                {
                    default: () => h(NIcon, { size: 16 }, { default: () => h(RocketMQIcon) })
                }
            ),
            h('div', null, [
                h('div', null, [h(NText, { depth: 2 }, { default: () => 'RocketMQ 集群列表' })]),
                h('div', { style: 'font-size: 12px;' }, [
                    h(NText, { depth: 3 }, { default: () => `在线 ${onlineInstanceCount.value} · 离线 ${offlineInstanceCount.value}` })
                ])
            ])
        ]
    )
}

const connectionDropdownOptions = computed(() => {
    const instanceOptions = instances.value.map((instance) => ({
        key: instance.value,
        label: () => h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
            h('span', null, instance.label),
            h(NBadge, {
                value: instance.status === 'online' ? '在线' : '离线',
                type: instance.status === 'online' ? 'success' : 'error'
            })
        ]),
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
            key: manageConnectionKey,
            label: '连接管理',
            icon: renderIcon(LinkOutline)
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
        return
    }

    if (selectedKey === manageConnectionKey) {
        navigateToPage('connections')
        showConnectionDropdown.value = false
        return
    }

    handleConnectionItemClick(selectedKey)
}

const handleSelect = (key: string) => {
    navigateToPage(key)
}

const openSettings = () => {
    emit('open:settings')
}

const openGithub = async () => {
    // Wails 3 的前端运行时 API
    // 这行代码会通过 IPC 发送消息给 Go 后端
    await Browser.OpenURL("https://github.com/codermast/rocket-leaf")
}

const handleCollapse = () => {
    collapsed.value = true
}

const handleExpand = () => {
    collapsed.value = false
}

onMounted(() => {
    const cachedCollapsed = readStorage(STORAGE_KEYS.collapsed)
    if (cachedCollapsed !== null) {
        collapsed.value = cachedCollapsed === '1'
    }

    const cachedActiveKey = readStorage(STORAGE_KEYS.activeKey)
    if (cachedActiveKey && menuKeys.has(cachedActiveKey)) {
        activeKey.value = cachedActiveKey
        emit('update:currentPage', cachedActiveKey)
    }

    const cachedInstance = readStorage(STORAGE_KEYS.selectedInstance)
    if (cachedInstance && instances.value.some(item => item.value === cachedInstance)) {
        selectedInstance.value = cachedInstance
    }
})

watch(collapsed, (value) => {
    writeStorage(STORAGE_KEYS.collapsed, value ? '1' : '0')
})

watch(activeKey, (value) => {
    writeStorage(STORAGE_KEYS.activeKey, value)
})

watch(selectedInstance, (value) => {
    writeStorage(STORAGE_KEYS.selectedInstance, value)
})
</script>

<template>
    <n-layout-sider bordered collapse-mode="width" :collapsed-width="64" :width="220" :collapsed="collapsed"
        show-trigger class="sidebar" @collapse="handleCollapse" @expand="handleExpand">
        <div class="instance-selector-wrapper">
            <n-dropdown trigger="click" placement="bottom-start" :show-arrow="true" :options="connectionDropdownOptions"
                v-model:show="showConnectionDropdown" :value="selectedInstance" @select="handleConnectionSelect">
                <div class="instance-card" :class="{ collapsed }">
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
                                <div class="instance-card-ip">{{ currentInstance.ip }}</div>
                            </div>
                            <n-icon :size="16" class="instance-card-arrow">
                                <ChevronDownOutline />
                            </n-icon>
                        </div>
                    </template>
                </div>
            </n-dropdown>

        </div>

        <n-menu :value="activeKey" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="20"
            :options="menuOptions" class="main-menu" @update:value="handleSelect" />

        <div class="sidebar-footer" :class="{ collapsed }">
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

.sidebar :deep(.n-layout-sider-scroll-container) {
    display: flex;
    flex-direction: column;
}

.instance-selector-wrapper {
    border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.06));
}

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

.instance-card-info {
    flex: 1;
    min-width: 0;
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
    margin-top: 4px;
    font-size: 12px;
    color: var(--text-secondary, #666);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
    letter-spacing: 0.2px;
    background: var(--chip-bg, rgba(0, 0, 0, 0.04));
    padding: 2px 6px;
    border-radius: 6px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.instance-card-arrow {
    color: #999;
    flex-shrink: 0;
}





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

.menu-option-label {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
}

.menu-option-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.main-menu {
    flex: 1;
    overflow-y: auto;
}

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

.sidebar-footer.collapsed .footer-item:hover::before {
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
}

.footer-label {
    font-size: 14px;
    font-weight: 500;
}

.github-item {
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
</style>
