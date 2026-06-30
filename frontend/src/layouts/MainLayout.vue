<template>
  <el-container class="layout-container">
    <!-- Sidebar -->
    <el-aside width="220px" class="sidebar">
      <div class="sidebar-header">
        <div class="logo-icon">✨</div>
        <span class="logo-text">Shiny Collection</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
        background-color="#1a1a2e"
        text-color="#a0aec0"
        active-text-color="#f6e05e"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/records">
          <el-icon><List /></el-icon>
          <span>狩猎记录</span>
        </el-menu-item>
        <el-menu-item index="/records/create">
          <el-icon><Plus /></el-icon>
          <span>新增记录</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- Main Content -->
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.meta.title">{{ route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <span class="header-tag">✨ {{ stats?.totalShiny || 0 }} 只异色</span>
        </div>
      </el-header>

      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { DataAnalysis, List, Plus } from '@element-plus/icons-vue'
import { useRecordStore } from '@/stores/record'

const route = useRoute()
const store = useRecordStore()

const stats = computed(() => store.stats)
const activeMenu = computed(() => route.path)

onMounted(() => {
  store.fetchStats()
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #1a1a2e;
  overflow-y: auto;
}

.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #f6e05e;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #2d2d4e;
}

.logo-icon {
  font-size: 24px;
}

.sidebar-menu {
  border-right: none !important;
}

.header {
  background: #fff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 24px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-tag {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 4px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.main-content {
  background-color: #f5f7fa;
  padding: 24px;
  overflow-y: auto;
}
</style>
