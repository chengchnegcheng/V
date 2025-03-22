<template>
  <div class="settings-container">
    <h1>系统设置</h1>
    
    <el-tabs type="border-card">
      <el-tab-pane label="服务器配置">
        <el-form :model="serverForm" label-width="120px" class="settings-form">
          <el-form-item label="面板监听地址">
            <el-input v-model="serverForm.panelListenIP" placeholder="0.0.0.0"></el-input>
            <div class="form-tips">默认为 0.0.0.0，代表监听所有 IP</div>
          </el-form-item>
          <el-form-item label="面板端口">
            <el-input-number v-model="serverForm.panelPort" :min="1" :max="65535"></el-input-number>
            <div class="form-tips">默认为 9000，修改后需要重启服务</div>
          </el-form-item>
          <el-form-item label="面板URL基础路径">
            <el-input v-model="serverForm.panelBasePath" placeholder="/"></el-input>
            <div class="form-tips">默认为 /，修改后需要重启服务</div>
          </el-form-item>
          <el-form-item label="代理服务模式">
            <el-select v-model="serverForm.proxyMode" style="width: 100%">
              <el-option label="兼容模式" value="compatible"></el-option>
              <el-option label="Xray 内核" value="xray"></el-option>
              <el-option label="V2Ray 内核" value="v2ray"></el-option>
            </el-select>
            <div class="form-tips">默认为兼容模式，可同时使用 Xray 和 V2Ray 协议</div>
          </el-form-item>
          <el-form-item label="服务时区">
            <el-select v-model="serverForm.timezone" style="width: 100%">
              <el-option label="Asia/Shanghai (UTC+8)" value="Asia/Shanghai"></el-option>
              <el-option label="UTC" value="UTC"></el-option>
              <el-option label="America/New_York (UTC-5)" value="America/New_York"></el-option>
              <el-option label="Europe/London (UTC+0)" value="Europe/London"></el-option>
            </el-select>
          </el-form-item>
          <el-divider></el-divider>
          <el-form-item>
            <el-button type="primary" @click="saveServerSettings">保存服务器配置</el-button>
            <el-button type="warning" @click="restartPanel">重启面板</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <el-tab-pane label="数据库配置">
        <el-form :model="dbForm" label-width="120px" class="settings-form">
          <el-form-item label="数据库类型">
            <el-select v-model="dbForm.dbType" style="width: 100%">
              <el-option label="SQLite" value="sqlite"></el-option>
              <el-option label="MySQL" value="mysql"></el-option>
              <el-option label="PostgreSQL" value="postgres"></el-option>
            </el-select>
          </el-form-item>
          
          <template v-if="dbForm.dbType !== 'sqlite'">
            <el-form-item label="数据库服务器">
              <el-input v-model="dbForm.dbHost" placeholder="localhost"></el-input>
            </el-form-item>
            <el-form-item label="数据库端口">
              <el-input-number 
                v-model="dbForm.dbPort" 
                :min="1" 
                :max="65535"
                :placeholder="dbForm.dbType === 'mysql' ? '3306' : '5432'"
              ></el-input-number>
            </el-form-item>
            <el-form-item label="数据库名称">
              <el-input v-model="dbForm.dbName" placeholder="v_panel"></el-input>
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="dbForm.dbUser" placeholder="root"></el-input>
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="dbForm.dbPassword" type="password" placeholder="密码" show-password></el-input>
            </el-form-item>
          </template>
          
          <template v-else>
            <el-form-item label="SQLite文件路径">
              <el-input v-model="dbForm.sqlitePath" placeholder="/usr/local/v-panel/data.db"></el-input>
              <div class="form-tips">默认在程序目录下的 data.db 文件</div>
            </el-form-item>
          </template>
          
          <el-divider></el-divider>
          <el-form-item>
            <el-button type="primary" @click="saveDbSettings">保存数据库配置</el-button>
            <el-button @click="testDbConnection">测试连接</el-button>
            <el-button type="success" @click="backupDb">备份数据库</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <el-tab-pane label="日志配置">
        <el-form :model="logForm" label-width="120px" class="settings-form">
          <el-form-item label="日志级别">
            <el-select v-model="logForm.logLevel" style="width: 100%">
              <el-option label="DEBUG" value="debug"></el-option>
              <el-option label="INFO" value="info"></el-option>
              <el-option label="WARN" value="warn"></el-option>
              <el-option label="ERROR" value="error"></el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="日志保留天数">
            <el-input-number v-model="logForm.logRetentionDays" :min="1" :max="365"></el-input-number>
            <div class="form-tips">超过该天数的日志将被自动清理</div>
          </el-form-item>
          <el-form-item label="日志存储路径">
            <el-input v-model="logForm.logPath"></el-input>
            <div class="form-tips">默认在程序目录下的 logs 文件夹</div>
          </el-form-item>
          <el-form-item label="启用访问日志">
            <el-switch v-model="logForm.enableAccessLog"></el-switch>
            <div class="form-tips">记录所有HTTP请求访问日志</div>
          </el-form-item>
          <el-form-item label="启用操作日志">
            <el-switch v-model="logForm.enableOperationLog"></el-switch>
            <div class="form-tips">记录所有用户操作日志</div>
          </el-form-item>
          <el-divider></el-divider>
          <el-form-item>
            <el-button type="primary" @click="saveLogSettings">保存日志配置</el-button>
            <el-button type="danger" @click="clearLogs">清理日志</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <el-tab-pane label="管理员配置">
        <el-form :model="adminForm" label-width="120px" class="settings-form">
          <el-alert
            title="管理员账号安全提示"
            type="warning"
            description="修改管理员密码后，当前会话将被注销，需要重新登录。请确保记住新密码，否则可能无法访问系统。"
            show-icon
            :closable="false"
            style="margin-bottom: 20px"
          />
          
          <el-form-item label="管理员用户名">
            <el-input v-model="adminForm.username" placeholder="admin" :disabled="true"></el-input>
            <div class="form-tips">默认管理员用户名不可修改</div>
          </el-form-item>
          <el-form-item label="当前密码">
            <el-input v-model="adminForm.currentPassword" type="password" placeholder="当前密码" show-password></el-input>
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="adminForm.newPassword" type="password" placeholder="新密码" show-password></el-input>
          </el-form-item>
          <el-form-item label="确认新密码">
            <el-input v-model="adminForm.confirmPassword" type="password" placeholder="确认新密码" show-password></el-input>
          </el-form-item>
          
          <el-divider></el-divider>
          <el-form-item>
            <el-button type="primary" @click="changeAdminPassword">修改密码</el-button>
            <el-button type="warning" @click="resetAdminPassword">重置为默认密码</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <el-tab-pane label="安全设置">
        <el-form :model="securityForm" label-width="120px" class="settings-form">
          <el-form-item label="会话超时时间">
            <el-input-number v-model="securityForm.sessionTimeout" :min="5" :max="1440"></el-input-number>
            <div class="form-tips">单位：分钟，超过该时间未操作将自动注销</div>
          </el-form-item>
          <el-form-item label="启用IP白名单">
            <el-switch v-model="securityForm.enableIpWhitelist"></el-switch>
          </el-form-item>
          <el-form-item label="IP白名单" v-if="securityForm.enableIpWhitelist">
            <el-input 
              v-model="securityForm.ipWhitelist" 
              type="textarea" 
              :rows="4"
              placeholder="每行一个IP地址，支持CIDR格式，如：192.168.1.0/24"
            ></el-input>
          </el-form-item>
          <el-form-item label="登录失败锁定">
            <el-switch v-model="securityForm.enableLoginLock"></el-switch>
            <div class="form-tips">连续登录失败将暂时锁定账号</div>
          </el-form-item>
          <el-form-item label="失败尝试次数" v-if="securityForm.enableLoginLock">
            <el-input-number v-model="securityForm.maxLoginAttempts" :min="3" :max="10"></el-input-number>
          </el-form-item>
          <el-form-item label="锁定时间(分钟)" v-if="securityForm.enableLoginLock">
            <el-input-number v-model="securityForm.lockDuration" :min="5" :max="60"></el-input-number>
          </el-form-item>
          
          <el-divider></el-divider>
          <el-form-item>
            <el-button type="primary" @click="saveSecuritySettings">保存安全设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'

// store
const userStore = useUserStore()

// 表单数据
const serverForm = reactive({
  panelListenIP: '0.0.0.0',
  panelPort: 9000,
  panelBasePath: '/',
  proxyMode: 'compatible',
  timezone: 'Asia/Shanghai'
})

const dbForm = reactive({
  dbType: 'sqlite',
  dbHost: 'localhost',
  dbPort: 3306,
  dbName: 'v_panel',
  dbUser: 'root',
  dbPassword: '',
  sqlitePath: '/usr/local/v-panel/data.db'
})

const logForm = reactive({
  logLevel: 'info',
  logRetentionDays: 30,
  logPath: '/usr/local/v-panel/logs',
  enableAccessLog: true,
  enableOperationLog: true
})

const adminForm = reactive({
  username: 'admin',
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const securityForm = reactive({
  sessionTimeout: 30,
  enableIpWhitelist: false,
  ipWhitelist: '',
  enableLoginLock: true,
  maxLoginAttempts: 5,
  lockDuration: 10
})

// 方法
const saveServerSettings = async () => {
  try {
    // 在实际项目中应调用API保存配置
    // await api.saveServerSettings(serverForm)
    ElMessage.success('服务器配置保存成功')
  } catch (error) {
    ElMessage.error('保存失败：' + error.message)
  }
}

const restartPanel = () => {
  ElMessageBox.confirm(
    '确定要重启面板吗？这将暂时中断所有连接。',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
  .then(async () => {
    try {
      // 在实际项目中应调用API重启面板
      // await api.restartPanel()
      ElMessage.success('面板重启指令已发送，请稍后刷新页面')
    } catch (error) {
      ElMessage.error('重启失败：' + error.message)
    }
  })
  .catch(() => {
    ElMessage.info('已取消重启')
  })
}

const saveDbSettings = async () => {
  try {
    // 在实际项目中应调用API保存配置
    // await api.saveDbSettings(dbForm)
    ElMessage.success('数据库配置保存成功')
  } catch (error) {
    ElMessage.error('保存失败：' + error.message)
  }
}

const testDbConnection = async () => {
  try {
    // 在实际项目中应调用API测试连接
    // await api.testDbConnection(dbForm)
    ElMessage.success('数据库连接测试成功')
  } catch (error) {
    ElMessage.error('连接测试失败：' + error.message)
  }
}

const backupDb = async () => {
  try {
    // 在实际项目中应调用API备份数据库
    // await api.backupDatabase()
    ElMessage.success('数据库备份成功')
  } catch (error) {
    ElMessage.error('备份失败：' + error.message)
  }
}

const saveLogSettings = async () => {
  try {
    // 在实际项目中应调用API保存配置
    // await api.saveLogSettings(logForm)
    ElMessage.success('日志配置保存成功')
  } catch (error) {
    ElMessage.error('保存失败：' + error.message)
  }
}

const clearLogs = () => {
  ElMessageBox.confirm(
    '确定要清理所有日志吗？此操作不可恢复。',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
  .then(async () => {
    try {
      // 在实际项目中应调用API清理日志
      // await api.clearLogs()
      ElMessage.success('日志清理成功')
    } catch (error) {
      ElMessage.error('清理失败：' + error.message)
    }
  })
  .catch(() => {
    ElMessage.info('已取消清理')
  })
}

const changeAdminPassword = async () => {
  // 表单验证
  if (!adminForm.currentPassword) {
    return ElMessage.warning('请输入当前密码')
  }
  if (!adminForm.newPassword) {
    return ElMessage.warning('请输入新密码')
  }
  if (adminForm.newPassword.length < 6) {
    return ElMessage.warning('新密码长度不能少于6个字符')
  }
  if (adminForm.newPassword !== adminForm.confirmPassword) {
    return ElMessage.warning('两次输入的密码不一致')
  }
  
  ElMessageBox.confirm(
    '修改密码后，当前会话将被注销，需要重新登录。是否继续？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
  .then(async () => {
    try {
      // 在实际项目中应调用API修改密码
      // await api.changeAdminPassword(adminForm)
      ElMessage.success('密码修改成功，请重新登录')
      
      // 清空表单
      adminForm.currentPassword = ''
      adminForm.newPassword = ''
      adminForm.confirmPassword = ''
      
      // 注销当前会话
      setTimeout(() => {
        userStore.logout()
        window.location.href = '/login'
      }, 1500)
    } catch (error) {
      ElMessage.error('修改失败：' + error.message)
    }
  })
  .catch(() => {
    ElMessage.info('已取消修改')
  })
}

const resetAdminPassword = () => {
  ElMessageBox.confirm(
    '确定要将管理员密码重置为默认密码吗？',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
  .then(async () => {
    try {
      // 在实际项目中应调用API重置密码
      // await api.resetAdminPassword()
      ElMessage.success('密码重置成功，默认密码为：admin')
    } catch (error) {
      ElMessage.error('重置失败：' + error.message)
    }
  })
  .catch(() => {
    ElMessage.info('已取消重置')
  })
}

const saveSecuritySettings = async () => {
  try {
    // 在实际项目中应调用API保存配置
    // await api.saveSecuritySettings(securityForm)
    ElMessage.success('安全设置保存成功')
  } catch (error) {
    ElMessage.error('保存失败：' + error.message)
  }
}
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.settings-form {
  max-width: 800px;
  margin-top: 20px;
}

.form-tips {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

.el-divider {
  margin: 20px 0;
}
</style> 