import { createRouter, createWebHistory } from 'vue-router'

// 使用懒加载方式导入组件
const MainLayout = () => import('../layouts/MainLayout.vue')
const Login = () => import('../views/Login.vue')
const Dashboard = () => import('../views/Dashboard.vue')
const Settings = () => import('../views/Settings.vue')
const Inbounds = () => import('../views/Inbounds.vue')
const NotFound = () => import('../views/NotFound.vue')
const Profile = () => import('../views/Profile.vue')
const ChangePassword = () => import('../views/ChangePassword.vue')

const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: Dashboard,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: Settings,
        meta: { requiresAuth: true }
      },
      {
        path: 'inbounds',
        name: 'Inbounds',
        component: Inbounds,
        meta: { requiresAuth: true }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('../views/Users.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('../views/RolesNew.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'system-monitor',
        name: 'SystemMonitor',
        component: () => import('../views/SystemMonitor.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'traffic-monitor',
        name: 'TrafficMonitor',
        component: () => import('../views/TrafficMonitor.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'stats',
        name: 'Stats',
        component: () => import('../views/StatsNew.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('../views/Logs.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'certificates',
        name: 'Certificates',
        component: () => import('../views/Certificates.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'backups',
        name: 'Backups',
        component: () => import('../views/Backups.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: Profile,
        meta: { requiresAuth: true }
      },
      {
        path: 'change-password',
        name: 'ChangePassword',
        component: ChangePassword,
        meta: { requiresAuth: true }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue')
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound
  }
]

// 添加路由守卫
const router = createRouter({
  history: createWebHistory(),
  routes
})

// 全局前置守卫
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

export default router 