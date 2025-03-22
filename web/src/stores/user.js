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
      // 实际项目中应替换为真实API端点
      // const response = await axios.post('/api/auth/login', credentials)
      // 模拟成功响应
      const response = { data: { token: 'mock-token', user: { id: 1, username: credentials.username, role: 'admin' } } }
      
      const { token: newToken, user: userInfo } = response.data
      setToken(newToken)
      setUser(userInfo)
      return true
    } catch (err) {
      error.value = err.response?.data?.message || '登录失败'
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
      // 实际项目中应替换为真实API端点
      // const response = await axios.get('/api/auth/user')
      // 模拟成功响应
      const response = { data: { user: { id: 1, username: 'admin', role: 'admin' } } }
      
      setUser(response.data.user)
      return user.value
    } catch (err) {
      error.value = err.response?.data?.message || '获取用户信息失败'
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
      // 实际项目中应替换为真实API端点
      // const response = await axios.get('/api/users', { params })
      // 模拟成功响应
      const mockUsers = [
        {
          id: 1,
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin',
          created: '2023-01-01 12:00:00',
          lastLogin: '2023-03-15 08:30:22',
          status: true
        },
        {
          id: 2,
          username: 'user1',
          email: 'user1@example.com',
          role: 'user',
          created: '2023-01-02 14:30:00',
          lastLogin: '2023-03-14 16:42:51',
          status: true
        },
        {
          id: 3,
          username: 'moderator',
          email: 'mod@example.com',
          role: 'user',
          created: '2023-01-03 10:15:30',
          lastLogin: '2023-03-10 11:20:15',
          status: true
        }
      ]
      
      const response = { 
        data: { 
          users: mockUsers,
          total: mockUsers.length
        } 
      }
      
      return {
        users: response.data.users,
        total: response.data.total
      }
    } catch (err) {
      error.value = err.response?.data?.message || '获取用户列表失败'
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
    updateUserStatus
  }
}) 