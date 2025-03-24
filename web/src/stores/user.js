import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '')
  const role = computed(() => user.value?.role || '')
  const userId = computed(() => user.value?.id || null)

  // 方法
  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (userInfo) => {
    user.value = userInfo
  }

  const clearAuth = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  // API方法
  const login = async (credentials) => {
    loading.value = true
    error.value = null
    
    try {
      console.log('Login attempt:', credentials.username, '********')
      
      // 使用完整URL直接访问后端API
      const response = await axios.post('http://localhost:8080/api/auth/login', credentials, {
        timeout: 15000, // 15秒超时
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        }
      })
      
      console.log('Login response:', response.data)
      
      // 检查响应中是否包含错误信息
      if (response.data && response.data.error) {
        console.error('Login error from server:', response.data.error)
        error.value = response.data.error
        throw new Error(response.data.error)
      }
      
      // 检查响应是否包含必要的数据
      if (!response.data.token || !response.data.user) {
        console.error('Invalid login response format:', response.data)
        error.value = '服务器返回的数据格式不正确'
        throw new Error('服务器返回的数据格式不正确')
      }
      
      const { token: newToken, user: userInfo } = response.data
      console.log('Login successful:', userInfo)
      
      setToken(newToken)
      setUser(userInfo)
      return true
    } catch (err) {
      console.error('Login error:', err)
      
      // 详细记录错误信息
      if (err.response) {
        console.error('Error response:', err.response.status, err.response.data)
        error.value = err.response.data.error || err.message || '登录失败'
      } else if (err.request) {
        console.error('No response received:', err.request)
        error.value = '网络错误，服务器未响应'
      } else {
        console.error('Request setup error:', err.message)
        error.value = err.message || '请求配置错误'
      }
      
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const logout = () => {
    // 实际项目中可能需要调用登出API
    clearAuth()
  }

  const getUser = async () => {
    if (!token.value) return null
    
    loading.value = true
    error.value = null
    
    try {
      // 使用完整URL直接访问后端API
      const response = await axios.get('http://localhost:8080/api/auth/user', {
        headers: {
          'Authorization': `Bearer ${token.value}`
        }
      })
      console.log('Get user response:', response.data)
      
      if (response.data && response.data.error) {
        error.value = response.data.error
        throw new Error(response.data.error)
      }
      
      if (!response.data.user) {
        error.value = '服务器返回的数据格式不正确'
        throw new Error('服务器返回的数据格式不正确')  
      }
      
      setUser(response.data.user)
      return user.value
    } catch (err) {
      console.error('Get user error:', err)
      
      if (err.response) {
        console.error('Error response:', err.response.status, err.response.data)
        error.value = err.response.data.error || '获取用户信息失败'
      } else if (err.request) {
        error.value = '网络错误，服务器未响应'
      } else {
        error.value = err.message || '获取用户信息失败'
      }
      
      throw error.value
    } finally {
      loading.value = false
    }
  }

  // 用户管理方法
  const fetchUsers = async (params) => {
    loading.value = true
    error.value = null
    
    try {
      console.log('Fetching users with params:', params)
      const response = await axios.get('http://localhost:8080/api/users', { 
        params,
        headers: {
          'Authorization': `Bearer ${token.value}`
        }
      })
      console.log('Users response:', response.data)
      
      // 适配不同的响应格式
      let users = [];
      let total = 0;
      
      if (Array.isArray(response.data)) {
        // 如果响应直接是数组
        users = response.data;
        total = response.data.length;
      } else if (response.data && response.data.users) {
        // 如果响应中有users字段
        users = response.data.users;
        total = response.data.total || users.length;
      } else if (response.data && response.data.data && Array.isArray(response.data.data)) {
        // 如果响应中有data字段
        users = response.data.data;
        total = response.data.total || users.length;
      } else {
        console.error('Unknown response format:', response.data);
        users = [];
        total = 0;
      }
      
      return {
        users,
        total
      }
    } catch (err) {
      console.error('Fetch users error:', err)
      
      if (err.response) {
        console.error('Error response:', err.response.status, err.response.data)
        error.value = err.response.data.error || '获取用户列表失败'
      } else if (err.request) {
        error.value = '网络错误，服务器未响应'
      } else {
        error.value = err.message || '获取用户列表失败'
      }
      
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const createUser = async (userData) => {
    loading.value = true
    error.value = null
    
    try {
      // 实际项目中应替换为真实API端点
      // const response = await axios.post('/api/users', userData)
      // 模拟成功响应
      const newUser = {
        ...userData,
        id: Date.now(),
        created: new Date().toLocaleString(),
        lastLogin: '-',
        status: true
      }
      
      return newUser
    } catch (err) {
      error.value = err.response?.data?.message || '创建用户失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const updateUser = async (userId, userData) => {
    loading.value = true
    error.value = null
    
    try {
      // 实际项目中应替换为真实API端点
      // const response = await axios.put(`/api/users/${userId}`, userData)
      // 模拟成功响应
      return { ...userData, id: userId }
    } catch (err) {
      error.value = err.response?.data?.message || '更新用户失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const deleteUser = async (userId) => {
    loading.value = true
    error.value = null
    
    try {
      // 实际项目中应替换为真实API端点
      // await axios.delete(`/api/users/${userId}`)
      // 模拟成功响应
      return true
    } catch (err) {
      error.value = err.response?.data?.message || '删除用户失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  const updateUserStatus = async (userId, status) => {
    loading.value = true
    error.value = null
    
    try {
      // 实际项目中应替换为真实API端点
      // await axios.patch(`/api/users/${userId}/status`, { status })
      // 模拟成功响应
      return true
    } catch (err) {
      error.value = err.response?.data?.message || '更新用户状态失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }
  
  const updateUserProfile = async (profileData) => {
    loading.value = true
    error.value = null
    
    try {
      // 实际项目中应替换为真实API端点
      // const response = await axios.put('/api/users/profile', profileData)
      
      // 模拟成功响应
      // 更新本地用户数据
      user.value = {
        ...user.value,
        ...profileData
      }
      
      return user.value
    } catch (err) {
      error.value = err.response?.data?.message || '更新个人资料失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }
  
  const changePassword = async (passwordData) => {
    loading.value = true
    error.value = null
    
    try {
      // 实际项目中应替换为真实API端点
      // await axios.post('/api/users/change-password', passwordData)
      
      // 模拟成功响应
      return true
    } catch (err) {
      error.value = err.response?.data?.message || '修改密码失败'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  return {
    // 状态
    token,
    user,
    loading,
    error,
    
    // 计算属性
    isLoggedIn,
    username,
    role,
    userId,
    
    // 方法
    login,
    logout,
    getUser,
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
    updateUserStatus,
    updateUserProfile,
    changePassword
  }
}) 