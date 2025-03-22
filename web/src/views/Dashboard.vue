<template>
  <div class="dashboard-container">
    <h1>系统控制面板</h1>
    
    <!-- 系统状态卡片 -->
    <div class="status-cards">
      <el-card class="status-card">
        <h3>CPU 使用率</h3>
        <div class="status-value">{{ cpuUsage }}%</div>
      </el-card>
      
      <el-card class="status-card">
        <h3>内存使用率</h3>
        <div class="status-value">{{ memoryUsage }}%</div>
      </el-card>
      
      <el-card class="status-card">
        <h3>磁盘使用率</h3>
        <div class="status-value">{{ diskUsage }}%</div>
      </el-card>
      
      <el-card class="status-card">
        <h3>活跃连接数</h3>
        <div class="status-value">{{ activeConnections }}</div>
      </el-card>
    </div>
    
    <!-- 流量统计 -->
    <el-card class="traffic-card">
      <div class="card-header">
        <h2>流量统计</h2>
      </div>
      <div class="traffic-info">
        <div class="traffic-item">
          <h3>今日上行流量</h3>
          <div class="traffic-value">{{ formatTraffic(todayUpload) }}</div>
        </div>
        <div class="traffic-item">
          <h3>今日下行流量</h3>
          <div class="traffic-value">{{ formatTraffic(todayDownload) }}</div>
        </div>
        <div class="traffic-item">
          <h3>本月上行流量</h3>
          <div class="traffic-value">{{ formatTraffic(monthUpload) }}</div>
        </div>
        <div class="traffic-item">
          <h3>本月下行流量</h3>
          <div class="traffic-value">{{ formatTraffic(monthDownload) }}</div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Dashboard',
  data() {
    return {
      cpuUsage: 15,
      memoryUsage: 32,
      diskUsage: 58,
      activeConnections: 11,
      todayUpload: 1024 * 1024 * 150, // 150MB
      todayDownload: 1024 * 1024 * 350, // 350MB
      monthUpload: 1024 * 1024 * 1024 * 5, // 5GB
      monthDownload: 1024 * 1024 * 1024 * 15 // 15GB
    }
  },
  methods: {
    formatTraffic(bytes) {
      if (bytes < 1024) {
        return bytes + ' B'
      } else if (bytes < 1024 * 1024) {
        return (bytes / 1024).toFixed(2) + ' KB'
      } else if (bytes < 1024 * 1024 * 1024) {
        return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
      } else {
        return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
      }
    }
  }
}
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}

.status-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

.status-card {
  text-align: center;
}

.status-value {
  font-size: 24px;
  font-weight: bold;
  color: #409EFF;
  margin-top: 10px;
}

.traffic-card {
  margin-bottom: 20px;
}

.card-header {
  margin-bottom: 20px;
}

.traffic-info {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.traffic-item {
  text-align: center;
}

.traffic-value {
  font-size: 18px;
  font-weight: bold;
  color: #67C23A;
  margin-top: 10px;
}

@media (max-width: 1200px) {
  .status-cards {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .traffic-info {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .status-cards {
    grid-template-columns: 1fr;
  }
  
  .traffic-info {
    grid-template-columns: 1fr;
  }
}
</style> 