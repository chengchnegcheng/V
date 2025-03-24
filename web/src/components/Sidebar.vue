<template>
  <el-menu
    router
    :default-active="activeMenu"
    class="sidebar-menu"
    :collapse="isCollapse"
  >
    <el-menu-item index="/">
      <el-icon><Odometer /></el-icon>
      <span>仪表盘</span>
    </el-menu-item>
    
    <el-menu-item index="/inbounds">
      <el-icon><Connection /></el-icon>
      <span>入站管理</span>
    </el-menu-item>
    
    <el-menu-item index="/settings">
      <el-icon><Setting /></el-icon>
      <span>系统设置</span>
    </el-menu-item>
    
    <el-menu-item index="/logs">
      <el-icon><DocumentCopy /></el-icon>
      <span>系统日志</span>
    </el-menu-item>
  </el-menu>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { Odometer, Setting, DocumentCopy, Connection } from '@element-plus/icons-vue'

const route = useRoute()
const isCollapse = ref(false)

const activeMenu = computed(() => {
  return route.path
})
</script>

<style scoped>
.sidebar-menu {
  height: 100%;
  border-right: none;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 200px;
}
</style> 