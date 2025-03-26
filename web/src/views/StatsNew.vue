<template>
  <div class="stats-container">
    <div class="page-header">
      <h1>统计分析</h1>
      <div class="header-actions">
        <el-button type="primary" @click="refreshData">
          <el-icon><Refresh /></el-icon> 刷新数据
        </el-button>
        <el-select v-model="timeRange" placeholder="选择时间范围" @change="handleTimeRangeChange">
          <el-option label="今日" value="today"></el-option>
          <el-option label="本周" value="week"></el-option>
          <el-option label="本月" value="month"></el-option>
          <el-option label="全部" value="all"></el-option>
        </el-select>
      </div>
    </div>

    <!-- 数据概览卡片 -->
    <el-row :gutter="20" class="stats-cards">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-card-icon system">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="stats-card-info">
              <div class="stats-card-title">系统总运行时间</div>
              <div class="stats-card-value">{{ stats.uptime || '暂无数据' }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-card-icon traffic">
              <el-icon><DataLine /></el-icon>
            </div>
            <div class="stats-card-info">
              <div class="stats-card-title">总流量</div>
              <div class="stats-card-value">{{ formatBytes(stats.totalTraffic) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-card-icon users">
              <el-icon><User /></el-icon>
            </div>
            <div class="stats-card-info">
              <div class="stats-card-title">总用户数</div>
              <div class="stats-card-value">{{ stats.totalUsers || 0 }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-card-icon active">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="stats-card-info">
              <div class="stats-card-title">活跃连接数</div>
              <div class="stats-card-value">{{ stats.activeConnections || 0 }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 流量统计图表 -->
    <el-card shadow="never" class="chart-card" v-loading="loading.traffic">
      <template #header>
        <div class="card-header">
          <span>流量统计</span>
          <el-radio-group v-model="trafficChartType" size="small" @change="updateTrafficChart">
            <el-radio-button value="hour">小时</el-radio-button>
            <el-radio-button value="day">天</el-radio-button>
            <el-radio-button value="month">月</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div ref="trafficChartRef" class="chart-container"></div>
    </el-card>

    <!-- 用户活跃度和协议使用情况图表 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card shadow="never" class="chart-card" v-loading="loading.users">
          <template #header>
            <div class="card-header">
              <span>用户活跃度</span>
            </div>
          </template>
          <div ref="userActivityChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="chart-card" v-loading="loading.protocols">
          <template #header>
            <div class="card-header">
              <span>协议使用情况</span>
            </div>
          </template>
          <div ref="protocolChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 系统资源使用情况 -->
    <el-card shadow="never" class="chart-card" v-loading="loading.system">
      <template #header>
        <div class="card-header">
          <span>系统资源使用情况</span>
          <el-switch
            v-model="realTimeMonitor"
            active-text="实时监控"
            inactive-text="静态数据"
            @change="toggleRealTimeMonitor"
          />
        </div>
      </template>
      <div ref="systemResourceChartRef" class="chart-container"></div>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import { ElMessage } from 'element-plus';
import { Monitor, DataLine, User, Connection, Refresh } from '@element-plus/icons-vue';
import axios from 'axios';
import * as echarts from 'echarts';

export default {
  name: 'Stats',
  components: {
    Monitor, DataLine, User, Connection, Refresh
  },
  setup() {
    // 图表实例
    const trafficChartRef = ref(null);
    const userActivityChartRef = ref(null);
    const protocolChartRef = ref(null);
    const systemResourceChartRef = ref(null);
    let trafficChart = null;
    let userActivityChart = null;
    let protocolChart = null;
    let systemResourceChart = null;
    
    // 数据状态
    const timeRange = ref('today');
    const trafficChartType = ref('hour');
    const realTimeMonitor = ref(false);
    const loading = reactive({
      traffic: false,
      users: false,
      protocols: false,
      system: false
    });
    const stats = reactive({
      uptime: '0天0小时0分钟',
      totalTraffic: 0,
      totalUsers: 0,
      activeConnections: 0,
      trafficData: [],
      userActivityData: [],
      protocolData: [],
      systemResourceData: []
    });
    
    // 定时器
    let refreshTimer = null;
    
    // 格式化字节数
    const formatBytes = (bytes) => {
      if (bytes === 0 || bytes === undefined) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };
    
    // 加载统计数据
    const loadStats = async () => {
      try {
        // 这里应该调用实际API，但现在使用模拟数据
        // const response = await axios.get(`/api/stats?timeRange=${timeRange.value}`);
        // const data = response.data;
        
        // 使用模拟数据
        setTimeout(() => {
          stats.uptime = '3天5小时27分钟';
          stats.totalTraffic = 1024 * 1024 * 1024 * 5.7; // 5.7 GB
          stats.totalUsers = 12;
          stats.activeConnections = 5;
          
          // 初始化各个图表
          initTrafficChart();
          initUserActivityChart();
          initProtocolChart();
          initSystemResourceChart();
        }, 500);
      } catch (error) {
        console.error('获取统计数据失败:', error);
        ElMessage.error('获取统计数据失败');
      }
    };
    
    // 刷新数据
    const refreshData = () => {
      loading.traffic = true;
      loading.users = true;
      loading.protocols = true;
      loading.system = true;
      
      setTimeout(() => {
        loadStats();
        loading.traffic = false;
        loading.users = false;
        loading.protocols = false;
        loading.system = false;
        ElMessage.success('数据刷新成功');
      }, 1000);
    };
    
    // 时间范围变更
    const handleTimeRangeChange = () => {
      refreshData();
    };
    
    // 更新流量图表
    const updateTrafficChart = () => {
      loading.traffic = true;
      
      setTimeout(() => {
        initTrafficChart();
        loading.traffic = false;
      }, 500);
    };
    
    // 切换实时监控
    const toggleRealTimeMonitor = () => {
      if (realTimeMonitor.value) {
        ElMessage.info('已开启实时监控，数据将每5秒自动刷新');
        refreshTimer = setInterval(updateSystemResourceChart, 5000);
      } else {
        ElMessage.info('已关闭实时监控');
        if (refreshTimer) {
          clearInterval(refreshTimer);
          refreshTimer = null;
        }
      }
    };
    
    // 更新系统资源图表
    const updateSystemResourceChart = () => {
      // 模拟数据更新
      const newCpuData = Math.floor(Math.random() * 40) + 20; // 20-60%
      const newMemoryData = Math.floor(Math.random() * 30) + 40; // 40-70%
      const newDiskData = Math.floor(Math.random() * 20) + 10; // 10-30%
      
      systemResourceChart.setOption({
        series: [
          { 
            name: 'CPU使用率', 
            data: [newCpuData] 
          },
          { 
            name: '内存使用率', 
            data: [newMemoryData] 
          },
          { 
            name: '磁盘使用率', 
            data: [newDiskData] 
          }
        ]
      });
    };
    
    // 初始化流量图表
    const initTrafficChart = () => {
      if (trafficChart) {
        trafficChart.dispose();
      }
      
      const chartDom = trafficChartRef.value;
      trafficChart = echarts.init(chartDom);
      
      // 生成模拟数据
      const generateTimeData = () => {
        const now = new Date();
        const result = [];
        
        if (trafficChartType.value === 'hour') {
          // 最近24小时的数据
          for (let i = 23; i >= 0; i--) {
            const time = new Date(now);
            time.setHours(now.getHours() - i);
            result.push(time.getHours() + ':00');
          }
        } else if (trafficChartType.value === 'day') {
          // 最近7天的数据
          for (let i = 6; i >= 0; i--) {
            const time = new Date(now);
            time.setDate(now.getDate() - i);
            result.push((time.getMonth() + 1) + '/' + time.getDate());
          }
        } else {
          // 最近6个月的数据
          for (let i = 5; i >= 0; i--) {
            const time = new Date(now);
            time.setMonth(now.getMonth() - i);
            result.push((time.getMonth() + 1) + '月');
          }
        }
        
        return result;
      };
      
      const generateTrafficData = (max) => {
        const result = [];
        for (let i = 0; i < (trafficChartType.value === 'hour' ? 24 : (trafficChartType.value === 'day' ? 7 : 6)); i++) {
          result.push(Math.floor(Math.random() * max));
        }
        return result;
      };
      
      const timeData = generateTimeData();
      const uploadData = generateTrafficData(100);
      const downloadData = generateTrafficData(200);
      
      const option = {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'shadow'
          },
          formatter: function(params) {
            let tooltip = params[0].axisValue + '<br/>';
            params.forEach(param => {
              tooltip += param.seriesName + ': ' + formatBytes(param.value * 1024 * 1024) + '<br/>';
            });
            return tooltip;
          }
        },
        legend: {
          data: ['上传流量', '下载流量']
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true
        },
        xAxis: {
          type: 'category',
          data: timeData
        },
        yAxis: {
          type: 'value',
          axisLabel: {
            formatter: function(value) {
              return formatBytes(value * 1024 * 1024);
            }
          }
        },
        series: [
          {
            name: '上传流量',
            type: 'bar',
            stack: 'total',
            data: uploadData,
            itemStyle: {
              color: '#4fc08d'
            }
          },
          {
            name: '下载流量',
            type: 'bar',
            stack: 'total',
            data: downloadData,
            itemStyle: {
              color: '#409EFF'
            }
          }
        ]
      };
      
      trafficChart.setOption(option);
      
      // 监听窗口变化，更新图表大小
      window.addEventListener('resize', () => {
        trafficChart.resize();
      });
    };
    
    // 初始化用户活跃度图表
    const initUserActivityChart = () => {
      if (userActivityChart) {
        userActivityChart.dispose();
      }
      
      const chartDom = userActivityChartRef.value;
      userActivityChart = echarts.init(chartDom);
      
      // 生成模拟数据
      const generateTimeData = () => {
        const result = [];
        for (let i = 0; i < 7; i++) {
          const date = new Date();
          date.setDate(date.getDate() - i);
          result.unshift((date.getMonth() + 1) + '/' + date.getDate());
        }
        return result;
      };
      
      const generateUserData = () => {
        const result = [];
        for (let i = 0; i < 7; i++) {
          result.push(Math.floor(Math.random() * 10) + 1);
        }
        return result;
      };
      
      const timeData = generateTimeData();
      const userData = generateUserData();
      
      const option = {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'line'
          }
        },
        xAxis: {
          type: 'category',
          data: timeData
        },
        yAxis: {
          type: 'value',
          minInterval: 1
        },
        series: [
          {
            name: '活跃用户数',
            type: 'line',
            data: userData,
            smooth: true,
            symbol: 'circle',
            symbolSize: 8,
            lineStyle: {
              width: 3,
              color: '#F56C6C'
            },
            itemStyle: {
              color: '#F56C6C'
            },
            areaStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: 'rgba(245, 108, 108, 0.5)' },
                { offset: 1, color: 'rgba(245, 108, 108, 0.1)' }
              ])
            }
          }
        ]
      };
      
      userActivityChart.setOption(option);
      
      // 监听窗口变化，更新图表大小
      window.addEventListener('resize', () => {
        userActivityChart.resize();
      });
    };
    
    // 初始化协议使用情况图表
    const initProtocolChart = () => {
      if (protocolChart) {
        protocolChart.dispose();
      }
      
      const chartDom = protocolChartRef.value;
      protocolChart = echarts.init(chartDom);
      
      // 生成模拟数据
      const protocolData = [
        { value: 40, name: 'VMess' },
        { value: 30, name: 'VLESS' },
        { value: 15, name: 'Trojan' },
        { value: 10, name: 'Shadowsocks' },
        { value: 5, name: 'Socks' }
      ];
      
      const option = {
        tooltip: {
          trigger: 'item',
          formatter: '{a} <br/>{b}: {c} ({d}%)'
        },
        legend: {
          orient: 'vertical',
          right: 10,
          top: 'center',
          data: protocolData.map(item => item.name)
        },
        series: [
          {
            name: '协议使用',
            type: 'pie',
            radius: ['40%', '70%'],
            avoidLabelOverlap: false,
            itemStyle: {
              borderRadius: 10,
              borderColor: '#fff',
              borderWidth: 2
            },
            label: {
              show: false,
              position: 'center'
            },
            emphasis: {
              label: {
                show: true,
                fontSize: 16,
                fontWeight: 'bold'
              }
            },
            labelLine: {
              show: false
            },
            data: protocolData
          }
        ]
      };
      
      protocolChart.setOption(option);
      
      // 监听窗口变化，更新图表大小
      window.addEventListener('resize', () => {
        protocolChart.resize();
      });
    };
    
    // 初始化系统资源使用情况图表
    const initSystemResourceChart = () => {
      if (systemResourceChart) {
        systemResourceChart.dispose();
      }
      
      const chartDom = systemResourceChartRef.value;
      systemResourceChart = echarts.init(chartDom);
      
      // 生成模拟数据
      const cpuUsage = Math.floor(Math.random() * 40) + 20; // 20-60%
      const memoryUsage = Math.floor(Math.random() * 30) + 40; // 40-70%
      const diskUsage = Math.floor(Math.random() * 20) + 10; // 10-30%
      
      const option = {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'shadow'
          }
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true
        },
        xAxis: {
          type: 'value',
          max: 100,
          axisLabel: {
            formatter: '{value}%'
          }
        },
        yAxis: {
          type: 'category',
          data: ['CPU使用率', '内存使用率', '磁盘使用率']
        },
        series: [
          {
            name: 'CPU使用率',
            type: 'bar',
            data: [cpuUsage],
            itemStyle: {
              color: function(params) {
                const value = params.value;
                if (value < 30) {
                  return '#67C23A';
                } else if (value < 60) {
                  return '#E6A23C';
                } else {
                  return '#F56C6C';
                }
              }
            },
            label: {
              show: true,
              position: 'right',
              formatter: '{c}%'
            }
          },
          {
            name: '内存使用率',
            type: 'bar',
            data: [memoryUsage],
            itemStyle: {
              color: function(params) {
                const value = params.value;
                if (value < 50) {
                  return '#67C23A';
                } else if (value < 80) {
                  return '#E6A23C';
                } else {
                  return '#F56C6C';
                }
              }
            },
            label: {
              show: true,
              position: 'right',
              formatter: '{c}%'
            }
          },
          {
            name: '磁盘使用率',
            type: 'bar',
            data: [diskUsage],
            itemStyle: {
              color: function(params) {
                const value = params.value;
                if (value < 60) {
                  return '#67C23A';
                } else if (value < 80) {
                  return '#E6A23C';
                } else {
                  return '#F56C6C';
                }
              }
            },
            label: {
              show: true,
              position: 'right',
              formatter: '{c}%'
            }
          }
        ]
      };
      
      systemResourceChart.setOption(option);
      
      // 监听窗口变化，更新图表大小
      window.addEventListener('resize', () => {
        systemResourceChart.resize();
      });
    };
    
    // 组件加载时初始化数据
    onMounted(() => {
      loadStats();
    });
    
    // 组件卸载时清理
    onUnmounted(() => {
      if (refreshTimer) {
        clearInterval(refreshTimer);
        refreshTimer = null;
      }
      
      if (trafficChart) {
        trafficChart.dispose();
        trafficChart = null;
      }
      
      if (userActivityChart) {
        userActivityChart.dispose();
        userActivityChart = null;
      }
      
      if (protocolChart) {
        protocolChart.dispose();
        protocolChart = null;
      }
      
      if (systemResourceChart) {
        systemResourceChart.dispose();
        systemResourceChart = null;
      }
      
      window.removeEventListener('resize', () => {});
    });
    
    return {
      trafficChartRef,
      userActivityChartRef,
      protocolChartRef,
      systemResourceChartRef,
      timeRange,
      trafficChartType,
      realTimeMonitor,
      loading,
      stats,
      formatBytes,
      refreshData,
      handleTimeRangeChange,
      updateTrafficChart,
      toggleRealTimeMonitor
    };
  }
};
</script>

<style scoped>
.stats-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.stats-cards {
  margin-bottom: 20px;
}

.stats-card {
  margin-bottom: 20px;
  transition: all 0.3s;
}

.stats-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.1);
}

.stats-card-content {
  display: flex;
  align-items: center;
}

.stats-card-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.stats-card-icon.system {
  background-color: rgba(64, 158, 255, 0.1);
  color: #409EFF;
}

.stats-card-icon.traffic {
  background-color: rgba(103, 194, 58, 0.1);
  color: #67C23A;
}

.stats-card-icon.users {
  background-color: rgba(230, 162, 60, 0.1);
  color: #E6A23C;
}

.stats-card-icon.active {
  background-color: rgba(245, 108, 108, 0.1);
  color: #F56C6C;
}

.stats-card-icon .el-icon {
  font-size: 24px;
}

.stats-card-info {
  flex: 1;
}

.stats-card-title {
  font-size: 14px;
  color: #606266;
  margin-bottom: 5px;
}

.stats-card-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.chart-card {
  margin-bottom: 20px;
}

.chart-row {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  height: 400px;
}
</style> 