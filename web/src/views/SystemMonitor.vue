<template>
  <div class="monitor-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>系统监控</span>
          <div class="header-actions">
            <el-button type="primary" @click="refreshData">刷新数据</el-button>
            <el-button type="success" @click="handleConfigureAlerts">配置告警</el-button>
          </div>
        </div>
      </template>

      <!-- 性能指标 -->
      <el-row :gutter="20" class="metrics-row">
        <el-col :span="6">
          <el-card class="metric-card" shadow="hover">
            <template #header>
              <div class="metric-header">
                <span>CPU 使用率</span>
                <el-tag :type="getCpuStatusType(cpuUsage)">{{ cpuUsage }}%</el-tag>
              </div>
            </template>
            <div class="metric-chart" ref="cpuChartRef"></div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="metric-card" shadow="hover">
            <template #header>
              <div class="metric-header">
                <span>内存使用率</span>
                <el-tag :type="getMemoryStatusType(memoryUsage)">{{ memoryUsage }}%</el-tag>
              </div>
            </template>
            <div class="metric-chart" ref="memoryChartRef"></div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="metric-card" shadow="hover">
            <template #header>
              <div class="metric-header">
                <span>磁盘使用率</span>
                <el-tag :type="getDiskStatusType(diskUsage)">{{ diskUsage }}%</el-tag>
              </div>
            </template>
            <div class="metric-chart" ref="diskChartRef"></div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="metric-card" shadow="hover">
            <template #header>
              <div class="metric-header">
                <span>网络流量</span>
                <el-tag type="info">{{ formatNetworkSpeed(networkSpeed) }}</el-tag>
              </div>
            </template>
            <div class="metric-chart" ref="networkChartRef"></div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 进程管理 -->
      <el-card class="process-card">
        <template #header>
          <div class="card-header">
            <span>进程管理</span>
            <div class="header-actions">
              <el-input
                v-model="processSearchQuery"
                placeholder="搜索进程"
                style="width: 200px"
                clearable
              />
              <el-button type="primary" @click="refreshProcesses">刷新</el-button>
            </div>
          </div>
        </template>
        <el-table
          :data="filteredProcesses"
          border
          style="width: 100%"
          v-loading="processLoading"
        >
          <el-table-column prop="pid" label="PID" width="80" />
          <el-table-column prop="name" label="进程名" width="150" />
          <el-table-column prop="user" label="用户" width="120" />
          <el-table-column prop="cpu" label="CPU" width="100">
            <template #default="scope">
              {{ scope.row.cpu }}%
            </template>
          </el-table-column>
          <el-table-column prop="memory" label="内存" width="100">
            <template #default="scope">
              {{ formatMemory(scope.row.memory) }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getProcessStatusType(scope.row.status)">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="scope">
              <el-button
                type="danger"
                size="small"
                @click="handleTerminateProcess(scope.row)"
              >
                终止
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="processCurrentPage"
            v-model:page-size="processPageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="processTotal"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleProcessSizeChange"
            @current-change="handleProcessCurrentChange"
          />
        </div>
      </el-card>

      <!-- 网络连接 -->
      <el-card class="network-card">
        <template #header>
          <div class="card-header">
            <span>网络连接</span>
            <div class="header-actions">
              <el-input
                v-model="connectionSearchQuery"
                placeholder="搜索连接"
                style="width: 200px"
                clearable
              />
              <el-button type="primary" @click="refreshConnections">刷新</el-button>
              <el-button type="danger" @click="handleClearConnections">清除连接</el-button>
            </div>
          </div>
        </template>
        <el-table
          :data="filteredConnections"
          border
          style="width: 100%"
          v-loading="connectionLoading"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="localAddress" label="本地地址" width="180" />
          <el-table-column prop="remoteAddress" label="远程地址" width="180" />
          <el-table-column prop="protocol" label="协议" width="100" />
          <el-table-column prop="state" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getConnectionStateType(scope.row.state)">
                {{ scope.row.state }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="scope">
              <el-button
                type="danger"
                size="small"
                @click="handleDisconnect(scope.row)"
              >
                断开
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="connectionCurrentPage"
            v-model:page-size="connectionPageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="connectionTotal"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleConnectionSizeChange"
            @current-change="handleConnectionCurrentChange"
          />
        </div>
      </el-card>
    </el-card>

    <!-- 告警配置对话框 -->
    <el-dialog
      v-model="alertDialogVisible"
      title="告警配置"
      width="50%"
    >
      <el-form
        ref="alertFormRef"
        :model="alertForm"
        :rules="alertRules"
        label-width="120px"
      >
        <el-form-item label="CPU 告警阈值" prop="cpuThreshold">
          <el-input-number
            v-model="alertForm.cpuThreshold"
            :min="0"
            :max="100"
            :step="5"
          />
          <span class="form-text">%</span>
        </el-form-item>
        <el-form-item label="内存告警阈值" prop="memoryThreshold">
          <el-input-number
            v-model="alertForm.memoryThreshold"
            :min="0"
            :max="100"
            :step="5"
          />
          <span class="form-text">%</span>
        </el-form-item>
        <el-form-item label="磁盘告警阈值" prop="diskThreshold">
          <el-input-number
            v-model="alertForm.diskThreshold"
            :min="0"
            :max="100"
            :step="5"
          />
          <span class="form-text">%</span>
        </el-form-item>
        <el-form-item label="告警通知方式">
          <el-checkbox-group v-model="alertForm.notificationMethods">
            <el-checkbox label="email">邮件通知</el-checkbox>
            <el-checkbox label="webhook">Webhook</el-checkbox>
            <el-checkbox label="sms">短信通知</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="告警间隔" prop="alertInterval">
          <el-input-number
            v-model="alertForm.alertInterval"
            :min="1"
            :max="60"
            :step="1"
          />
          <span class="form-text">分钟</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="alertDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmAlertConfig">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as echarts from 'echarts'
import { monitorApi } from '@/api'

// 性能指标数据
const cpuUsage = ref(0)
const memoryUsage = ref(0)
const diskUsage = ref(0)
const networkSpeed = ref(0)

// 图表实例
const cpuChartRef = ref(null)
const memoryChartRef = ref(null)
const diskChartRef = ref(null)
const networkChartRef = ref(null)
let cpuChart = null
let memoryChart = null
let diskChart = null
let networkChart = null

// 进程列表数据
const processes = ref([])
const processLoading = ref(false)
const processSearchQuery = ref('')
const processCurrentPage = ref(1)
const processPageSize = ref(20)
const processTotal = ref(0)

// 网络连接数据
const connections = ref([])
const connectionLoading = ref(false)
const connectionSearchQuery = ref('')
const connectionCurrentPage = ref(1)
const connectionPageSize = ref(20)
const connectionTotal = ref(0)

// 告警配置
const alertDialogVisible = ref(false)
const alertFormRef = ref(null)
const alertForm = ref({
  cpuThreshold: 80,
  memoryThreshold: 80,
  diskThreshold: 80,
  notificationMethods: ['email'],
  alertInterval: 5
})

// 告警规则
const alertRules = {
  cpuThreshold: [
    { required: true, message: '请输入 CPU 告警阈值', trigger: 'blur' }
  ],
  memoryThreshold: [
    { required: true, message: '请输入内存告警阈值', trigger: 'blur' }
  ],
  diskThreshold: [
    { required: true, message: '请输入磁盘告警阈值', trigger: 'blur' }
  ],
  alertInterval: [
    { required: true, message: '请输入告警间隔', trigger: 'blur' }
  ]
}

// 过滤后的进程列表
const filteredProcesses = computed(() => {
  if (!processSearchQuery.value) return processes.value
  const query = processSearchQuery.value.toLowerCase()
  return processes.value.filter(process =>
    process.name.toLowerCase().includes(query) ||
    process.user.toLowerCase().includes(query)
  )
})

// 过滤后的连接列表
const filteredConnections = computed(() => {
  if (!connectionSearchQuery.value) return connections.value
  const query = connectionSearchQuery.value.toLowerCase()
  return connections.value.filter(connection =>
    connection.localAddress.toLowerCase().includes(query) ||
    connection.remoteAddress.toLowerCase().includes(query) ||
    connection.protocol.toLowerCase().includes(query)
  )
})

// 初始化图表
const initCharts = () => {
  // CPU 使用率图表
  cpuChart = echarts.init(cpuChartRef.value)
  cpuChart.setOption({
    title: { text: 'CPU 使用率趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', max: 100 },
    series: [{
      type: 'line',
      data: [],
      smooth: true,
      areaStyle: {}
    }]
  })

  // 内存使用率图表
  memoryChart = echarts.init(memoryChartRef.value)
  memoryChart.setOption({
    title: { text: '内存使用率趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', max: 100 },
    series: [{
      type: 'line',
      data: [],
      smooth: true,
      areaStyle: {}
    }]
  })

  // 磁盘使用率图表
  diskChart = echarts.init(diskChartRef.value)
  diskChart.setOption({
    title: { text: '磁盘使用率趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', max: 100 },
    series: [{
      type: 'line',
      data: [],
      smooth: true,
      areaStyle: {}
    }]
  })

  // 网络流量图表
  networkChart = echarts.init(networkChartRef.value)
  networkChart.setOption({
    title: { text: '网络流量趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value' },
    series: [{
      type: 'line',
      data: [],
      smooth: true,
      areaStyle: {}
    }]
  })
}

// 更新图表数据
const updateCharts = (data) => {
  const now = new Date()
  const cpuData = cpuChart.getOption().series[0].data
  const memoryData = memoryChart.getOption().series[0].data
  const diskData = diskChart.getOption().series[0].data
  const networkData = networkChart.getOption().series[0].data

  // 保持最近 30 个数据点
  if (cpuData.length > 30) cpuData.shift()
  if (memoryData.length > 30) memoryData.shift()
  if (diskData.length > 30) diskData.shift()
  if (networkData.length > 30) networkData.shift()

  // 添加新数据
  cpuData.push([now, data.cpu])
  memoryData.push([now, data.memory])
  diskData.push([now, data.disk])
  networkData.push([now, data.network])

  // 更新图表
  cpuChart.setOption({
    series: [{ data: cpuData }]
  })
  memoryChart.setOption({
    series: [{ data: memoryData }]
  })
  diskChart.setOption({
    series: [{ data: diskData }]
  })
  networkChart.setOption({
    series: [{ data: networkData }]
  })
}

// 获取性能指标数据
const getMetrics = async () => {
  try {
    const data = await monitorApi.getMetrics()
    cpuUsage.value = data.cpu
    memoryUsage.value = data.memory
    diskUsage.value = data.disk
    networkSpeed.value = data.network
    updateCharts(data)
  } catch (error) {
    ElMessage.error('获取性能指标失败')
  }
}

// 获取进程列表
const getProcesses = async () => {
  processLoading.value = true
  try {
    const data = await monitorApi.getProcesses()
    processes.value = data.processes
    processTotal.value = data.total
  } catch (error) {
    ElMessage.error('获取进程列表失败')
  } finally {
    processLoading.value = false
  }
}

// 获取网络连接
const getConnections = async () => {
  connectionLoading.value = true
  try {
    const data = await monitorApi.getConnections()
    connections.value = data.connections
    connectionTotal.value = data.total
  } catch (error) {
    ElMessage.error('获取网络连接失败')
  } finally {
    connectionLoading.value = false
  }
}

// 获取告警配置
const getAlertConfig = async () => {
  try {
    const data = await monitorApi.getAlertConfig()
    alertForm.value = data
  } catch (error) {
    ElMessage.error('获取告警配置失败')
  }
}

// 刷新所有数据
const refreshData = () => {
  getMetrics()
  getProcesses()
  getConnections()
}

// 刷新进程列表
const refreshProcesses = () => {
  getProcesses()
}

// 刷新网络连接
const refreshConnections = () => {
  getConnections()
}

// 终止进程
const handleTerminateProcess = async (process) => {
  try {
    await ElMessageBox.confirm(
      `确定要终止进程 ${process.name} (PID: ${process.pid}) 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await monitorApi.terminateProcess(process.pid)
    ElMessage.success('进程已终止')
    getProcesses()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('终止进程失败')
    }
  }
}

// 断开连接
const handleDisconnect = async (connection) => {
  try {
    await ElMessageBox.confirm(
      '确定要断开此连接吗？',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await monitorApi.disconnect(connection.id)
    ElMessage.success('连接已断开')
    getConnections()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('断开连接失败')
    }
  }
}

// 清除所有连接
const handleClearConnections = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清除所有连接吗？此操作不可恢复。',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await monitorApi.clearConnections()
    ElMessage.success('所有连接已清除')
    getConnections()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('清除连接失败')
    }
  }
}

// 配置告警
const handleConfigureAlerts = () => {
  alertDialogVisible.value = true
}

// 确认告警配置
const confirmAlertConfig = async () => {
  if (!alertFormRef.value) return
  await alertFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        await monitorApi.updateAlertConfig(alertForm.value)
        ElMessage.success('告警配置已更新')
        alertDialogVisible.value = false
      } catch (error) {
        ElMessage.error('更新告警配置失败')
      }
    }
  })
}

// 进程分页处理
const handleProcessSizeChange = (val) => {
  processPageSize.value = val
  getProcesses()
}

const handleProcessCurrentChange = (val) => {
  processCurrentPage.value = val
  getProcesses()
}

// 连接分页处理
const handleConnectionSizeChange = (val) => {
  connectionPageSize.value = val
  getConnections()
}

const handleConnectionCurrentChange = (val) => {
  connectionCurrentPage.value = val
  getConnections()
}

// 工具函数
const getCpuStatusType = (usage) => {
  if (usage >= 90) return 'danger'
  if (usage >= 70) return 'warning'
  return 'success'
}

const getMemoryStatusType = (usage) => {
  if (usage >= 90) return 'danger'
  if (usage >= 70) return 'warning'
  return 'success'
}

const getDiskStatusType = (usage) => {
  if (usage >= 90) return 'danger'
  if (usage >= 70) return 'warning'
  return 'success'
}

const getProcessStatusType = (status) => {
  const types = {
    'running': 'success',
    'sleeping': 'info',
    'stopped': 'warning',
    'zombie': 'danger'
  }
  return types[status] || 'info'
}

const getConnectionStateType = (state) => {
  const types = {
    'established': 'success',
    'time_wait': 'warning',
    'close_wait': 'info',
    'fin_wait': 'info',
    'listening': 'primary'
  }
  return types[state] || 'info'
}

const formatMemory = (bytes) => {
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(2)} ${units[unitIndex]}`
}

const formatNetworkSpeed = (bytesPerSecond) => {
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  let speed = bytesPerSecond
  let unitIndex = 0
  while (speed >= 1024 && unitIndex < units.length - 1) {
    speed /= 1024
    unitIndex++
  }
  return `${speed.toFixed(2)} ${units[unitIndex]}`
}

// 定时刷新数据
let metricsTimer = null

onMounted(() => {
  initCharts()
  refreshData()
  getAlertConfig()
  // 每 5 秒刷新一次性能指标
  metricsTimer = setInterval(getMetrics, 5000)
})

onUnmounted(() => {
  if (metricsTimer) {
    clearInterval(metricsTimer)
  }
  if (cpuChart) cpuChart.dispose()
  if (memoryChart) memoryChart.dispose()
  if (diskChart) diskChart.dispose()
  if (networkChart) networkChart.dispose()
})
</script>

<style scoped>
.monitor-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.metrics-row {
  margin-bottom: 20px;
}

.metric-card {
  height: 300px;
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.metric-chart {
  height: 220px;
}

.process-card,
.network-card {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.form-text {
  margin-left: 10px;
  color: #909399;
}
</style> 