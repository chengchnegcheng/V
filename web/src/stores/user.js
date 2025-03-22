import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || null,
    user: null
  }),
  
  getters: {
    isAuthenticated: (state) => !!state.token,
    username: (state) => state.user?.username || 'Admin'
  },
  
  actions: {
    login(credentials) {
      // 这里应该调用登录API，现在只是模拟
      if (credentials.username === 'admin' && credentials.password === 'admin') {
        const token = 'mock-token-' + Date.now()
        localStorage.setItem('token', token)
        this.token = token
        this.user = {
          id: 1,
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin'
        }
        return Promise.resolve(true)
      }
      return Promise.reject(new Error('用户名或密码错误'))
    },
    
    logout() {
      localStorage.removeItem('token')
      this.token = null
      this.user = null
    },
    
    fetchUserInfo() {
      // 这里应该调用获取用户信息API，现在只是模拟
      if (this.token) {
        this.user = {
          id: 1,
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin'
        }
        return Promise.resolve(this.user)
      }
      return Promise.reject(new Error('未登录'))
    }
  }
}) 