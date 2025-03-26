<template>
  <div class="xray-version-switcher">
    <div class="version-info">
      <div class="current-version">
        <span class="label">当前版本:</span>
        <el-tag size="large" type="success">{{ currentVersion }}</el-tag>
      </div>
      
      <div class="version-selector">
        <el-select v-model="selectedVersion" placeholder="选择版本" size="large" style="width: 160px;">
          <el-option
            v-for="version in versions"
            :key="version"
            :label="version"
            :value="version"
          />
        </el-select>
        <el-button 
          type="primary" 
          @click="switchVersion" 
          :disabled="currentVersion === selectedVersion"
          :loading="switching"
        >
          切换版本
        </el-button>
      </div>
    </div>
    
    <div class="controls">
      <el-button 
        type="success" 
        @click="restartService" 
        :loading="restarting"
        icon="Refresh"
      >
        重启服务
      </el-button>
      <el-checkbox v-model="autoUpdate" @change="updateSettings">自动更新</el-checkbox>
    </div>
  </div>
</template>

<script>
import { defineComponent, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

export default defineComponent({
  name: 'XraySimpleManager',
  setup() {
    const currentVersion = ref('v1.8.23')
    const selectedVersion = ref('')
    const versions = ref([
      'v1.8.24',
      'v1.8.23',
      'v1.8.22',
      'v1.8.21',
      'v1.8.20'
    ])
    const autoUpdate = ref(false)
    const switching = ref(false)
    const restarting = ref(false)

    // 初始化
    onMounted(() => {
      fetchVersions()
      selectedVersion.value = currentVersion.value
    })

    // 获取版本列表
    const fetchVersions = async () => {
      try {
        // 实际实现时应该从API获取
        // 这里使用模拟数据
        console.log('获取版本列表')
      } catch (error) {
        console.error('获取版本列表失败:', error)
        ElMessage.error('获取版本列表失败')
      }
    }

    // 切换版本
    const switchVersion = async () => {
      if (!selectedVersion.value || selectedVersion.value === currentVersion.value) {
        return
      }

      switching.value = true
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        currentVersion.value = selectedVersion.value
        ElMessage.success(`已切换到 ${selectedVersion.value}`)
      } catch (error) {
        console.error('版本切换失败:', error)
        ElMessage.error('版本切换失败')
      } finally {
        switching.value = false
      }
    }

    // 重启服务
    const restartService = async () => {
      restarting.value = true
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1500))
        
        ElMessage.success('服务已重启')
      } catch (error) {
        console.error('重启服务失败:', error)
        ElMessage.error('重启服务失败')
      } finally {
        restarting.value = false
      }
    }

    // 更新设置
    const updateSettings = async () => {
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 500))
        
        ElMessage.success(`自动更新已${autoUpdate.value ? '启用' : '禁用'}`)
      } catch (error) {
        console.error('更新设置失败:', error)
        ElMessage.error('更新设置失败')
      }
    }

    return {
      currentVersion,
      selectedVersion,
      versions,
      autoUpdate,
      switching,
      restarting,
      switchVersion,
      restartService,
      updateSettings
    }
  }
})
</script>

<style scoped>
.xray-version-switcher {
  background-color: #f9fafc;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  max-width: 600px;
  margin: 0 auto;
}

.version-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 20px;
}

.current-version {
  display: flex;
  align-items: center;
  gap: 10px;
}

.version-selector {
  display: flex;
  align-items: center;
  gap: 10px;
}

.controls {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.label {
  font-weight: 500;
  min-width: 80px;
}

/* 响应式调整 */
@media (min-width: 768px) {
  .version-info {
    flex-direction: row;
    justify-content: space-between;
  }
}
</style>
