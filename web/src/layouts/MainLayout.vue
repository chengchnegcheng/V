<template>
  <div class="layout-container">
    <!-- 侧边栏 -->
    <div class="sidebar">
      <div class="logo">V 管理面板</div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <el-menu-item index="/">
          <el-icon><DataBoard /></el-icon>
          <span>控制面板</span>
        </el-menu-item>

        <el-sub-menu index="user-management">
          <template #title>
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </template>
          <el-menu-item index="/users">
            <el-icon><Avatar /></el-icon>
            <span>用户列表</span>
          </el-menu-item>
          <el-menu-item index="/roles">
            <el-icon><Collection /></el-icon>
            <span>角色管理</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="proxy-management">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>代理服务</span>
          </template>
          <el-menu-item index="/proxies">
            <el-icon><Promotion /></el-icon>
            <span>协议管理</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="monitoring">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>监控与统计</span>
          </template>
          <el-menu-item index="/traffic">
            <el-icon><DataAnalysis /></el-icon>
            <span>流量监控</span>
          </el-menu-item>
          <el-menu-item index="/monitor">
            <el-icon><Cpu /></el-icon>
            <span>系统监控</span>
          </el-menu-item>
          <el-menu-item index="/stats">
            <el-icon><Histogram /></el-icon>
            <span>统计分析</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="system-tools">
          <template #title>
            <el-icon><Tools /></el-icon>
            <span>系统工具</span>
          </template>
          <el-menu-item index="/certificates">
            <el-icon><Document /></el-icon>
            <span>证书管理</span>
          </el-menu-item>
          <el-menu-item index="/backups">
            <el-icon><Files /></el-icon>
            <span>备份恢复</span>
          </el-menu-item>
          <el-menu-item index="/logs">
            <el-icon><Notebook /></el-icon>
            <span>日志管理</span>
          </el-menu-item>
          <el-menu-item index="/alerts">
            <el-icon><Bell /></el-icon>
            <span>告警设置</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </div>
    
    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- 顶部导航 -->
      <div class="navbar">
        <div class="left-menu">
          <el-button type="text" @click="toggleSidebar">
            <el-icon><Fold v-if="!sidebarCollapsed" /><Expand v-else /></el-icon>
          </el-button>
        </div>
        <div class="right-menu">
          <el-dropdown trigger="click">
            <span class="user-dropdown">
              <el-avatar size="small" class="avatar">A</el-avatar>
              {{ username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
      
      <!-- 内容 -->
      <div class="content">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import {
  DataBoard, User, Avatar, Collection, Connection, 
  Promotion, Monitor, DataAnalysis, Cpu, Histogram, 
  Tools, Document, Files, Notebook, Bell, Setting,
  Fold, Expand, ArrowDown
} from '@element-plus/icons-vue'

export default {
  name: 'MainLayout',
  components: {
    DataBoard, User, Avatar, Collection, Connection, 
    Promotion, Monitor, DataAnalysis, Cpu, Histogram, 
    Tools, Document, Files, Notebook, Bell, Setting,
    Fold, Expand, ArrowDown
  },
  setup() {
    const router = useRouter()
    const userStore = useUserStore()
    const sidebarCollapsed = ref(false)
    
    const activeMenu = computed(() => {
      return router.currentRoute.value.path
    })
    
    const username = computed(() => {
      return userStore.username || '管理员'
    })
    
    const toggleSidebar = () => {
      sidebarCollapsed.value = !sidebarCollapsed.value
    }
    
    const handleLogout = () => {
      userStore.logout()
      router.push('/login')
    }
    
    return {
      activeMenu,
      sidebarCollapsed,
      username,
      toggleSidebar,
      handleLogout
    }
  }
}
</script>

<style scoped>
.layout-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: 200px;
  background-color: #304156;
  color: white;
  height: 100%;
  overflow-y: auto;
  transition: width 0.3s;
}

.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  font-size: 16px;
  font-weight: bold;
  color: white;
  border-bottom: 1px solid #1f2d3d;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.navbar {
  height: 50px;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background-color: white;
}

.left-menu {
  display: flex;
  align-items: center;
}

.right-menu {
  display: flex;
  align-items: center;
}

.user-dropdown {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.avatar {
  margin-right: 8px;
  background-color: #409EFF;
}

.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background-color: #f5f7f9;
}

:deep(.el-menu-item) {
  display: flex;
  align-items: center;
}

:deep(.el-sub-menu__title) {
  display: flex;
  align-items: center;
}

:deep(.el-menu-item .el-icon),
:deep(.el-sub-menu__title .el-icon) {
  margin-right: 8px;
}
</style> 