<template>
  <div class="system-monitor">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>系统监控</span>
          <el-button type="primary" @click="refreshData">刷新数据</el-button>
        </div>
      </template>
      
      <el-row :gutter="20" class="stats-row">
        <el-col :span="8">
          <el-card class="stats-card">
            <template #header>
              <div class="stats-header">CPU 使用率</div>
            </template>
            <div class="stats-value">
              <el-progress type="dashboard" :percentage="cpuUsage" :color="getColorByPercentage" />
              <div class="stats-details">
                <p>核心数: {{ cpuInfo.cores }}</p>
                <p>型号: {{ cpuInfo.model }}</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="stats-card">
            <template #header>
              <div class="stats-header">内存使用率</div>
            </template>
            <div class="stats-value">
              <el-progress type="dashboard" :percentage="memoryUsage" :color="getColorByPercentage" />
              <div class="stats-details">
                <p>已用: {{ formatBytes(memoryInfo.used) }}</p>
                <p>总计: {{ formatBytes(memoryInfo.total) }}</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="stats-card">
            <template #header>
              <div class="stats-header">磁盘使用率</div>
            </template>
            <div class="stats-value">
              <el-progress type="dashboard" :percentage="diskUsage" :color="getColorByPercentage" />
              <div class="stats-details">
                <p>已用: {{ formatBytes(diskInfo.used) }}</p>
                <p>总计: {{ formatBytes(diskInfo.total) }}</p>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">CPU/内存历史趋势</div>
            </template>
            <div class="chart" ref="resourceChartRef"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">磁盘 I/O</div>
            </template>
            <div class="chart" ref="diskChartRef"></div>
          </el-card>
        </el-col>
      </el-row>
      
      <el-card class="system-info">
        <template #header>
          <div class="card-header">
            <span>系统信息</span>
          </div>
        </template>
        <el-descriptions border :column="2">
          <el-descriptions-item label="操作系统">{{ systemInfo.os }}</el-descriptions-item>
          <el-descriptions-item label="内核版本">{{ systemInfo.kernel }}</el-descriptions-item>
          <el-descriptions-item label="主机名">{{ systemInfo.hostname }}</el-descriptions-item>
          <el-descriptions-item label="运行时间">{{ systemInfo.uptime }}</el-descriptions-item>
          <el-descriptions-item label="负载均衡">{{ systemInfo.load.join(' / ') }}</el-descriptions-item>
          <el-descriptions-item label="IP 地址">{{ systemInfo.ipAddress }}</el-descriptions-item>
        </el-descriptions>
      </el-card>
      
      <el-card class="process-table">
        <template #header>
          <div class="card-header">
            <span>进程列表</span>
            <el-input
              v-model="processSearch"
              placeholder="搜索进程"
              style="width: 200px"
              clearable
            />
          </div>
        </template>
        <el-table :data="filteredProcesses" v-loading="loading" style="width: 100%">
          <el-table-column prop="pid" label="PID" width="80" />
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="user" label="用户" width="100" />
          <el-table-column prop="cpu" label="CPU %" width="80" />
          <el-table-column prop="memory" label="内存 %" width="80" />
          <el-table-column prop="memoryUsed" label="内存使用" width="120">
            <template #default="{ row }">
              {{ formatBytes(row.memoryUsed) }}
            </template>
          </el-table-column>
          <el-table-column prop="started" label="开始时间" width="150" />
          <el-table-column prop="state" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.state)">{{ row.state }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'

// 图表引用
const resourceChartRef = ref(null)
const diskChartRef = ref(null)
let resourceChart = null
let diskChart = null

// 数据状态
const loading = ref(false)
const processSearch = ref('')
const cpuUsage = ref(0)
const memoryUsage = ref(0)
const diskUsage = ref(0)
const cpuInfo = ref({ cores: 0, model: 'Unknown' })
const memoryInfo = ref({ used: 0, total: 1 })
const diskInfo = ref({ used: 0, total: 1 })
const systemInfo = ref({
  os: 'Unknown',
  kernel: 'Unknown',
  hostname: 'Unknown',
  uptime: '0 days, 0 hours, 0 minutes',
  load: [0, 0, 0],
  ipAddress: '0.0.0.0'
})
const processes = ref([])

// 计算属性
const filteredProcesses = computed(() => {
  if (!processSearch.value) return processes.value
  const search = processSearch.value.toLowerCase()
  return processes.value.filter(p => 
    p.name.toLowerCase().includes(search) || 
    p.user.toLowerCase().includes(search) ||
    p.pid.toString().includes(search)
  )
})

// 根据百分比获取颜色
const getColorByPercentage = (percentage) => {
  if (percentage < 60) return '#67C23A'
  if (percentage < 80) return '#E6A23C'
  return '#F56C6C'
}

// 获取进程状态类型
const getStatusType = (state) => {
  const types = {
    running: 'success',
    sleeping: 'info',
    stopped: 'warning',
    zombie: 'danger',
    idle: 'info'
  }
  return types[state.toLowerCase()] || 'info'
}

// 格式化字节大小
const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 初始化图表
const initCharts = () => {
  // CPU/内存历史趋势图表
  resourceChart = echarts.init(resourceChartRef.value)
  resourceChart.setOption({
    title: {
      text: 'CPU/内存使用率趋势'
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        label: {
          backgroundColor: '#6a7985'
        }
      }
    },
    legend: {
      data: ['CPU', '内存']
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: generateTimePoints(10)
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      name: '使用率 (%)'
    },
    series: [
      {
        name: 'CPU',
        type: 'line',
        data: generateRandomData(10, 10, 70),
        areaStyle: {}
      },
      {
        name: '内存',
        type: 'line',
        data: generateRandomData(10, 30, 90),
        areaStyle: {}
      }
    ]
  })
  
  // 磁盘 I/O 图表
  diskChart = echarts.init(diskChartRef.value)
  diskChart.setOption({
    title: {
      text: '磁盘 I/O 活动'
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    legend: {
      data: ['读取', '写入']
    },
    xAxis: {
      type: 'category',
      data: generateTimePoints(7)
    },
    yAxis: {
      type: 'value',
      name: '速率 (MB/s)'
    },
    series: [
      {
        name: '读取',
        type: 'bar',
        data: generateRandomData(7, 0, 50)
      },
      {
        name: '写入',
        type: 'bar',
        data: generateRandomData(7, 5, 70)
      }
    ]
  })
}

// 生成随机数据
const generateRandomData = (count, min, max) => {
  return Array(count).fill(0).map(() => Math.floor(Math.random() * (max - min + 1)) + min)
}

// 生成时间点
const generateTimePoints = (count) => {
  const now = new Date()
  return Array(count).fill(0).map((_, i) => {
    const d = new Date(now - (count - i - 1) * 60 * 1000)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  })
}

// 生成模拟数据
const generateMockData = () => {
  // CPU信息
  cpuUsage.value = Math.floor(Math.random() * 60) + 10
  cpuInfo.value = {
    cores: 8,
    model: 'Intel Core i7-10700K @ 3.80GHz'
  }
  
  // 内存信息
  const totalMem = 16 * 1024 * 1024 * 1024 // 16GB
  const usedMem = totalMem * (Math.random() * 0.5 + 0.2)
  memoryUsage.value = Math.floor((usedMem / totalMem) * 100)
  memoryInfo.value = {
    used: usedMem,
    total: totalMem
  }
  
  // 磁盘信息
  const totalDisk = 500 * 1024 * 1024 * 1024 // 500GB
  const usedDisk = totalDisk * (Math.random() * 0.6 + 0.2)
  diskUsage.value = Math.floor((usedDisk / totalDisk) * 100)
  diskInfo.value = {
    used: usedDisk,
    total: totalDisk
  }
  
  // 系统信息
  systemInfo.value = {
    os: 'Ubuntu 22.04.2 LTS',
    kernel: 'Linux 5.15.0-76-generic',
    hostname: 'v-server',
    uptime: '10 days, 5 hours, 30 minutes',
    load: [0.8, 1.2, 1.5],
    ipAddress: '192.168.1.100'
  }
  
  // 进程信息
  processes.value = Array(20).fill(0).map((_, i) => {
    const states = ['running', 'sleeping', 'stopped', 'idle']
    const names = ['nginx', 'python3', 'node', 'v-core', 'mysqld', 'sshd', 'systemd', 'cron']
    const users = ['root', 'www-data', 'admin', 'system']
    return {
      pid: 1000 + i,
      name: names[Math.floor(Math.random() * names.length)],
      user: users[Math.floor(Math.random() * users.length)],
      cpu: (Math.random() * 5).toFixed(1),
      memory: (Math.random() * 3).toFixed(1),
      memoryUsed: Math.floor(Math.random() * 500) * 1024 * 1024,
      started: '2023-03-15 08:' + Math.floor(Math.random() * 60).toString().padStart(2, '0') + ':00',
      state: states[Math.floor(Math.random() * states.length)]
    }
  }).sort((a, b) => parseFloat(b.cpu) - parseFloat(a.cpu))
}

// 更新图表数据
const updateCharts = () => {
  // 更新CPU/内存图表
  resourceChart.setOption({
    xAxis: {
      data: generateTimePoints(10)
    },
    series: [
      {
        data: generateRandomData(10, 10, 70)
      },
      {
        data: generateRandomData(10, 30, 90)
      }
    ]
  })
  
  // 更新磁盘I/O图表
  diskChart.setOption({
    xAxis: {
      data: generateTimePoints(7)
    },
    series: [
      {
        data: generateRandomData(7, 0, 50)
      },
      {
        data: generateRandomData(7, 5, 70)
      }
    ]
  })
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    // 这里应该是从API获取数据
    // const { data } = await api.getSystemStatus()
    // 使用模拟数据
    generateMockData()
    
    // 更新图表
    updateCharts()
  } catch (error) {
    console.error('获取系统数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 窗口大小变化时重新调整图表大小
const handleResize = () => {
  resourceChart?.resize()
  diskChart?.resize()
}

onMounted(() => {
  // 初始化图表
  initCharts()
  
  // 加载初始数据
  refreshData()
  
  // 开始定时更新
  const timer = setInterval(refreshData, 30000) // 每30秒更新一次
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)
  
  // 清理函数
  onUnmounted(() => {
    clearInterval(timer)
    window.removeEventListener('resize', handleResize)
    resourceChart?.dispose()
    diskChart?.dispose()
  })
})
</script>

<style scoped>
.system-monitor {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats-row {
  margin-bottom: 20px;
}

.stats-card {
  height: 100%;
}

.stats-header {
  font-weight: bold;
  text-align: center;
}

.stats-value {
  text-align: center;
  padding: 20px 0;
}

.stats-details {
  margin-top: 15px;
  text-align: center;
}

.stats-details p {
  margin: 5px 0;
  color: #606266;
}

.chart-card {
  margin-bottom: 20px;
}

.chart-header {
  font-weight: bold;
}

.chart {
  height: 300px;
}

.system-info {
  margin-bottom: 20px;
}

.process-table {
  margin-top: 20px;
}
</style> 