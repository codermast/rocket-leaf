<script setup lang="ts">
import { ref, h, computed, onMounted, onUnmounted, watch } from 'vue'
import type { Component } from 'vue'
import {
    NMenu, NIcon, NLayoutSider, NTooltip, NDropdown, NAvatar, NText, NBadge,
    NCard, NPopover, NTag, NModal, NForm, NFormItem, NInput, NInputNumber,
    NSelect, NSpace, NButton, NSwitch, NAlert, useMessage
} from 'naive-ui'
import type { MenuOption, FormInst, FormRules } from 'naive-ui'
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
import * as ConnectionService from '../../bindings/rocket-leaf/internal/service/connectionservice'
import { addConnectionCompat } from '../utils/connectionServiceCompat'
import { emitConnectionsChanged, onConnectionsChanged } from '../utils/connectionEvents'

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

const props = defineProps<{
    currentPage?: string
}>()

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

// Apache RocketMQ 官方 SVG 图标组件
const RocketMQIcon = {
    render() {
        return h('svg', {
            viewBox: '0 0 24 24',
            xmlns: 'http://www.w3.org/2000/svg',
            fill: 'currentColor',
            width: '1em',
            height: '1em'
        }, [
            h('path', {
                d: 'M11.438 23.467c-.517-.638-1.106-1.89-1.217-2.587l-.082-.511h1.835c1.435 0 1.835.036 1.835.165 0 .352-.412 1.553-.709 2.066-.333.577-1.021 1.41-1.155 1.4-.043-.004-.272-.244-.507-.533zm-4.532-4.193c-1.251-3.005-1.231-6.784.056-10.63.786-2.35 2.652-5.689 4.413-7.9L11.967 0l.422.493c.763.893 2.612 3.731 3.28 5.036 1.32 2.578 2.055 4.993 2.264 7.438.197 2.302-.176 4.837-.962 6.533l-.338.731-.727-.433-.727-.433H11.95c-2.466 0-3.287.039-3.476.166-.136.091-.453.29-.705.441l-.458.276-.405-.974zm9.338-1.79c.779-2.623.532-6.253-.635-9.344-.683-1.81-2.085-4.319-3.211-5.747-.357-.452-.387-.466-.609-.265-.441.398-1.854 2.622-2.544 4.002-1.927 3.856-2.484 7.995-1.521 11.308l.196.672h8.138l.186-.626zM3.311 19.835c.037-.155.108-.565.157-.909.079-.549.189-.729.885-1.443l.795-.815.002.507c.003.641.302 1.799.631 2.445l.254.498H4.64c-1.384-.001-1.396-.003-1.329-.283zm14.944-.376c.271-.613.529-1.616.606-2.352.031-.299.066-.282.762.379s.738.735.908 1.631c.098.516.179.952.179.97 0 .017-.618.031-1.373.031h-1.373l.291-.659zm-6.477-4.504a2.173 2.173 0 0 1-2.17-2.17c0-1.196.973-2.17 2.17-2.17s2.17.973 2.17 2.17-.973 2.17-2.17 2.17zm0-3.865c-.935 0-1.696.761-1.696 1.695s.761 1.696 1.696 1.696c.935 0 1.696-.761 1.696-1.696s-.761-1.695-1.696-1.695zM9.455 9.457a.657.657 0 1 1 0 1.314.657.657 0 0 1 0-1.314zm-.357 4.665a.8.8 0 1 1 0 1.6.8.8 0 0 1 0-1.6zm5.212-5.18a1.069 1.069 0 1 1 0 2.138 1.069 1.069 0 0 1 0-2.138zm0 5.75a1.418 1.418 0 1 1 0 2.836 1.418 1.418 0 0 1 0-2.836zM9.447 10.68l.491-.491.729.729-.491.491-.729-.729zm4.066-.336l.539.539-.729.729-.539-.539.729-.729zm-3.572 3.362l.491.491-.729.729-.491-.491.729-.729zm2.721 1.064l.61-.59.779.754-.61.59-.779-.754zm-1.717-2.167a.277.277 0 1 1 0 .554.277.277 0 0 1 0-.554zm.794 0a.277.277 0 1 1 0 .554.277.277 0 0 1 0-.554zm.794 0a.277.277 0 1 1 0 .554.277.277 0 0 1 0-.554z'
            })
        ])
    }
}

const collapsed = ref(false)
const activeKey = ref<string>('dashboard')
const showConnectionDropdown = ref(false)
const message = useMessage()

// 添加新连接对话框
interface ConnectionFormModel {
    name: string
    env: string
    nameServer: string
    timeoutSec: number
    enableACL: boolean
    accessKey: string
    secretKey: string
    remark: string
}

const showAddConnectionModal = ref(false)
const isSubmittingConnection = ref(false)
const connectionFormRef = ref<FormInst | null>(null)

const createDefaultConnectionForm = (): ConnectionFormModel => ({
    name: '',
    env: '开发',
    nameServer: '',
    timeoutSec: 5,
    enableACL: false,
    accessKey: '',
    secretKey: '',
    remark: ''
})

const connectionFormModel = ref<ConnectionFormModel>(createDefaultConnectionForm())

const envOptions = [
    { label: '生产', value: '生产' },
    { label: '测试', value: '测试' },
    { label: '开发', value: '开发' }
]

const connectionFormRules: FormRules = {
    name: [{ required: true, message: '请输入连接名称', trigger: ['blur', 'input'] }],
    env: [{ required: true, message: '请选择环境', trigger: ['change'] }],
    nameServer: [{ required: true, message: '请输入 NameServer 地址', trigger: ['blur', 'input'] }],
    timeoutSec: [{ required: true, type: 'number', message: '请设置超时时间', trigger: ['blur', 'change'] }],
    accessKey: [{
        validator: (_rule, value: string) => {
            if (connectionFormModel.value.enableACL && !value.trim()) {
                return new Error('启用 ACL 时 AccessKey 不能为空')
            }
            return true
        },
        trigger: ['blur', 'input']
    }],
    secretKey: [{
        validator: (_rule, value: string) => {
            if (connectionFormModel.value.enableACL && !value.trim()) {
                return new Error('启用 ACL 时 SecretKey 不能为空')
            }
            return true
        },
        trigger: ['blur', 'input']
    }]
}

const handleACLToggle = (enabled: boolean) => {
    connectionFormModel.value.enableACL = enabled
    if (!enabled) {
        connectionFormModel.value.accessKey = ''
        connectionFormModel.value.secretKey = ''
    }
}

const fallbackInstance: Instance = {
    label: '未连接实例',
    value: 'fallback',
    ip: '-',
    status: 'offline',
    icon: PeopleOutline
}

const instances = ref<Instance[]>([])

const selectedInstance = ref<string>('')

// 加载连接列表
const loadConnections = async (preferredInstance?: string) => {
    try {
        const connections = await ConnectionService.GetConnections()
        instances.value = connections
            .filter(conn => conn !== null)
            .map(conn => ({
                label: conn!.name,
                value: String(conn!.id),
                ip: conn!.nameServer,
                status: conn!.status === 'online' ? 'online' : 'offline' as 'online' | 'offline',
                icon: ServerOutline
            }))

        const expectedSelected = preferredInstance ?? selectedInstance.value
        if (expectedSelected && instances.value.some(item => item.value === expectedSelected)) {
            selectedInstance.value = expectedSelected
            return
        }

        const defaultConn = connections.find(c => c?.isDefault)
        if (defaultConn && defaultConn.id !== undefined) {
            selectedInstance.value = String(defaultConn.id)
        } else if (instances.value.length > 0 && instances.value[0]) {
            selectedInstance.value = instances.value[0].value
        } else {
            selectedInstance.value = ''
        }
    } catch (err) {
        console.error('加载连接列表失败:', err)
    }
}

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
    showConnectionDropdown.value = false
    connectionFormModel.value = createDefaultConnectionForm()
    showAddConnectionModal.value = true
}

// 保存新连接
let removeConnectionsChangedListener: (() => void) | null = null

const saveNewConnection = async () => {
    if (!connectionFormRef.value) return
    await connectionFormRef.value.validate()
    isSubmittingConnection.value = true

    try {
        const form = connectionFormModel.value
        const newConn = await addConnectionCompat({
            ...form,
            accessKey: form.accessKey.trim(),
            secretKey: form.secretKey.trim()
        })
        const newConnId = newConn?.id
        await loadConnections(newConnId !== undefined ? String(newConnId) : undefined)
        emitConnectionsChanged()
        message.success('连接创建成功')
        showAddConnectionModal.value = false
    } catch (err) {
        console.error('创建连接失败:', err)
        const errorMessage = err instanceof Error ? err.message : '创建连接失败'
        message.error(errorMessage)
    } finally {
        isSubmittingConnection.value = false
    }
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
                    style: 'margin-right: 12px; background: #22c372; color: #fff;'
                },
                {
                    default: () => h(NIcon, { size: 16 }, { default: () => h(RocketMQIcon) })
                }
            ),
            h('div', null, [
                h('div', null, [h(NText, { depth: 2 }, { default: () => '集群列表' })]),
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

onMounted(async () => {
    removeConnectionsChangedListener = onConnectionsChanged(() => {
        loadConnections()
    })

    // 加载连接列表
    await loadConnections()

    const cachedCollapsed = readStorage(STORAGE_KEYS.collapsed)
    if (cachedCollapsed !== null) {
        collapsed.value = cachedCollapsed === '1'
    }

    const propActiveKey = props.currentPage
    if (propActiveKey && menuKeys.has(propActiveKey)) {
        activeKey.value = propActiveKey
    } else {
        const cachedActiveKey = readStorage(STORAGE_KEYS.activeKey)
        if (cachedActiveKey && menuKeys.has(cachedActiveKey)) {
            activeKey.value = cachedActiveKey
            emit('update:currentPage', cachedActiveKey)
        }
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

watch(() => props.currentPage, (value) => {
    if (!value || !menuKeys.has(value)) return
    if (activeKey.value !== value) {
        activeKey.value = value
    }
})

onUnmounted(() => {
    removeConnectionsChangedListener?.()
    removeConnectionsChangedListener = null
})
</script>

<template>
    <n-layout-sider bordered collapse-mode="width" :collapsed-width="64" :width="220" :collapsed="collapsed"
        show-trigger class="sidebar" @collapse="handleCollapse" @expand="handleExpand">
        <div class="instance-selector-wrapper">
            <div class="instance-card-list">
                <n-dropdown trigger="click" placement="right-start" :show-arrow="true"
                    :options="connectionDropdownOptions" v-model:show="showConnectionDropdown" :value="selectedInstance"
                    @select="handleConnectionSelect">
                    <div class="instance-trigger">
                        <n-popover v-if="collapsed" placement="right" trigger="hover">
                            <template #trigger>
                                <div class="instance-popover-trigger">
                                    <div class="instance-card collapsed">
                                        <n-badge dot :type="currentInstance.status === 'online' ? 'success' : 'error'">
                                            <n-avatar>MQ</n-avatar>
                                        </n-badge>
                                    </div>
                                </div>
                            </template>
                            <div>
                                <div style="font-weight: 600;">{{ currentInstance.label }} <n-tag
                                        :type="currentInstance.status === 'online' ? 'success' : 'error'" size="tiny">
                                        {{ currentInstance.status === 'online' ? '在线' : '离线' }} </n-tag>
                                </div>
                                <div style="font-size: 12px; opacity: 0.78; margin-top: 4px;">
                                    {{ currentInstance.ip }}
                                </div>
                            </div>
                        </n-popover>
                        <n-card v-else class="instance-card-wrapper" :bordered="false" size="small">
                            <div class="instance-card-content">
                                <n-icon :size="20">
                                    <RocketMQIcon />
                                </n-icon>
                                <div class="instance-card-info">
                                    <div class="instance-card-name">
                                        {{ currentInstance.label }}
                                        <n-tag :type="currentInstance.status === 'online' ? 'success' : 'error'"
                                            size="tiny" round :bordered="false">
                                            {{ currentInstance.status === 'online' ? '在线' : '离线' }}
                                        </n-tag>
                                    </div>
                                    <div class="instance-card-ip">{{ currentInstance.ip }}</div>
                                </div>
                                <n-icon :size="16" class="instance-card-arrow"
                                    :class="{ expanded: showConnectionDropdown }">
                                    <ChevronDownOutline />
                                </n-icon>
                            </div>
                        </n-card>
                    </div>
                </n-dropdown>
            </div>
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

    <!-- 快速添加连接对话框 -->
    <n-modal v-model:show="showAddConnectionModal" preset="card" title="添加新连接" style="width: 620px"
        :mask-closable="false">
        <n-form ref="connectionFormRef" class="quick-connection-form" :model="connectionFormModel"
            :rules="connectionFormRules" label-placement="left" label-width="120px">
            <n-form-item label="连接名称" path="name">
                <n-input v-model:value="connectionFormModel.name" placeholder="如：生产集群" />
            </n-form-item>
            <n-form-item label="环境" path="env">
                <n-select v-model:value="connectionFormModel.env" :options="envOptions" placeholder="请选择环境" />
            </n-form-item>
            <n-form-item label="NameServer" path="nameServer">
                <n-input v-model:value="connectionFormModel.nameServer" placeholder="如：192.168.1.100:9876" />
            </n-form-item>
            <n-form-item label="超时时间" path="timeoutSec">
                <n-input-number v-model:value="connectionFormModel.timeoutSec" :min="1" :max="60" placeholder="秒"
                    style="width: 100%" />
            </n-form-item>

            <n-form-item label="启用 ACL">
                <n-switch :value="connectionFormModel.enableACL" @update:value="handleACLToggle">
                    <template #checked>已启用</template>
                    <template #unchecked>未启用</template>
                </n-switch>
            </n-form-item>

            <n-alert v-if="connectionFormModel.enableACL" type="info" :show-icon="false" style="margin-bottom: 12px;">
                已启用 ACL 鉴权，请填写 AccessKey 与 SecretKey。
            </n-alert>

            <n-form-item v-if="connectionFormModel.enableACL" label="AccessKey" path="accessKey">
                <n-input v-model:value="connectionFormModel.accessKey" placeholder="请输入 ACL AccessKey" />
            </n-form-item>

            <n-form-item v-if="connectionFormModel.enableACL" label="SecretKey" path="secretKey">
                <n-input v-model:value="connectionFormModel.secretKey" type="password" show-password-on="click"
                    placeholder="请输入 ACL SecretKey" />
            </n-form-item>

            <n-form-item label="备注" path="remark">
                <n-input v-model:value="connectionFormModel.remark" type="textarea" placeholder="连接备注信息（可选）"
                    :rows="2" />
            </n-form-item>
        </n-form>

        <template #footer>
            <n-space justify="end">
                <n-button @click="showAddConnectionModal = false">取消</n-button>
                <n-button type="primary" :loading="isSubmittingConnection" @click="saveNewConnection">确定</n-button>
            </n-space>
        </template>
    </n-modal>
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

.instance-trigger {
    width: 100%;
}

.instance-popover-trigger {
    width: 100%;
}

/* 展开状态 - 使用 n-card 原生 hover 效果 */
.instance-card-wrapper {
    margin: auto;
    border-radius: 1px;
}

/* 折叠状态 - 保持原有图标样式 */
.instance-card.collapsed {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    padding: 0;
    height: var(--menu-item-height, 42px);
    margin: 6px 0;
    background: transparent;
    position: relative;
    cursor: pointer;
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


.instance-card.collapsed:hover .instance-card-icon-only {
    transform: translateY(-1px);
    box-shadow: 0 8px 18px rgba(24, 160, 88, 0.3);
    filter: saturate(1.06) brightness(1.03);
}

/* RocketMQ 官方图标已使用内联 SVG，无需额外样式 */

.instance-card-content {
    display: flex;
    align-items: center;
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
    font-size: 10px;
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
    transition: transform 0.2s ease;
    transform-origin: center;
}

.instance-card-arrow.expanded {
    transform: rotate(-90deg);
}

.instance-card-list {
    margin-top: 5px;
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

.quick-connection-form :deep(.n-form-item-label) {
    white-space: nowrap;
    word-break: keep-all;
}
</style>
