<script setup lang="ts">
import { NLayout, NLayoutContent } from 'naive-ui'
import TitleBar from './components/TitleBar.vue'
import Sidebar from './components/Sidebar.vue'
</script>

<template>
  <div class="app-container">
    <TitleBar />
    <n-layout has-sider class="main-layout">
      <Sidebar />
      <n-layout-content class="content">
        <!-- Main content will go here -->
        <div class="placeholder">
          <h1>Welcome to Rocket Leaf</h1>
          <p>A lightweight RocketMQ client.</p>
          <p>请从左侧菜单选择功能</p>
        </div>
      </n-layout-content>
    </n-layout>
  </div>
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

.content {
  padding: 20px;
  overflow: auto;
  background: var(--bg-color, #f5f5f5);
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
</style>
