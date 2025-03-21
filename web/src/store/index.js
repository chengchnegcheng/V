import { createStore } from 'vuex'
import { userApi } from '@/api'

export default createStore({
  state: {
    user: null,
    token: localStorage.getItem('token') || '',
    settings: null,
    permissions: []
  },
  mutations: {
    SET_TOKEN(state, token) {
      state.token = token
      localStorage.setItem('token', token)
    },
    SET_USER(state, user) {
      state.user = user
    },
    SET_SETTINGS(state, settings) {
      state.settings = settings
    },
    SET_PERMISSIONS(state, permissions) {
      state.permissions = permissions
    },
    CLEAR_USER(state) {
      state.user = null
      state.token = ''
      state.permissions = []
      localStorage.removeItem('token')
    }
  },
  actions: {
    // 登录
    async login({ commit }, loginData) {
      const { token, user, permissions } = await userApi.login(loginData)
      commit('SET_TOKEN', token)
      commit('SET_USER', user)
      commit('SET_PERMISSIONS', permissions)
      return user
    },
    // 登出
    async logout({ commit }) {
      await userApi.logout()
      commit('CLEAR_USER')
    },
    // 获取用户信息
    async getUserInfo({ commit }) {
      const { user, permissions } = await userApi.getInfo()
      commit('SET_USER', user)
      commit('SET_PERMISSIONS', permissions)
      return user
    },
    // 更新用户信息
    async updateUserInfo({ commit }, userData) {
      const user = await userApi.updateInfo(userData)
      commit('SET_USER', user)
      return user
    },
    // 修改密码
    async changePassword({ commit }, passwordData) {
      await userApi.changePassword(passwordData)
    }
  },
  getters: {
    isAuthenticated: state => !!state.token,
    currentUser: state => state.user,
    userSettings: state => state.settings,
    hasPermission: state => permission => {
      if (!state.permissions) return false
      return state.permissions.includes(permission)
    },
    hasAnyPermission: state => permissions => {
      if (!state.permissions) return false
      return permissions.some(permission => state.permissions.includes(permission))
    }
  }
}) 