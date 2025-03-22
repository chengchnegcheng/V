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
          path: 'roles', 
          name: 'roles',
          component: () => import('./views/Roles.vue') 
        },
        { 
          path: 'proxies', 
          name: 'proxies',
          component: () => import('./views/Proxies.vue') 
        },
        { 
          path: 'traffic', 
          name: 'traffic',
          component: () => import('./views/TrafficMonitor.vue') 
        },
        { 
          path: 'monitor', 
          name: 'monitor',
          component: () => import('./views/SystemMonitor.vue') 
        },
        { 
          path: 'logs', 
          name: 'logs',
          component: () => import('./views/Logs.vue') 
        },
        { 
          path: 'certificates', 
          name: 'certificates',
          component: () => import('./views/Certificates.vue') 
        },
        { 
          path: 'backups', 
          name: 'backups',
          component: () => import('./views/Backups.vue') 
        },
        { 
          path: 'alerts', 
          name: 'alerts',
          component: () => import('./views/AlertSettings.vue') 
        },
        { 
          path: 'settings', 
          name: 'settings',
          component: () => import('./views/Settings.vue') 
        },
        { 
          path: 'stats', 
          name: 'stats',
          component: () => import('./views/Stats.vue') 
        }
      ],
      meta: { requiresAuth: true }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'notFound',
      component: () => import('./views/NotFound.vue')
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const isAuthenticated = localStorage.getItem('token')
  
  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && isAuthenticated) {
    next('/')
  } else {
    next()
  }
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