import Mock from 'mockjs'

// 生成模拟系统状态数据
const generateMockSystemStatus = () => {
  // CPU信息
  const cpuCores = 8
  const cpuUsage = Math.floor(Math.random() * 60) + 10
  
  // 根据浏览器环境推断操作系统类型
  let osType = 'Unknown'
  let cpuModel = 'Unknown CPU Model'
  
  if (navigator.platform.indexOf('Win') !== -1) {
    osType = 'Windows'
    cpuModel = 'Intel Core i7-10700K @ 3.80GHz (Windows)'
  } else if (navigator.platform.indexOf('Mac') !== -1) {
    osType = 'macOS'
    cpuModel = 'Apple M1 (macOS)'
  } else if (navigator.platform.indexOf('Linux') !== -1) {
    osType = 'Linux'
    cpuModel = 'Intel Xeon E5-2680 (Linux)'
  }
  
  const cpuInfo = {
    cores: cpuCores,
    model: cpuModel
  }
  
  // 内存信息
  const totalMem = 16 * 1024 * 1024 * 1024 // 16GB
  const usedMem = totalMem * (Math.random() * 0.5 + 0.2)
  const memoryUsage = Math.floor((usedMem / totalMem) * 100)
  const memoryInfo = {
    used: usedMem,
    total: totalMem
  }
  
  // 磁盘信息
  const totalDisk = 500 * 1024 * 1024 * 1024 // 500GB
  const usedDisk = totalDisk * (Math.random() * 0.6 + 0.2)
  const diskUsage = Math.floor((usedDisk / totalDisk) * 100)
  const diskInfo = {
    used: usedDisk,
    total: totalDisk
  }
  
  // 系统信息
  const systemInfo = {
    os: osType,
    kernel: navigator.userAgent.indexOf('Windows') > -1 ? 'Windows NT 10.0' : 
            navigator.userAgent.indexOf('Mac') > -1 ? 'Darwin 21.6.0' : 
            'Linux 5.15.0-76-generic',
    hostname: window.location.hostname || 'localhost',
    uptime: '0 days, 0 hours, 0 minutes', // 模拟的运行时间
    load: osType === 'Windows' ? [0, 0, 0] : [0.8, 1.0, 1.2], // Windows 不显示负载
    ipAddress: '0.0.0.0'
  }
  
  // 进程信息 - 根据操作系统生成不同的进程列表
  let processes = []
  
  if (osType === 'Windows') {
    processes = [
      { pid: 4, name: 'System', user: 'SYSTEM', cpu: '0.1', memory: '0.5', memoryUsed: 50 * 1024 * 1024, started: '2023-03-10 00:00:00', state: 'running' },
      { pid: 728, name: 'svchost.exe', user: 'SYSTEM', cpu: '1.2', memory: '0.8', memoryUsed: 80 * 1024 * 1024, started: '2023-03-15 08:30:00', state: 'running' },
      { pid: 1524, name: 'v.exe', user: 'USER', cpu: '2.5', memory: '1.2', memoryUsed: 120 * 1024 * 1024, started: '2023-03-15 08:31:20', state: 'running' }
    ]
  } else if (osType === 'macOS') {
    processes = [
      { pid: 1, name: 'launchd', user: 'root', cpu: '0.1', memory: '0.3', memoryUsed: 30 * 1024 * 1024, started: '2023-03-10 00:00:00', state: 'running' },
      { pid: 324, name: 'WindowServer', user: 'root', cpu: '1.5', memory: '1.0', memoryUsed: 100 * 1024 * 1024, started: '2023-03-15 08:30:00', state: 'running' },
      { pid: 1524, name: 'v', user: 'user', cpu: '2.0', memory: '1.1', memoryUsed: 110 * 1024 * 1024, started: '2023-03-15 08:31:20', state: 'running' }
    ]
  } else {
    processes = [
      { pid: 1, name: 'systemd', user: 'root', cpu: '0.5', memory: '0.8', memoryUsed: 80 * 1024 * 1024, started: '2023-03-10 00:00:00', state: 'running' },
      { pid: 854, name: 'v-core', user: 'root', cpu: '2.1', memory: '1.2', memoryUsed: 120 * 1024 * 1024, started: '2023-03-15 08:30:00', state: 'running' },
      { pid: 1275, name: 'nginx', user: 'www-data', cpu: '1.5', memory: '0.7', memoryUsed: 70 * 1024 * 1024, started: '2023-03-15 08:31:20', state: 'running' }
    ]
  }
  
  // 添加一些随机进程来填充列表
  for (let i = 0; i < 5; i++) {
    const randomProc = {
      pid: 2000 + i,
      name: ['chrome', 'firefox', 'code', 'node', 'python'][Math.floor(Math.random() * 5)],
      user: osType === 'Windows' ? 'USER' : 'user',
      cpu: (Math.random() * 5).toFixed(1),
      memory: (Math.random() * 3).toFixed(1),
      memoryUsed: Math.floor(Math.random() * 500) * 1024 * 1024,
      started: '2023-03-15 09:' + Math.floor(Math.random() * 60).toString().padStart(2, '0') + ':00',
      state: 'running'
    }
    processes.push(randomProc)
  }
  
  // 按CPU使用率排序
  processes.sort((a, b) => parseFloat(b.cpu) - parseFloat(a.cpu))
  
  return {
    cpuInfo,
    cpuUsage,
    memoryInfo,
    memoryUsage,
    diskInfo,
    diskUsage,
    systemInfo,
    processes
  }
}

// 设置响应延迟
Mock.setup({
  timeout: '200-600'
})

// =============================================
// 注释掉所有的Mock拦截器，防止它们拦截API请求
// =============================================

// Mock.mock('/api/system/status', 'get', () => {
//   return {
//     code: 200,
//     message: 'success',
//     data: generateMockSystemStatus()
//   }
// })

// Mock.mock('/api/system/info', 'get', () => {
//   // 根据浏览器环境推断操作系统类型
//   let osType = 'Unknown'
//   
//   if (navigator.platform.indexOf('Win') !== -1) {
//     osType = 'Windows'
//   } else if (navigator.platform.indexOf('Mac') !== -1) {
//     osType = 'macOS'
//   } else if (navigator.platform.indexOf('Linux') !== -1) {
//     osType = 'Linux'
//   }
//   
//   return {
//     code: 200,
//     message: 'success',
//     data: {
//       os: osType,
//       kernel: navigator.userAgent.indexOf('Windows') > -1 ? 'Windows NT 10.0' : 
//               navigator.userAgent.indexOf('Mac') > -1 ? 'Darwin 21.6.0' : 
//               'Linux 5.15.0-76-generic',
//       hostname: window.location.hostname || 'localhost',
//       uptime: '0 days, 0 hours, 0 minutes',
//       load: osType === 'Windows' ? [0, 0, 0] : [0.8, 1.0, 1.2],
//       ipAddress: '0.0.0.0'
//     }
//   }
// })

// 标记这个模块已被禁用
console.log('Mock module has been disabled to use real backend APIs')

export default Mock 