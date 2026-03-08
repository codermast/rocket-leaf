<script setup lang="ts">
// TitleBar.vue
// 跨平台标题栏组件
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { Window } from '@wailsio/runtime'
import { useDialog, NIcon } from 'naive-ui'
import { RemoveOutline, SquareOutline, CopyOutline, CloseOutline } from '@vicons/ionicons5'

// 弹窗确认实例
const dialog = useDialog()

// 定义应用标题
const appTitle = 'Rocket Leaf'

// 平台检测 - 优先使用 Wails v3 运行时
const isMac = ref(false)
const isWindows = ref(false)
const isLinux = ref(false)

// 调试开关：在 macOS 上也展示自定义窗口控制按钮
const forceShowWindowControls = true
const showWindowControls = computed(() => !isMac.value || forceShowWindowControls)

const isMaximised = ref(false)
const eventOffs: Array<() => void> = []

const getWails = () => (window as unknown as { wails?: any; _wails?: any }).wails
const getWailsEnvOS = () => (window as unknown as { _wails?: any })._wails?.environment?.OS as
    | 'darwin'
    | 'windows'
    | 'linux'
    | undefined

const resolvePlatform = () => {
    const wails = getWails()
    if (wails?.System?.IsMac) {
        isMac.value = !!wails.System.IsMac()
        isWindows.value = !!wails.System.IsWindows()
        isLinux.value = !!wails.System.IsLinux()
        return
    }

    const envOS = getWailsEnvOS()
    if (envOS) {
        isMac.value = envOS === 'darwin'
        isWindows.value = envOS === 'windows'
        isLinux.value = envOS === 'linux'
        return
    }

    const platform = navigator.platform.toLowerCase()
    isMac.value = platform.includes('mac')
    isWindows.value = platform.includes('win')
    isLinux.value = platform.includes('linux')
}

const syncMaximisedState = async () => {
    const wails = getWails()
    if (!wails?.Window?.IsMaximised) return
    try {
        isMaximised.value = await wails.Window.IsMaximised()
    } catch {
        // 忽略运行时不可用的情况
    }
}

// 窗口控制函数 (Wails v3 runtime)
const minimizeWindow = () => {
    Window.Minimise()
}

const maximizeWindow = () => {
    Window.ToggleMaximise()
}

const closeWindow = () => {
    dialog.error({
        title: '确认关闭',
        content: '确定要关闭应用吗？',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: () => {
            Window.Close()
        }
    })
}

onMounted(() => {
    resolvePlatform()
    syncMaximisedState()

    const wails = getWails()
    const events = wails?.Events
    const types = wails?.Events?.Types
    if (events?.On && types?.Common) {
        eventOffs.push(events.On(types.Common.WindowMaximise, () => (isMaximised.value = true)))
        eventOffs.push(events.On(types.Common.WindowUnMaximise, () => (isMaximised.value = false)))
        eventOffs.push(events.On(types.Common.WindowRestore, () => (isMaximised.value = false)))
    }
})

onBeforeUnmount(() => {
    eventOffs.forEach(off => off?.())
    eventOffs.length = 0
})
</script>

<template>
    <div class="titlebar" style="--wails-draggable:drag">
        <div class="titlebar-left">
            <!-- macOS 布局: 红绿灯后面跟图标和标题，靠左 -->
            <template v-if="isMac">
                <div class="mac-traffic-light-spacer"></div>
                <div class="title-left">
                    <img src="../assets/leaf-icon.svg" class="app-icon" alt="icon" draggable="false" />
                    <span class="app-title">{{ appTitle }}</span>
                </div>
            </template>

            <!-- Windows/Linux 布局: 图标和标题左侧 -->
            <template v-else>
                <div class="title-left">
                    <img src="../assets/leaf-icon.svg" class="app-icon" alt="icon" draggable="false" />
                    <span class="app-title">{{ appTitle }}</span>
                </div>
            </template>
        </div>

        <!-- 窗口控制按钮：Windows/Linux 及 macOS 调试显示 -->
        <div v-if="showWindowControls" class="window-controls" style="--wails-draggable:no-drag">
            <button class="control-btn min-btn" @click="minimizeWindow" title="最小化">
                <n-icon size="16">
                    <RemoveOutline />
                </n-icon>
            </button>
            <button class="control-btn max-btn" @click="maximizeWindow" :title="isMaximised ? '向下还原' : '最大化'">
                <n-icon size="14" v-if="isMaximised">
                    <CopyOutline />
                </n-icon>
                <n-icon size="14" v-else>
                    <SquareOutline />
                </n-icon>
            </button>
            <button class="control-btn close-btn" @click="closeWindow" title="关闭">
                <n-icon size="16">
                    <CloseOutline />
                </n-icon>
            </button>
        </div>
    </div>
</template>

<style scoped>
.titlebar {
    height: var(--titlebar-height, 30px);
    width: 100%;
    display: flex;
    align-items: center;
    background: var(--titlebar-bg, rgba(255, 255, 255, 0.8));
    user-select: none;
    cursor: default;
    flex-shrink: 0;
    border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.06));
}

.titlebar-left {
    display: flex;
    align-items: center;
    flex: 1;
}

/* macOS 布局 */
.mac-traffic-light-spacer {
    width: 78px;
    /* 红绿灯区域宽度 */
    flex-shrink: 0;
}

.title-center {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
}

/* Windows/Linux 布局 */
.title-left {
    flex: 1;
    display: flex;
    align-items: center;
    padding-left: 12px;
    gap: 8px;
}

.app-icon {
    width: 16px;
    height: 16px;
    pointer-events: none;
    -webkit-user-drag: none;
}

.app-title {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-color, #333333);
    user-select: none;
    -webkit-user-select: none;
    pointer-events: none;
}

/* 窗口控制按钮 */
.window-controls {
    display: flex;
    height: 100%;
}

.control-btn {
    width: 46px;
    height: 100%;
    border: none;
    background: transparent;
    color: var(--text-color, #333);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background 0.15s;
}

.control-btn:hover {
    background: var(--control-hover-bg, rgba(0, 0, 0, 0.08));
}

.close-btn:hover {
    background: #e81123;
    color: white;
}
</style>
