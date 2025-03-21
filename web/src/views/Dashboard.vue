<template>
  <div class="dashboard-container">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>总用户数</span>
              <el-icon><User /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ stats.totalUsers }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>在线用户</span>
              <el-icon><UserFilled /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ stats.onlineUsers }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>总流量</span>
              <el-icon><DataLine /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ formatBytes(stats.totalTraffic) }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>系统负载</span>
              <el-icon><Monitor /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ stats.systemLoad }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>流量趋势</span>
            </div>
          </template>
          <div class="chart-container">
            <!-- TODO: 添加流量趋势图表 -->
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>协议分布</span>
            </div>
          </template>
          <div class="chart-container">
            <!-- TODO: 添加协议分布图表 -->
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="recent-activities">
      <template #header>
        <div class="card-header">
          <span>最近活动</span>
        </div>
      </template>
      <el-table :data="activities" style="width: 100%">
        <el-table-column prop="time" label="时间" width="180" />
        <el-table-column prop="user" label="用户" width="180" />
        <el-table-column prop="action" label="操作" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'success' ? 'success' : 'danger'">
              {{ scope.row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { User, UserFilled, DataLine, Monitor } from '@element-plus/icons-vue'

export default {
  name: 'Dashboard',
  components: {
    User,
    UserFilled,
    DataLine,
    Monitor
  },
  setup() {
    const stats = ref({
      totalUsers: 0,
      onlineUsers: 0,
      totalTraffic: 0,
      systemLoad: 0
    })

    const activities = ref([])

    const formatBytes = (bytes) => {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    const fetchDashboardData = async () => {
      try {
        // TODO: 调用仪表盘数据 API
        const response = await fetch('/api/dashboard/stats', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (!response.ok) {
          throw new Error('获取数据失败')
        }
        
        const data = await response.json()
        stats.value = data.stats
        activities.value = data.activities
      } catch (error) {
        console.error('Failed to fetch dashboard data:', error)
      }
    }

    onMounted(() => {
      fetchDashboardData()
      // 每分钟更新一次数据
      setInterval(fetchDashboardData, 60000)
    })

    return {
      stats,
      activities,
      formatBytes
    }
  }
}
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}

.stat-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  text-align: center;
  margin-top: 10px;
}

.chart-row {
  margin-top: 20px;
}

.chart-container {
  height: 300px;
}

.recent-activities {
  margin-top: 20px;
}
</style> 