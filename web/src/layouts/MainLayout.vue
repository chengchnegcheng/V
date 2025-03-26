<template>
  <div class="app-container">
    <!-- 侧边栏 -->
    <div class="sidebar">
      <div class="logo">
        <h1>V 管理面板</h1>
      </div>
      <el-menu
        class="sidebar-menu"
        background-color="#2b3a4d"
        text-color="#fff"
        active-text-color="#409EFF"
        :default-active="activeMenu"
        :collapse="isCollapse"
        router
      >
        <el-menu-item index="/">
          <el-icon><Monitor /></el-icon>
          <span>控制面板</span>
        </el-menu-item>
        
        <el-menu-item index="/inbounds">
          <el-icon><Connection /></el-icon>
          <span>代理服务</span>
        </el-menu-item>
        
        <el-sub-menu index="user">
          <template #title>
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </template>
          <el-menu-item index="/users">用户列表</el-menu-item>
          <el-menu-item index="/roles">角色管理</el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="monitor">
          <template #title>
            <el-icon><DataAnalysis /></el-icon>
            <span>监控与统计</span>
          </template>
          <el-menu-item index="/system-monitor">系统监控</el-menu-item>
          <el-menu-item index="/traffic-monitor">流量监控</el-menu-item>
          <el-menu-item index="/stats">统计数据</el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="tools">
          <template #title>
            <el-icon><Tools /></el-icon>
            <span>系统工具</span>
          </template>
          <el-menu-item index="/logs">系统日志</el-menu-item>
          <el-menu-item index="/backups">备份恢复</el-menu-item>
          <el-menu-item index="/certificates">证书管理</el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="settings">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </template>
          <el-menu-item index="/settings">配置管理</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </div>
    
    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 顶部栏 -->
      <header class="header">
        <div class="header-left">
          <el-icon
            class="collapse-btn"
            @click="toggleSidebar"
          >
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
        </div>
        <div class="header-right">
          <div class="user-info">
            <el-avatar size="small" class="user-avatar">{{ username.charAt(0).toUpperCase() }}</el-avatar>
            <span class="username">{{ username }}</span>
            <el-button link size="small" @click="goToProfile" class="user-action-btn">个人资料</el-button>
            <el-button link size="small" @click="goToChangePassword" class="user-action-btn">修改密码</el-button>
            <el-button link size="small" @click="confirmLogout" class="user-action-btn logout-btn">退出登录</el-button>
          </div>
        </div>
      </header>
      
      <!-- 内容区域 -->
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  Monitor,
  Connection,
  User,
  DataAnalysis,
  Tools,
  Setting,
  Fold,
  Expand,
  ArrowDown
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isCollapse = ref(false)
const username = computed(() => userStore.user?.username || '管理员')
const activeMenu = computed(() => route.path)

// 切换侧边栏
const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

// 导航到个人资料页面
const goToProfile = () => {
  router.push('/profile')
}

// 导航到修改密码页面
const goToChangePassword = () => {
  router.push('/change-password')
}

// 确认退出登录
const confirmLogout = () => {
  ElMessageBox.confirm('确定要退出登录吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    userStore.logout()
    router.push('/login')
  }).catch(() => {})
}
</script>

<style scoped>
.app-container {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 200px;
  height: 100%;
  background-color: #2b3a4d;
  color: #fff;
  transition: all 0.3s;
  overflow-y: auto;
  flex-shrink: 0;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
}

.sidebar.collapsed {
  width: 64px;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #263445;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.logo h1 {
  font-size: 18px;
  font-weight: 500;
  color: #fff;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
}

.sidebar-menu {
  border-right: none;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 100%;
}

:deep(.el-menu-item) {
  height: 50px;
  line-height: 50px;
  font-size: 14px;
}

:deep(.el-sub-menu__title) {
  height: 50px;
  line-height: 50px;
  font-size: 14px;
}

:deep(.el-menu-item.is-active) {
  background-color: #1890ff !important;
  color: white !important;
}

:deep(.el-menu-item:hover) {
  background-color: #263445 !important;
}

:deep(.el-sub-menu__title:hover) {
  background-color: #263445 !important;
}

:deep(.el-sub-menu .el-menu-item) {
  min-width: 200px;
  padding-left: 40px !important;
  background-color: #1f2d3d;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  height: 60px;
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left, .header-right {
  display: flex;
  align-items: center;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.collapse-btn:hover {
  color: #409EFF;
}

/* User dropdown styles */
.user-dropdown {
  height: 40px;
}

.dropdown-link {
  display: flex;
  align-items: center;
  cursor: pointer;
  font-size: 14px;
  color: #333;
  height: 40px;
  padding: 0 10px;
  border-radius: 4px;
  transition: background-color 0.3s;
  border: 1px solid transparent;
}

.dropdown-link:hover {
  background-color: #f5f7fa;
  border-color: #e6e6e6;
}

.username {
  margin: 0 8px;
  font-weight: 500;
}

.user-avatar {
  background-color: #1890ff !important;
  color: white !important;
  font-weight: bold;
}

.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background-color: #f0f2f5;
}

:deep(.el-dropdown-menu) {
  min-width: 130px;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

:deep(.el-dropdown-menu__item) {
  padding: 10px 16px;
  font-size: 14px;
  line-height: 20px;
  display: flex;
  align-items: center;
}

:deep(.el-dropdown-menu__item:not(.is-disabled):hover) {
  background-color: #f5f7fa;
  color: #409EFF;
}

:deep(.el-dropdown-menu__item--divided) {
  margin-top: 6px;
  border-top: 1px solid #ebeef5;
}

:deep(.el-dropdown-menu__item--divided:before) {
  height: 6px;
  margin-top: -6px;
}

:deep(.el-popper) {
  z-index: 9999 !important;
}

.user-info {
  display: flex;
  align-items: center;
  height: 40px;
  padding: 0 5px;
}

.user-avatar {
  background-color: #1890ff !important;
  color: white !important;
  font-weight: bold;
}

.username {
  margin: 0 8px;
  font-weight: 500;
}

.user-action-btn {
  margin: 0 2px;
  font-size: 12px;
  color: #606266;
}

.user-action-btn:hover {
  color: #409EFF;
}

.logout-btn {
  color: #f56c6c;
}

.logout-btn:hover {
  color: #f56c6c;
  opacity: 0.8;
}
</style> 