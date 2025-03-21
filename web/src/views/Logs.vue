<template>
  <div class="logs-container">
    <div class="header">
      <h1>日志管理</h1>
      <div class="actions">
        <el-select v-model="logType" placeholder="日志类型" style="width: 150px;">
          <el-option label="全部日志" value="all" />
          <el-option label="系统日志" value="system" />
          <el-option label="错误日志" value="error" />
          <el-option label="访问日志" value="access" />
          <el-option label="用户操作" value="user" />
        </el-select>
        <el-select v-model="logLevel" placeholder="日志级别" style="width: 150px;">
          <el-option label="全部级别" value="all" />
          <el-option label="INFO" value="info" />
          <el-option label="WARNING" value="warning" />
          <el-option label="ERROR" value="error" />
          <el-option label="CRITICAL" value="critical" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          :shortcuts="dateShortcuts"
        />
        <el-input
          v-model="searchQuery"
          placeholder="搜索关键词"
          style="width: 200px;"
          clearable
        />
        <el-button type="primary" @click="fetchLogs">查询</el-button>
        <el-button type="success" @click="exportLogs">导出</el-button>
      </div>
    </div>

    <!-- 日志表格 -->
    <el-card class="log-table-card">
      <el-table 
        :data="filteredLogs" 
        v-loading="loading" 
        style="width: 100%"
        height="calc(100vh - 250px)"
      >
        <el-table-column type="expand">
          <template #default="props">
            <div class="log-detail">
              <p><strong>日志详情：</strong></p>
              <pre>{{ props.row.message }}</pre>
              <p v-if="props.row.stackTrace"><strong>堆栈跟踪：</strong></p>
              <pre v-if="props.row.stackTrace">{{ props.row.stackTrace }}</pre>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="timestamp" label="时间" width="180" sortable>
          <template #default="scope">
            {{ formatDateTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="120">
          <template #default="scope">
            <el-tag 
              :type="getLogTypeTag(scope.row.type)"
              size="small"
            >
              {{ getLogTypeText(scope.row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="level" label="级别" width="100">
          <template #default="scope">
            <el-tag 
              :type="getLogLevelTag(scope.row.level)"
              size="small"
            >
              {{ scope.row.level.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="150" />
        <el-table-column prop="message" label="内容">
          <template #default="scope">
            <div class="message-preview">{{ scope.row.message }}</div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button 
              size="small" 
              type="primary" 
              @click="viewLogDetail(scope.row)"
            >
              详情
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteLog(scope.row)"
              v-if="hasDeletePermission"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="totalLogs"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 日志详情对话框 -->
    <el-dialog 
      title="日志详情" 
      v-model="logDetailVisible" 
      width="700px"
    >
      <template v-if="selectedLog">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="时间">{{ formatDateTime(selectedLog.timestamp) }}</el-descriptions-item>
          <el-descriptions-item label="来源">{{ selectedLog.source }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="getLogTypeTag(selectedLog.type)">
              {{ getLogTypeText(selectedLog.type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="级别">
            <el-tag :type="getLogLevelTag(selectedLog.level)">
              {{ selectedLog.level.toUpperCase() }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="IP地址" v-if="selectedLog.ip">{{ selectedLog.ip }}</el-descriptions-item>
          <el-descriptions-item label="用户" v-if="selectedLog.username">{{ selectedLog.username }}</el-descriptions-item>
        </el-descriptions>

        <div class="log-content">
          <p><strong>日志内容：</strong></p>
          <pre>{{ selectedLog.message }}</pre>
        </div>

        <div class="stack-trace" v-if="selectedLog.stackTrace">
          <p><strong>堆栈跟踪：</strong></p>
          <pre>{{ selectedLog.stackTrace }}</pre>
        </div>

        <div class="metadata" v-if="selectedLog.metadata">
          <p><strong>元数据：</strong></p>
          <pre>{{ JSON.stringify(selectedLog.metadata, null, 2) }}</pre>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'Logs',
  setup() {
    // 状态
    const loading = ref(false)
    const logs = ref([])
    const totalLogs = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(20)
    const logType = ref('all')
    const logLevel = ref('all')
    const searchQuery = ref('')
    const dateRange = ref([])
    const logDetailVisible = ref(false)
    const selectedLog = ref(null)
    const hasDeletePermission = ref(true) // 根据实际权限控制

    // 日期快捷选项
    const dateShortcuts = [
      {
        text: '今天',
        value: () => {
          const end = new Date()
          const start = new Date()
          start.setHours(0, 0, 0, 0)
          return [start, end]
        }
      },
      {
        text: '昨天',
        value: () => {
          const end = new Date()
          end.setHours(23, 59, 59, 999)
          end.setDate(end.getDate() - 1)
          const start = new Date()
          start.setHours(0, 0, 0, 0)
          start.setDate(start.getDate() - 1)
          return [start, end]
        }
      },
      {
        text: '最近一周',
        value: () => {
          const end = new Date()
          const start = new Date()
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 7)
          return [start, end]
        }
      },
      {
        text: '最近一个月',
        value: () => {
          const end = new Date()
          const start = new Date()
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 30)
          return [start, end]
        }
      }
    ]

    // 过滤后的日志数据
    const filteredLogs = computed(() => {
      return logs.value
    })

    // 获取日志数据
    const fetchLogs = async () => {
      loading.value = true
      
      try {
        // TODO: 替换为实际的 API 调用
        // const params = {
        //   page: currentPage.value,
        //   size: pageSize.value,
        //   type: logType.value !== 'all' ? logType.value : undefined,
        //   level: logLevel.value !== 'all' ? logLevel.value : undefined,
        //   query: searchQuery.value || undefined,
        //   startTime: dateRange.value && dateRange.value[0] ? dateRange.value[0].toISOString() : undefined,
        //   endTime: dateRange.value && dateRange.value[1] ? dateRange.value[1].toISOString() : undefined
        // }
        // const response = await logsService.getLogs(params)
        // logs.value = response.data.logs
        // totalLogs.value = response.data.total
        
        // 模拟数据
        setTimeout(() => {
          const sampleLogs = []
          const logTypes = ['system', 'error', 'access', 'user']
          const logLevels = ['info', 'warning', 'error', 'critical']
          const sources = ['系统', '用户模块', '认证模块', '流量模块', '代理服务', '协议处理', '控制台']
          
          const getRandomMessage = (type, level) => {
            if (type === 'system') {
              if (level === 'info') return '系统启动完成，监听端口: 8080'
              if (level === 'warning') return '系统资源使用率超过80%，请关注服务器负载'
              if (level === 'error') return '配置文件加载失败: 无法解析 config.json'
              if (level === 'critical') return '严重错误: 数据库连接失败，无法启动服务'
            } else if (type === 'error') {
              if (level === 'info') return '错误已被自动修复: 连接超时问题'
              if (level === 'warning') return '发现潜在问题: 内存泄漏可能出现在用户连接处理中'
              if (level === 'error') return 'HTTP请求处理错误: 404 找不到页面 /api/v1/users/123'
              if (level === 'critical') return '严重错误: 无法写入日志文件，磁盘空间已满'
            } else if (type === 'access') {
              if (level === 'info') return '用户登录成功: admin 从 192.168.1.100'
              if (level === 'warning') return '多次失败的登录尝试: 用户 user1 从 192.168.1.101'
              if (level === 'error') return '访问被拒绝: 用户 user2 尝试访问未授权资源'
              if (level === 'critical') return '检测到潜在的暴力破解攻击，来源IP: 192.168.1.102'
            } else if (type === 'user') {
              if (level === 'info') return '用户 admin 修改了系统设置'
              if (level === 'warning') return '用户 user1 尝试修改管理员密码'
              if (level === 'error') return '用户操作失败: user2 尝试删除系统日志'
              if (level === 'critical') return '检测到可疑的用户操作: user3 尝试修改系统关键文件'
            }
            return '日志消息'
          }
          
          for (let i = 0; i < 200; i++) {
            const type = logTypes[Math.floor(Math.random() * logTypes.length)]
            const level = logLevels[Math.floor(Math.random() * logLevels.length)]
            const source = sources[Math.floor(Math.random() * sources.length)]
            
            // 生成随机日期，最近30天内
            const timestamp = new Date()
            timestamp.setDate(timestamp.getDate() - Math.floor(Math.random() * 30))
            timestamp.setHours(Math.floor(Math.random() * 24))
            timestamp.setMinutes(Math.floor(Math.random() * 60))
            timestamp.setSeconds(Math.floor(Math.random() * 60))
            
            const message = getRandomMessage(type, level)
            
            const log = {
              id: i + 1,
              timestamp,
              type,
              level,
              source,
              message,
              stackTrace: level === 'error' || level === 'critical' ? 
                '错误在 /app/services/userService.js:123\n' +
                'at processRequest (/app/middleware/auth.js:45)\n' +
                'at handleRequest (/app/server.js:67)\n' +
                'at Server.emit (events.js:315)' : null,
              ip: type === 'access' ? `192.168.1.${Math.floor(Math.random() * 254) + 1}` : null,
              username: type === 'user' || type === 'access' ? 
                Math.random() > 0.7 ? 'admin' : `user${Math.floor(Math.random() * 10) + 1}` : null,
              metadata: {
                browser: 'Chrome',
                os: 'Windows',
                duration: Math.floor(Math.random() * 500) + 'ms'
              }
            }
            
            sampleLogs.push(log)
          }
          
          // 筛选
          let filtered = sampleLogs
          
          if (logType.value !== 'all') {
            filtered = filtered.filter(log => log.type === logType.value)
          }
          
          if (logLevel.value !== 'all') {
            filtered = filtered.filter(log => log.level === logLevel.value)
          }
          
          if (searchQuery.value) {
            const query = searchQuery.value.toLowerCase()
            filtered = filtered.filter(log => 
              log.message.toLowerCase().includes(query) || 
              (log.username && log.username.toLowerCase().includes(query)) ||
              log.source.toLowerCase().includes(query)
            )
          }
          
          if (dateRange.value && dateRange.value.length === 2) {
            const start = dateRange.value[0]
            const end = dateRange.value[1]
            filtered = filtered.filter(log => {
              const timestamp = new Date(log.timestamp)
              return timestamp >= start && timestamp <= end
            })
          }
          
          // 排序 - 按时间倒序
          filtered.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
          
          totalLogs.value = filtered.length
          
          // 分页
          const startIndex = (currentPage.value - 1) * pageSize.value
          const endIndex = startIndex + pageSize.value
          logs.value = filtered.slice(startIndex, endIndex)
          
          loading.value = false
        }, 500)
      } catch (error) {
        ElMessage.error('获取日志数据失败')
        console.error(error)
        loading.value = false
      }
    }

    // 导出日志数据
    const exportLogs = () => {
      ElMessage.success('日志导出成功')
    }

    // 查看日志详情
    const viewLogDetail = (log) => {
      selectedLog.value = log
      logDetailVisible.value = true
    }

    // 删除日志
    const deleteLog = (log) => {
      ElMessageBox.confirm(
        '确定要删除这条日志记录吗？此操作不可恢复。',
        '警告',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(async () => {
        try {
          // TODO: 替换为实际的 API 调用
          // await logsService.deleteLog(log.id)
          
          // 模拟删除
          logs.value = logs.value.filter(item => item.id !== log.id)
          totalLogs.value--
          
          ElMessage.success('日志删除成功')
        } catch (error) {
          ElMessage.error('日志删除失败')
          console.error(error)
        }
      }).catch(() => {})
    }

    // 分页处理
    const handleSizeChange = (size) => {
      pageSize.value = size
      fetchLogs()
    }

    const handleCurrentChange = (page) => {
      currentPage.value = page
      fetchLogs()
    }

    // 获取日志类型标签样式
    const getLogTypeTag = (type) => {
      const typeMap = {
        'system': 'primary',
        'error': 'danger',
        'access': 'success',
        'user': 'warning'
      }
      return typeMap[type] || 'info'
    }

    // 获取日志类型文本
    const getLogTypeText = (type) => {
      const typeMap = {
        'system': '系统日志',
        'error': '错误日志',
        'access': '访问日志',
        'user': '用户操作'
      }
      return typeMap[type] || type
    }

    // 获取日志级别标签样式
    const getLogLevelTag = (level) => {
      const levelMap = {
        'info': 'info',
        'warning': 'warning',
        'error': 'danger',
        'critical': 'danger'
      }
      return levelMap[level] || 'info'
    }

    // 格式化日期时间
    const formatDateTime = (date) => {
      if (!date) return '未知'
      const d = new Date(date)
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
    }

    // 初始化
    onMounted(() => {
      fetchLogs()
    })

    return {
      logs,
      filteredLogs,
      loading,
      totalLogs,
      currentPage,
      pageSize,
      logType,
      logLevel,
      searchQuery,
      dateRange,
      dateShortcuts,
      logDetailVisible,
      selectedLog,
      hasDeletePermission,
      fetchLogs,
      exportLogs,
      viewLogDetail,
      deleteLog,
      handleSizeChange,
      handleCurrentChange,
      getLogTypeTag,
      getLogTypeText,
      getLogLevelTag,
      formatDateTime
    }
  }
}
</script>

<style scoped>
.logs-container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.actions {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.log-table-card {
  margin-bottom: 20px;
}

.message-preview {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 400px;
}

.log-detail {
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.log-detail pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  font-family: monospace;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.log-content,
.stack-trace,
.metadata {
  margin-top: 20px;
}

.log-content pre,
.stack-trace pre,
.metadata pre {
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  font-family: monospace;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style> 