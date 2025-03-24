import axios from 'axios'

// 配置axios拦截器，确保正确处理API响应
axios.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API error:', error)
    return Promise.reject(error)
  }
)

const api = {
  /**
   * 获取系统监控数据
   * @returns {Promise} 返回系统监控数据
   */
  getSystemStatus() {
    return axios.get('/api/system/status')
  },

  /**
   * 获取系统信息
   * @returns {Promise} 返回系统信息
   */
  getSystemInfo() {
    return axios.get('/api/system/info')
  },

  /**
   * 获取CPU使用情况
   * @returns {Promise} 返回CPU使用情况
   */
  getCpuUsage() {
    return axios.get('/api/system/cpu')
  },

  /**
   * 获取内存使用情况
   * @returns {Promise} 返回内存使用情况
   */
  getMemoryUsage() {
    return axios.get('/api/system/memory')
  },

  /**
   * 获取磁盘使用情况
   * @returns {Promise} 返回磁盘使用情况
   */
  getDiskUsage() {
    return axios.get('/api/system/disk')
  },

  /**
   * 获取进程列表
   * @returns {Promise} 返回进程列表
   */
  getProcessList() {
    return axios.get('/api/system/processes')
  }
}

export default api 