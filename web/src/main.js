import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import App from './App.vue'
import './assets/styles/main.css'
import './assets/styles/base.scss'
import router from './router'

// 导入事件源客户端
import xrayEventSource from './utils/eventSourceClient'

// 创建Vue实例和状态管理
const app = createApp(App)
const pinia = createPinia()

// 注册所有Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 使用插件
app.use(router)
app.use(pinia)
app.use(ElementPlus)

// 完全禁用所有Mock数据
console.log('Mock functionality is completely disabled. Using real backend API.')

// 初始化SSE连接监听Xray版本事件
xrayEventSource.init()

// 挂载应用
app.mount('#app') 