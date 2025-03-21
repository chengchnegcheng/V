import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: () => import('@/layouts/DefaultLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/Dashboard.vue')
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('@/views/Users.vue')
        },
        {
          path: 'proxies',
          name: 'proxies',
          component: () => import('@/views/Proxies.vue')
        },
        {
          path: 'traffic',
          name: 'traffic',
          component: () => import('@/views/TrafficMonitor.vue')
        },
        {
          path: 'monitor',
          name: 'monitor',
          component: () => import('@/views/SystemMonitor.vue')
        },
        {
          path: 'logs',
          name: 'logs',
          component: () => import('@/views/Logs.vue')
        },
        {
          path: 'certificates',
          name: 'certificates',
          component: () => import('@/views/Certificates.vue')
        },
        {
          path: 'backups',
          name: 'backups',
          component: () => import('@/views/Backups.vue')
        },
        {
          path: 'alerts',
          name: 'alerts',
          component: () => import('@/views/AlertSettings.vue')
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/Settings.vue')
        }
      ]
    }
  ]
})

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)

  if (requiresAuth && !userStore.isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && userStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router 