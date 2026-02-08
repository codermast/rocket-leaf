<script setup lang="ts">
// Sidebar.vue
// 使用 Naive UI 和 Ionicons5 图标构建的左侧边栏
import { ref, h, Component } from 'vue'
import { NMenu, NIcon, NLayoutSider } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import {
    SpeedometerOutline,
    LinkOutline,
    AddCircleOutline,
    ListOutline,
    FolderOutline,
    CreateOutline,
    PeopleOutline,
    ChatbubblesOutline,
    ServerOutline,
    SettingsOutline
} from '@vicons/ionicons5'

// 渲染图标的辅助函数
const renderIcon = (icon: Component) => {
    return () => h(NIcon, null, { default: () => h(icon) })
}

// 侧边栏是否折叠
const collapsed = ref(false)

// 当前选中的菜单项
const activeKey = ref<string>('dashboard')

// 菜单选项
const menuOptions: MenuOption[] = [
    {
        label: '仪表盘',
        key: 'dashboard',
        icon: renderIcon(SpeedometerOutline)
    },
    {
        label: '连接管理',
        key: 'connections',
        icon: renderIcon(LinkOutline),
        children: [
            {
                label: '新建连接',
                key: 'new-connection',
                icon: renderIcon(AddCircleOutline)
            },
            {
                label: '连接列表',
                key: 'connection-list',
                icon: renderIcon(ListOutline)
            }
        ]
    },
    {
        label: 'Topic 管理',
        key: 'topics',
        icon: renderIcon(FolderOutline),
        children: [
            {
                label: 'Topic 列表',
                key: 'topic-list',
                icon: renderIcon(ListOutline)
            },
            {
                label: '创建 Topic',
                key: 'create-topic',
                icon: renderIcon(CreateOutline)
            }
        ]
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
    },
    {
        label: '设置',
        key: 'settings',
        icon: renderIcon(SettingsOutline)
    }
]

// 处理菜单选择
const handleSelect = (key: string) => {
    activeKey.value = key
    console.log('Selected:', key)
}
</script>

<template>
    <n-layout-sider bordered collapse-mode="width" :collapsed-width="64" :width="220" :collapsed="collapsed"
        show-trigger @collapse="collapsed = true" @expand="collapsed = false" class="sidebar">
        <n-menu :value="activeKey" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="20"
            :options="menuOptions" @update:value="handleSelect" />
    </n-layout-sider>
</template>

<style scoped>
.sidebar {
    height: 100%;
    background: var(--bg-color, #ffffff);
}
</style>
