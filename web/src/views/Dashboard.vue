<template>
  <div class="dashboard-container">
    <div class="panel-box">
      <div class="panel-header">
        <span class="panel-title">系统概览</span>
        <el-button type="primary" size="small" @click="refreshStats">刷新</el-button>
      </div>
      <div class="stats-cards">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-card shadow="hover" class="stats-card cpu-card">
              <template #header>
                <div class="card-header">
                  <span>CPU 使用率</span>
                  <el-tag>{{ systemStats.cpu }}%</el-tag>
                </div>
              </template>
              <div class="stats-progress">
                <el-progress type="dashboard" :percentage="systemStats.cpu" :color="getCpuColor"></el-progress>
              </div>
            </el-card>
          </el-col>
          
          <el-col :span="8">
            <el-card shadow="hover" class="stats-card memory-card">
              <template #header>
                <div class="card-header">
                  <span>内存使用率</span>
                  <el-tag>{{ systemStats.memory }}%</el-tag>
                </div>
              </template>
              <div class="stats-progress">
                <el-progress type="dashboard" :percentage="systemStats.memory" :color="getMemoryColor"></el-progress>
              </div>
            </el-card>
          </el-col>
          
          <el-col :span="8">
            <el-card shadow="hover" class="stats-card disk-card">
              <template #header>
                <div class="card-header">
                  <span>磁盘使用率</span>
                  <el-tag>{{ systemStats.disk }}%</el-tag>
                </div>
              </template>
              <div class="stats-progress">
                <el-progress type="dashboard" :percentage="systemStats.disk" :color="getDiskColor"></el-progress>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
    
    <div class="panel-box">
      <div class="panel-header">
        <span class="panel-title">流量统计</span>
        <el-radio-group v-model="trafficPeriod" size="small" @change="changeTrafficPeriod">
          <el-radio-button value="today">今日</el-radio-button>
          <el-radio-button value="week">本周</el-radio-button>
          <el-radio-button value="month">本月</el-radio-button>
        </el-radio-group>
      </div>
      <div class="traffic-stats">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-card shadow="hover" class="traffic-card">
              <template #header>
                <div class="card-header">
                  <span>总流量</span>
                </div>
              </template>
              <div class="traffic-info">
                <div class="traffic-data">
                  <div class="traffic-value">{{ formatTraffic(trafficStats.total) }}</div>
                  <div class="traffic-label">总流量</div>
                </div>
                <div class="traffic-chart">
                  <el-progress type="circle" :percentage="Math.min(trafficStats.percentage, 100)" :width="120"></el-progress>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <el-col :span="12">
            <el-card shadow="hover" class="traffic-card">
              <template #header>
                <div class="card-header">
                  <span>上/下行流量</span>
                </div>
              </template>
              <div class="traffic-details">
                <div class="traffic-item">
                  <span class="traffic-item-label">上行流量</span>
                  <span class="traffic-item-value">{{ formatTraffic(trafficStats.up) }}</span>
                </div>
                <div class="traffic-item">
                  <span class="traffic-item-label">下行流量</span>
                  <span class="traffic-item-value">{{ formatTraffic(trafficStats.down) }}</span>
                </div>
                <div class="traffic-chart-small">
                  <div class="chart-placeholder">
                    <!-- 这里可以放实际的图表组件 -->
                    <div class="up-down-ratio">
                      <div class="up-bar" :style="{ width: getUpPercentage + '%' }"></div>
                      <div class="down-bar" :style="{ width: getDownPercentage + '%' }"></div>
                    </div>
                  </div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
    
    <div class="panel-box">
      <div class="panel-header">
        <span class="panel-title">协议概览</span>
      </div>
      <div class="protocols-stats">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-card shadow="hover" class="protocol-card">
              <template #header>
                <div class="card-header">
                  <span>活跃协议</span>
                </div>
              </template>
              <div class="protocol-list">
                <el-table :data="protocolStats" border style="width: 100%">
                  <el-table-column prop="protocol" label="协议类型">
                    <template #default="scope">
                      <span class="protocol-tag" :class="scope.row.protocol">{{ scope.row.protocol }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="count" label="数量" width="80"></el-table-column>
                  <el-table-column label="状态" width="100">
                    <template #default="scope">
                      <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                        {{ scope.row.status === 'active' ? '运行中' : '已停止' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-card>
          </el-col>
          
          <el-col :span="12">
            <el-card shadow="hover" class="protocol-card">
              <template #header>
                <div class="card-header">
                  <span>流量分布</span>
                </div>
              </template>
              <div class="protocol-chart">
                <div class="chart-placeholder text-center">
                  <div class="traffic-distribution">
                    <div v-for="(item, index) in protocolTraffic" :key="index" class="traffic-bar">
                      <div class="bar-label">{{ item.protocol }}</div>
                      <div class="bar-container">
                        <div class="bar-fill" :style="{ width: item.percentage + '%', backgroundColor: getProtocolColor(item.protocol) }"></div>
                      </div>
                      <div class="bar-value">{{ formatTraffic(item.traffic) }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'

// 系统状态数据
const systemStats = ref({
  cpu: 35,
  memory: 52,
  disk: 68
})

// 流量统计数据
const trafficPeriod = ref('today')
const trafficStats = ref({
  total: 1024 * 1024 * 1024 * 2.5, // 2.5 GB
  up: 1024 * 1024 * 1024 * 0.8,    // 0.8 GB
  down: 1024 * 1024 * 1024 * 1.7,  // 1.7 GB
  limit: 1024 * 1024 * 1024 * 10,  // 10 GB
  percentage: 25 // 25%
})

// 协议统计数据
const protocolStats = ref([
  { protocol: 'vmess', count: 3, status: 'active' },
  { protocol: 'vless', count: 2, status: 'active' },
  { protocol: 'trojan', count: 1, status: 'active' },
  { protocol: 'shadowsocks', count: 4, status: 'active' }
])

// 协议流量分布
const protocolTraffic = ref([
  { protocol: 'vmess', traffic: 1024 * 1024 * 1024 * 1.2, percentage: 45 },
  { protocol: 'vless', traffic: 1024 * 1024 * 1024 * 0.8, percentage: 30 },
  { protocol: 'trojan', traffic: 1024 * 1024 * 1024 * 0.4, percentage: 15 },
  { protocol: 'shadowsocks', traffic: 1024 * 1024 * 1024 * 0.2, percentage: 10 }
])

// 计算上传流量百分比
const getUpPercentage = computed(() => {
  const total = trafficStats.value.up + trafficStats.value.down
  return total > 0 ? Math.round((trafficStats.value.up / total) * 100) : 0
})

// 计算下载流量百分比
const getDownPercentage = computed(() => {
  const total = trafficStats.value.up + trafficStats.value.down
  return total > 0 ? Math.round((trafficStats.value.down / total) * 100) : 0
})

// CPU 颜色
const getCpuColor = computed(() => {
  const cpu = systemStats.value.cpu
  if (cpu < 50) return '#67c23a'
  if (cpu < 80) return '#e6a23c'
  return '#f56c6c'
})

// 内存颜色
const getMemoryColor = computed(() => {
  const memory = systemStats.value.memory
  if (memory < 50) return '#67c23a'
  if (memory < 80) return '#e6a23c'
  return '#f56c6c'
})

// 磁盘颜色
const getDiskColor = computed(() => {
  const disk = systemStats.value.disk
  if (disk < 50) return '#67c23a'
  if (disk < 80) return '#e6a23c'
  return '#f56c6c'
})

// 格式化流量
const formatTraffic = (bytes) => {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 获取协议颜色
const getProtocolColor = (protocol) => {
  switch (protocol) {
    case 'vmess': return '#409eff'
    case 'vless': return '#67c23a'
    case 'trojan': return '#e6a23c'
    case 'shadowsocks': return '#f56c6c'
    default: return '#909399'
  }
}

// 加载数据
const loadData = async () => {
  try {
    // 实际项目中，这里应该调用API获取数据
    // const response = await axios.get('/api/dashboard')
    // systemStats.value = response.data.system
    // trafficStats.value = response.data.traffic
    // protocolStats.value = response.data.protocols
    
    // 模拟数据更新
    systemStats.value = {
      cpu: Math.floor(Math.random() * 60) + 20,
      memory: Math.floor(Math.random() * 70) + 10,
      disk: Math.floor(Math.random() * 50) + 30
    }
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
    ElMessage.error('加载数据失败')
  }
}

// 刷新统计数据
const refreshStats = () => {
  loadData()
  ElMessage.success('数据已刷新')
}

// 切换流量统计周期
const changeTrafficPeriod = (period) => {
  // 实际项目中，应该根据不同周期调用API获取数据
  console.log('Changed traffic period to:', period)
  
  // 模拟数据更新
  if (period === 'today') {
    trafficStats.value.total = 1024 * 1024 * 1024 * 2.5
    trafficStats.value.percentage = 25
  } else if (period === 'week') {
    trafficStats.value.total = 1024 * 1024 * 1024 * 8.2
    trafficStats.value.percentage = 82
  } else {
    trafficStats.value.total = 1024 * 1024 * 1024 * 9.5
    trafficStats.value.percentage = 95
  }
  
  // 更新上下行流量
  trafficStats.value.up = trafficStats.value.total * 0.3
  trafficStats.value.down = trafficStats.value.total * 0.7
}

// 初始化
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard-container {
  padding-bottom: 20px;
}

.panel-box {
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}

.panel-header {
  padding: 15px 20px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.stats-cards,
.traffic-stats,
.protocols-stats {
  padding: 20px;
}

.stats-card {
  height: 240px;
  margin-bottom: 10px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats-progress {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

.traffic-card {
  height: 220px;
  margin-bottom: 10px;
}

.protocol-card {
  height: 320px;
  margin-bottom: 10px;
}

.traffic-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  height: 140px;
}

.traffic-data {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 20px;
}

.traffic-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 10px;
}

.traffic-label {
  font-size: 14px;
  color: #666;
}

.traffic-details {
  padding: 20px;
  height: 140px;
}

.traffic-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 15px;
}

.traffic-item-label {
  color: #666;
}

.traffic-item-value {
  font-weight: bold;
}

.traffic-chart-small {
  margin-top: 20px;
}

.up-down-ratio {
  height: 20px;
  background-color: #f5f7fa;
  border-radius: 10px;
  overflow: hidden;
  display: flex;
}

.up-bar {
  background-color: #409eff;
  height: 100%;
}

.down-bar {
  background-color: #67c23a;
  height: 100%;
}

.protocol-tag {
  display: inline-block;
  padding: 2px 8px;
  font-size: 12px;
  border-radius: 3px;
  color: #fff;
  background-color: #909399;
}

.protocol-tag.vmess {
  background-color: #409eff;
}

.protocol-tag.vless {
  background-color: #67c23a;
}

.protocol-tag.trojan {
  background-color: #e6a23c;
}

.protocol-tag.shadowsocks {
  background-color: #f56c6c;
}

.text-center {
  text-align: center;
}

.traffic-distribution {
  padding: 10px 0;
}

.traffic-bar {
  margin-bottom: 15px;
}

.bar-label {
  font-size: 14px;
  margin-bottom: 5px;
}

.bar-container {
  height: 20px;
  background-color: #f5f7fa;
  border-radius: 10px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 10px;
}

.bar-value {
  text-align: right;
  font-size: 12px;
  margin-top: 2px;
  color: #666;
}
</style> 