import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './assets/styles/main.css'

// 创建路由
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('./views/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: () => import('./layouts/MainLayout.vue'),
      children: [
        { 
          path: '', 
          name: 'dashboard',
          component: () => import('./views/Dashboard.vue') 
        },
        { 
          path: 'users', 
          name: 'users',
          component: () => import('./views/Users.vue') 
        },
        { 
          path: 'proxies', 
          name: 'proxies',
          component: () => import('./views/Proxies.vue') 
        },
        { 
          path: 'settings', 
          name: 'settings',
          component: () => import('./views/Settings.vue') 
        }
      ],
      meta: { requiresAuth: true }
    }
  ]
})

// 创建Pinia状态管理
const pinia = createPinia()

// 创建应用实例
const app = createApp(App)

// 注册所有Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 使用插件
app.use(router)
app.use(pinia)
app.use(ElementPlus)

// 挂载应用
app.mount('#app') 