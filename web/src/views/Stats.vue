<template>
  <div class="stats-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>流量统计</span>
              <el-radio-group v-model="timeRange" size="small">
                <el-radio-button label="day">24小时</el-radio-button>
                <el-radio-button label="week">7天</el-radio-button>
                <el-radio-button label="month">30天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="trafficChartOption" autoresize />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="mt-20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>用户流量排行</span>
            </div>
          </template>
          <el-table :data="userStats" style="width: 100%">
            <el-table-column prop="username" label="用户名" />
            <el-table-column prop="upload" label="上传流量">
              <template #default="scope">
                {{ formatBytes(scope.row.upload) }}
              </template>
            </el-table-column>
            <el-table-column prop="download" label="下载流量">
              <template #default="scope">
                {{ formatBytes(scope.row.download) }}
              </template>
            </el-table-column>
            <el-table-column prop="total" label="总流量">
              <template #default="scope">
                {{ formatBytes(scope.row.total) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>协议流量分布</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="protocolChartOption" autoresize />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="mt-20">
      <template #header>
        <div class="card-header">
          <span>详细统计</span>
          <el-button type="primary" size="small" @click="exportStats">导出数据</el-button>
        </div>
      </template>
      <el-table :data="detailedStats" style="width: 100%">
        <el-table-column prop="date" label="日期" width="180" />
        <el-table-column prop="username" label="用户名" width="180" />
        <el-table-column prop="protocol" label="协议" width="120" />
        <el-table-column prop="upload" label="上传流量">
          <template #default="scope">
            {{ formatBytes(scope.row.upload) }}
          </template>
        </el-table-column>
        <el-table-column prop="download" label="下载流量">
          <template #default="scope">
            {{ formatBytes(scope.row.download) }}
          </template>
        </el-table-column>
        <el-table-column prop="total" label="总流量">
          <template #default="scope">
            {{ formatBytes(scope.row.total) }}
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer,
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent
])

export default {
  name: 'Stats',
  components: {
    VChart
  },
  setup() {
    const timeRange = ref('day')
    const userStats = ref([])
    const protocolStats = ref([])
    const detailedStats = ref([])
    const currentPage = ref(1)
    const pageSize = ref(20)
    const total = ref(0)
    const trafficData = ref([])

    const trafficChartOption = computed(() => ({
      tooltip: {
        trigger: 'axis',
        formatter: (params) => {
          const date = params[0].axisValue
          const upload = formatBytes(params[0].data)
          const download = formatBytes(params[1].data)
          return `${date}<br/>上传：${upload}<br/>下载：${download}`
        }
      },
      legend: {
        data: ['上传', '下载']
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: trafficData.value.map(item => item.date)
      },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: value => formatBytes(value)
        }
      },
      series: [
        {
          name: '上传',
          type: 'line',
          data: trafficData.value.map(item => item.upload)
        },
        {
          name: '下载',
          type: 'line',
          data: trafficData.value.map(item => item.download)
        }
      ]
    }))

    const protocolChartOption = computed(() => ({
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      series: [
        {
          type: 'pie',
          radius: '50%',
          data: protocolStats.value.map(item => ({
            name: item.protocol,
            value: item.total
          })),
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }
      ]
    }))

    const formatBytes = (bytes) => {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    const fetchTrafficData = async () => {
      try {
        const response = await fetch(`/api/stats/traffic?range=${timeRange.value}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (!response.ok) {
          throw new Error('获取流量数据失败')
        }
        
        trafficData.value = await response.json()
      } catch (error) {
        console.error('Failed to fetch traffic data:', error)
      }
    }

    const fetchUserStats = async () => {
      try {
        const response = await fetch('/api/stats/users', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (!response.ok) {
          throw new Error('获取用户统计失败')
        }
        
        userStats.value = await response.json()
      } catch (error) {
        console.error('Failed to fetch user stats:', error)
      }
    }

    const fetchProtocolStats = async () => {
      try {
        const response = await fetch('/api/stats/protocols', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (!response.ok) {
          throw new Error('获取协议统计失败')
        }
        
        protocolStats.value = await response.json()
      } catch (error) {
        console.error('Failed to fetch protocol stats:', error)
      }
    }

    const fetchDetailedStats = async () => {
      try {
        const response = await fetch(`/api/stats/detailed?page=${currentPage.value}&page_size=${pageSize.value}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (!response.ok) {
          throw new Error('获取详细统计失败')
        }
        
        const data = await response.json()
        detailedStats.value = data.items
        total.value = data.total
      } catch (error) {
        console.error('Failed to fetch detailed stats:', error)
      }
    }

    const handleSizeChange = (val) => {
      pageSize.value = val
      fetchDetailedStats()
    }

    const handleCurrentChange = (val) => {
      currentPage.value = val
      fetchDetailedStats()
    }

    const exportStats = async () => {
      try {
        const response = await fetch('/api/stats/export', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
        
        if (!response.ok) {
          throw new Error('导出数据失败')
        }
        
        const blob = await response.blob()
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `stats_${new Date().toISOString().split('T')[0]}.csv`
        document.body.appendChild(a)
        a.click()
        window.URL.revokeObjectURL(url)
        document.body.removeChild(a)
      } catch (error) {
        console.error('Failed to export stats:', error)
      }
    }

    let timer

    onMounted(() => {
      fetchTrafficData()
      fetchUserStats()
      fetchProtocolStats()
      fetchDetailedStats()
      
      // 每5分钟更新一次数据
      timer = setInterval(() => {
        fetchTrafficData()
        fetchUserStats()
        fetchProtocolStats()
        fetchDetailedStats()
      }, 300000)
    })

    onUnmounted(() => {
      if (timer) {
        clearInterval(timer)
      }
    })

    return {
      timeRange,
      userStats,
      protocolStats,
      detailedStats,
      currentPage,
      pageSize,
      total,
      trafficChartOption,
      protocolChartOption,
      formatBytes,
      handleSizeChange,
      handleCurrentChange,
      exportStats
    }
  }
}
</script>

<style scoped>
.stats-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  height: 300px;
}

.mt-20 {
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style> 