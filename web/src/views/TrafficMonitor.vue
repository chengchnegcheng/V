<template>
  <div class="page-container">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stats-card">
          <template #header>
            <div class="card-header">
              <span>总流量</span>
              <el-tag type="info">{{ formatBytes(totalTraffic) }}</el-tag>
            </div>
          </template>
          <div class="stats-value">
            <div class="value">{{ formatBytes(currentTraffic) }}/s</div>
            <div class="label">当前速率</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <template #header>
            <div class="card-header">
              <span>上传流量</span>
              <el-tag type="success">{{ formatBytes(totalUpload) }}</el-tag>
            </div>
          </template>
          <div class="stats-value">
            <div class="value">{{ formatBytes(currentUpload) }}/s</div>
            <div class="label">当前速率</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <template #header>
            <div class="card-header">
              <span>下载流量</span>
              <el-tag type="warning">{{ formatBytes(totalDownload) }}</el-tag>
            </div>
          </template>
          <div class="stats-value">
            <div class="value">{{ formatBytes(currentDownload) }}/s</div>
            <div class="label">当前速率</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <template #header>
            <div class="card-header">
              <span>连接数</span>
              <el-tag type="info">{{ connections }}</el-tag>
            </div>
          </template>
          <div class="stats-value">
            <div class="value">{{ activeConnections }}</div>
            <div class="label">活跃连接</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>流量趋势</span>
              <el-radio-group v-model="timeRange" size="small">
                <el-radio-button label="1h">1小时</el-radio-button>
                <el-radio-button label="24h">24小时</el-radio-button>
                <el-radio-button label="7d">7天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container" ref="trafficChartRef"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>协议分布</span>
            </div>
          </template>
          <div class="chart-container" ref="protocolChartRef"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="connection-card">
      <template #header>
        <div class="card-header">
          <span>实时连接</span>
          <el-button type="danger" size="small" @click="handleClearConnections">
            清除连接
          </el-button>
        </div>
      </template>
      <el-table :data="connectionList" v-loading="loading" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user" label="用户" />
        <el-table-column prop="protocol" label="协议">
          <template #default="{ row }">
            <el-tag :type="getProtocolType(row.protocol)">
              {{ row.protocol.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remote" label="远程地址" />
        <el-table-column prop="startTime" label="开始时间" />
        <el-table-column prop="duration" label="持续时间" />
        <el-table-column prop="upload" label="上传">
          <template #default="{ row }">
            {{ formatBytes(row.upload) }}
          </template>
        </el-table-column>
        <el-table-column prop="download" label="下载">
          <template #default="{ row }">
            {{ formatBytes(row.download) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button 
              size="small" 
              type="danger"
              @click="handleKillConnection(row)"
            >
              断开
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { monitorApi } from '@/api'
import * as echarts from 'echarts'

export default {
  name: 'TrafficMonitor',
  setup() {
    const loading = ref(false)
    const timeRange = ref('1h')
    const trafficChartRef = ref(null)
    const protocolChartRef = ref(null)
    let trafficChart = null
    let protocolChart = null
    let updateTimer = null

    // 统计数据
    const totalTraffic = ref(0)
    const currentTraffic = ref(0)
    const totalUpload = ref(0)
    const currentUpload = ref(0)
    const totalDownload = ref(0)
    const currentDownload = ref(0)
    const connections = ref(0)
    const activeConnections = ref(0)
    const connectionList = ref([])

    const getProtocolType = (protocol) => {
      const types = {
        vmess: 'primary',
        vless: 'success',
        trojan: 'warning',
        shadowsocks: 'info'
      }
      return types[protocol] || 'info'
    }

    const formatBytes = (bytes) => {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    const formatDuration = (seconds) => {
      const hours = Math.floor(seconds / 3600)
      const minutes = Math.floor((seconds % 3600) / 60)
      const remainingSeconds = seconds % 60

      if (hours > 0) {
        return `${hours}小时${minutes}分${remainingSeconds}秒`
      } else if (minutes > 0) {
        return `${minutes}分${remainingSeconds}秒`
      } else {
        return `${remainingSeconds}秒`
      }
    }

    const fetchStats = async () => {
      try {
        const response = await monitorApi.getNetworkStats()
        const { stats, connections } = response

        // 更新统计数据
        totalTraffic.value = stats.totalTraffic
        currentTraffic.value = stats.currentTraffic
        totalUpload.value = stats.totalUpload
        currentUpload.value = stats.currentUpload
        totalDownload.value = stats.totalDownload
        currentDownload.value = stats.currentDownload
        connections.value = stats.connections
        activeConnections.value = stats.activeConnections

        // 更新连接列表
        connectionList.value = connections.map(conn => ({
          ...conn,
          duration: formatDuration(conn.duration)
        }))

        // 更新图表
        updateTrafficChart(stats.trafficHistory)
        updateProtocolChart(stats.protocolDistribution)
      } catch (error) {
        console.error('获取统计数据失败:', error)
      }
    }

    const updateTrafficChart = (data) => {
      if (!trafficChartRef.value) return

      if (!trafficChart) {
        trafficChart = echarts.init(trafficChartRef.value)
      }

      const option = {
        tooltip: {
          trigger: 'axis',
          formatter: (params) => {
            const date = new Date(params[0].axisValue)
            return `${date.toLocaleString()}<br/>
                    ${params[0].seriesName}: ${formatBytes(params[0].value)}<br/>
                    ${params[1].seriesName}: ${formatBytes(params[1].value)}`
          }
        },
        legend: {
          data: ['上传', '下载']
        },
        xAxis: {
          type: 'category',
          data: data.times
        },
        yAxis: {
          type: 'value',
          axisLabel: {
            formatter: (value) => formatBytes(value)
          }
        },
        series: [
          {
            name: '上传',
            type: 'line',
            data: data.upload,
            smooth: true
          },
          {
            name: '下载',
            type: 'line',
            data: data.download,
            smooth: true
          }
        ]
      }

      trafficChart.setOption(option)
    }

    const updateProtocolChart = (data) => {
      if (!protocolChartRef.value) return

      if (!protocolChart) {
        protocolChart = echarts.init(protocolChartRef.value)
      }

      const option = {
        tooltip: {
          trigger: 'item',
          formatter: (params) => {
            return `${params.name}: ${formatBytes(params.value)} (${params.percent}%)`
          }
        },
        legend: {
          orient: 'vertical',
          left: 'left'
        },
        series: [
          {
            type: 'pie',
            radius: '50%',
            data: data,
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)'
              }
            }
          }
        ]
      }

      protocolChart.setOption(option)
    }

    const handleClearConnections = async () => {
      try {
        await ElMessageBox.confirm(
          '确定要清除所有连接吗？',
          '警告',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        await monitorApi.clearConnections()
        ElMessage.success('清除连接成功')
        fetchStats()
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage.error('清除连接失败')
        }
      }
    }

    const handleKillConnection = async (connection) => {
      try {
        await ElMessageBox.confirm(
          '确定要断开该连接吗？',
          '警告',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        await monitorApi.killConnection(connection.id)
        ElMessage.success('断开连接成功')
        fetchStats()
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage.error('断开连接失败')
        }
      }
    }

    onMounted(() => {
      fetchStats()
      updateTimer = setInterval(fetchStats, 5000) // 每5秒更新一次
    })

    onUnmounted(() => {
      if (updateTimer) {
        clearInterval(updateTimer)
      }
      if (trafficChart) {
        trafficChart.dispose()
      }
      if (protocolChart) {
        protocolChart.dispose()
      }
    })

    return {
      loading,
      timeRange,
      trafficChartRef,
      protocolChartRef,
      totalTraffic,
      currentTraffic,
      totalUpload,
      currentUpload,
      totalDownload,
      currentDownload,
      connections,
      activeConnections,
      connectionList,
      getProtocolType,
      formatBytes,
      handleClearConnections,
      handleKillConnection
    }
  }
}
</script>

<style scoped>
.page-container {
  padding: 20px;
}

.stats-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats-value {
  text-align: center;
  padding: 20px 0;
}

.stats-value .value {
  font-size: 24px;
  font-weight: bold;
  color: #409EFF;
}

.stats-value .label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.chart-row {
  margin-bottom: 20px;
}

.chart-container {
  height: 300px;
}

.connection-card {
  margin-top: 20px;
}
</style> 