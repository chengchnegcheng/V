import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const api = axios.create({
  baseURL: import.meta.env.VITE_APP_API_URL || '/api',
  timeout: 10000
})

// Request interceptor
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  response => {
    // 直接返回响应数据，不做额外处理
    return response.data
  },
  error => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          // Unauthorized
          localStorage.removeItem('token')
          router.push('/login')
          ElMessage.error(error.response.data.error || '登录已过期，请重新登录')
          break
        case 403:
          // Forbidden
          ElMessage.error(error.response.data.error || '没有权限访问')
          break
        case 404:
          // Not Found
          ElMessage.error(error.response.data.error || '请求的资源不存在')
          break
        case 500:
          // Internal Server Error
          ElMessage.error(error.response.data.error || '服务器内部错误')
          break
        default:
          ElMessage.error(error.response.data.error || '请求失败')
      }
    } else if (error.request) {
      // Request was made but no response was received
      ElMessage.error('网络错误，请检查网络连接')
    } else {
      // Something happened in setting up the request
      ElMessage.error('请求配置错误')
    }
    return Promise.reject(error)
  }
)

// Auth API
export const auth = {
  login: (data) => api.post('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  refresh: () => api.post('/auth/refresh'),
  profile: () => api.get('/auth/profile'),
  updateProfile: (data) => api.put('/auth/profile', data),
  changePassword: (data) => api.put('/auth/password', data)
}

// User API
export const users = {
  list: (params) => api.get('/users', { params }),
  get: (id) => api.get(`/users/${id}`),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
  enable: (id) => api.post(`/users/${id}/enable`),
  disable: (id) => api.post(`/users/${id}/disable`),
  resetPassword: (id) => api.post(`/users/${id}/reset-password`)
}

// Proxy API
export const proxies = {
  list: (params) => api.get('/proxies', { params }),
  get: (id) => api.get(`/proxies/${id}`),
  create: (data) => api.post('/proxies', data),
  update: (id, data) => api.put(`/proxies/${id}`, data),
  delete: (id) => api.delete(`/proxies/${id}`),
  start: (id) => api.post(`/proxies/${id}/start`),
  stop: (id) => api.post(`/proxies/${id}/stop`),
  stats: (id) => api.get(`/proxies/${id}/stats`)
}

// SSL证书管理API
export const certificatesApi = {
  // 获取证书列表
  getCertificates: () => {
    return api.get('/certificates')
  },

  // 申请证书
  applyCertificate: (data) => {
    return api.post('/certificates/apply', data)
  },

  // 上传证书
  uploadCertificate: (data) => {
    const formData = new FormData()
    formData.append('domain', data.domain)
    formData.append('cert', data.certFile)
    formData.append('key', data.keyFile)
    return api.post('/certificates/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  // 续期证书
  renewCertificate: (id) => {
    return api.post(`/certificates/${id}/renew`)
  },

  // 验证证书
  validateCertificate: (id) => {
    return api.get(`/certificates/${id}/validate`)
  },

  // 备份证书
  backupCertificate: (id) => {
    return api.get(`/certificates/${id}/backup`, {
      responseType: 'blob'
    }).then(response => {
      const url = window.URL.createObjectURL(new Blob([response]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `certificate_${id}_${new Date().getTime()}.zip`)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    })
  },

  // 删除证书
  deleteCertificate: (id) => {
    return api.delete(`/certificates/${id}`)
  },

  // 更新自动续期设置
  updateAutoRenew: (id, autoRenew) => {
    return api.put(`/certificates/${id}/auto-renew`, { autoRenew })
  }
}

// System API
export const systemApi = {
  getStats: () => api.get('/system/stats'),
  getProcesses: () => api.get('/system/processes'),
  killProcess: (pid) => api.post(`/system/processes/${pid}/kill`),
  getCpuTrend: (timeRange) => api.get('/system/trends/cpu', { params: { timeRange } }),
  getMemoryTrend: (timeRange) => api.get('/system/trends/memory', { params: { timeRange } }),
  getNetworkTrend: (timeRange) => api.get('/system/trends/network', { params: { timeRange } }),
  getLoadTrend: (timeRange) => api.get('/system/trends/load', { params: { timeRange } }),
  getDiskTrend: (timeRange) => api.get('/system/trends/disk', { params: { timeRange } }),
  getNetworkConnections: () => api.get('/system/network/connections'),
  getNetworkInterfaces: () => api.get('/system/network/interfaces'),
  getInterfaceTraffic: (interfaceName) => api.get(`/system/network/interfaces/${interfaceName}/traffic`),
  getDiskPartitions: () => api.get('/system/disk/partitions'),
  getIOStats: (timeRange) => api.get('/system/io/stats', { params: { timeRange } }),
  getSystemInfo: () => api.get('/system/info'),
  clearConnections: () => api.post('/system/network/connections/clear'),
  killConnection: (id) => api.post(`/system/network/connections/${id}/kill`)
}

// Event API
export const events = {
  list: (params) => api.get('/events', { params }),
  get: (id) => api.get(`/events/${id}`),
  clear: () => api.delete('/events')
}

// User API
export const userApi = {
  getInfo: () => api.get('/user/info'),
  updateInfo: (data) => api.put('/user/info', data),
  changePassword: (data) => api.put('/user/password', data)
}

// System Settings API
export const settingsApi = {
  getSettings: () => api.get('/settings'),
  updateSettings: (data) => api.put('/settings', data),
  createBackup: () => api.post('/settings/backup'),
  restoreBackup: (data) => api.post('/settings/restore', data)
}

// Stats API
export const statsApi = {
  getDashboardStats: () => api.get('/stats/dashboard'),
  getUserStats: (params) => api.get('/stats/user', { params }),
  getProtocolStats: (params) => api.get('/stats/protocol', { params }),
  getDetailedStats: (params) => api.get('/stats/detailed', { params })
}

// Monitor API
export const monitorApi = {
  getSystemStats: () => api.get('/monitor/system'),
  getNetworkStats: () => api.get('/monitor/network'),
  getProcessStats: () => api.get('/monitor/process'),
  getAlertSettings: () => api.get('/monitor/alerts/settings'),
  updateAlertSettings: (data) => api.put('/monitor/alerts/settings', data),
  getAlertHistory: () => api.get('/monitor/alerts/history'),
  testEmailNotification: (data) => api.post('/monitor/alerts/test-email', data),
  testWebhookNotification: (data) => api.post('/monitor/alerts/test-webhook', data),
  clearConnections: () => api.post('/monitor/connections/clear'),
  killConnection: (id) => api.post(`/monitor/connections/${id}/kill`),
  getTrafficHistory: (params) => api.get('/monitor/traffic/history', { params }),
  getProtocolDistribution: () => api.get('/monitor/traffic/protocols'),
  getConnectionList: () => api.get('/monitor/connections')
}

// Logs API
export const logsApi = {
  // 获取日志列表
  list: (params) => {
    return api.get('/logs', { params })
  },

  // 获取日志分析数据
  getAnalysis: () => {
    return api.get('/logs/analysis')
  },

  // 获取日志轮转配置
  getRotationConfig: () => {
    return api.get('/logs/rotation')
  },

  // 更新日志轮转配置
  updateRotationConfig: (data) => {
    return api.put('/logs/rotation', data)
  },

  // 导出日志
  export: (params) => {
    return api.get('/logs/export', {
      params,
      responseType: 'blob'
    })
  },

  // 清理日志
  clear: () => {
    return api.delete('/logs')
  },

  // 获取日志统计信息
  getStats: () => {
    return api.get('/logs/stats')
  },

  // 搜索日志
  search: (params) => {
    return api.get('/logs/search', { params })
  },

  // 批量删除日志
  batchDelete: (ids) => {
    return api.delete('/logs/batch', { data: { ids } })
  },

  // 获取日志级别统计
  getLevelStats: () => {
    return api.get('/logs/levels')
  },

  // 获取日志来源统计
  getSourceStats: () => {
    return api.get('/logs/sources')
  },

  // 获取日志时间分布
  getTimeDistribution: (params) => {
    return api.get('/logs/time-distribution', { params })
  }
}

// Backup API
export const backupsApi = {
  list: () => api.get('/backups'),
  get: (id) => api.get(`/backups/${id}`),
  create: (data) => api.post('/backups', data),
  delete: (id) => api.delete(`/backups/${id}`),
  download: (id) => api.get(`/backups/${id}/download`, {
    responseType: 'blob'
  }),
  verify: (id) => api.post(`/backups/${id}/verify`),
  restore: (data) => api.post('/backups/restore', data),
  upload: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/backups/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },
  getConfig: () => api.get('/backups/config'),
  updateConfig: (data) => api.put('/backups/config', data),
  testStorage: (data) => api.post('/backups/test-storage', data),
  getScheduleConfig: () => api.get('/backups/schedule'),
  updateScheduleConfig: (data) => api.put('/backups/schedule', data),
  getStorageConfig: () => api.get('/backups/storage'),
  updateStorageConfig: (data) => api.put('/backups/storage', data),
  getProgress: (id) => api.get(`/backups/${id}/progress`),
  cancel: (id) => api.post(`/backups/${id}/cancel`),
  getStats: () => api.get('/backups/stats'),
  cleanup: () => api.post('/backups/cleanup')
}

// 用户管理相关接口
export const usersApi = {
  // 获取用户列表
  list: () => {
    return api.get('/users')
  },

  // 创建用户
  create: (data) => {
    return api.post('/users', data)
  },

  // 更新用户
  update: (id, data) => {
    return api.put(`/users/${id}`, data)
  },

  // 删除用户
  delete: (id) => {
    return api.delete(`/users/${id}`)
  },

  // 重置密码
  resetPassword: (id) => {
    return api.post(`/users/${id}/reset-password`)
  },

  // 切换用户状态
  toggleStatus: (id) => {
    return api.post(`/users/${id}/toggle-status`)
  },

  // 获取角色列表
  getRoles: () => {
    return api.get('/roles')
  },

  // 创建角色
  createRole: (data) => {
    return api.post('/roles', data)
  },

  // 更新角色
  updateRole: (id, data) => {
    return api.put(`/roles/${id}`, data)
  },

  // 删除角色
  deleteRole: (id) => {
    return api.delete(`/roles/${id}`)
  },

  // 获取用户权限
  getPermissions: () => {
    return api.get('/permissions')
  },

  // 更新用户权限
  updatePermissions: (id, data) => {
    return api.put(`/users/${id}/permissions`, data)
  },

  // 获取用户登录历史
  getLoginHistory: (id) => {
    return api.get(`/users/${id}/login-history`)
  },

  // 清除用户登录历史
  clearLoginHistory: (id) => {
    return api.delete(`/users/${id}/login-history`)
  }
}

// Roles API
export const roles = {
  getRoles: () => api.get('/roles'),
  getRole: (id) => api.get(`/roles/${id}`),
  createRole: (data) => api.post('/roles', data),
  updateRole: (id, data) => api.put(`/roles/${id}`, data),
  deleteRole: (id) => api.delete(`/roles/${id}`)
}

// 客户端API
export const clientsApi = {
  // 获取客户端列表
  list: (params) => api.get('/clients', { params }),
  
  // 获取单个客户端详情
  get: (id) => api.get(`/clients/${id}`),
  
  // 创建新客户端
  create: (data) => api.post('/clients', data),
  
  // 更新客户端信息
  update: (id, data) => api.put(`/clients/${id}`, data),
  
  // 删除客户端
  delete: (id) => api.delete(`/clients/${id}`),
  
  // 启用客户端
  enable: (id) => api.post(`/clients/${id}/enable`),
  
  // 禁用客户端
  disable: (id) => api.post(`/clients/${id}/disable`),
  
  // 重置客户端流量
  resetTraffic: (id) => api.post(`/clients/${id}/reset-traffic`),
  
  // 生成客户端配置文件
  generateConfig: (id, type) => api.get(`/clients/${id}/config/${type}`, {
    responseType: 'blob'
  }).then(response => {
    const url = window.URL.createObjectURL(new Blob([response]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `client_config_${id}_${type}.json`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  }),
  
  // 获取客户端流量统计
  getTrafficStats: (id, params) => api.get(`/clients/${id}/traffic`, { params }),
  
  // 获取客户端连接记录
  getConnections: (id, params) => api.get(`/clients/${id}/connections`, { params })
}

export default api 