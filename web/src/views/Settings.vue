<template>
  <div class="settings-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>系统设置</span>
          <el-button type="primary" @click="saveAllSettings">保存所有设置</el-button>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <!-- 基本设置 -->
        <el-tab-pane label="基本设置" name="basic">
          <el-form
            ref="basicFormRef"
            :model="basicSettings"
            label-width="120px"
          >
            <el-form-item label="系统语言">
              <el-select v-model="basicSettings.language">
                <el-option label="简体中文" value="zh-CN" />
                <el-option label="English" value="en-US" />
              </el-select>
            </el-form-item>
            <el-form-item label="时区">
              <el-select v-model="basicSettings.timezone">
                <el-option label="(GMT+08:00) 北京" value="Asia/Shanghai" />
                <el-option label="(GMT+00:00) 伦敦" value="Europe/London" />
                <el-option label="(GMT-05:00) 纽约" value="America/New_York" />
              </el-select>
            </el-form-item>
            <el-form-item label="主题">
              <el-select v-model="basicSettings.theme">
                <el-option label="浅色" value="light" />
                <el-option label="深色" value="dark" />
                <el-option label="跟随系统" value="system" />
              </el-select>
            </el-form-item>
            <el-form-item label="面板端口">
              <el-input-number
                v-model="basicSettings.port"
                :min="1"
                :max="65535"
                :step="1"
              />
            </el-form-item>
            <el-form-item label="面板路径">
              <el-input v-model="basicSettings.path" />
            </el-form-item>
            <el-form-item label="面板标题">
              <el-input v-model="basicSettings.title" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 安全设置 -->
        <el-tab-pane label="安全设置" name="security">
          <el-form
            ref="securityFormRef"
            :model="securitySettings"
            label-width="120px"
          >
            <el-form-item label="登录安全">
              <el-checkbox-group v-model="securitySettings.loginSecurity">
                <el-checkbox label="enableCaptcha">启用验证码</el-checkbox>
                <el-checkbox label="enable2FA">启用双因素认证</el-checkbox>
                <el-checkbox label="enableIPWhitelist">启用IP白名单</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="密码策略">
              <el-checkbox-group v-model="securitySettings.passwordPolicy">
                <el-checkbox label="requireUppercase">必须包含大写字母</el-checkbox>
                <el-checkbox label="requireLowercase">必须包含小写字母</el-checkbox>
                <el-checkbox label="requireNumbers">必须包含数字</el-checkbox>
                <el-checkbox label="requireSpecialChars">必须包含特殊字符</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="最小密码长度">
              <el-input-number
                v-model="securitySettings.minPasswordLength"
                :min="6"
                :max="32"
                :step="1"
              />
            </el-form-item>
            <el-form-item label="密码有效期">
              <el-input-number
                v-model="securitySettings.passwordExpiry"
                :min="0"
                :max="365"
                :step="1"
              />
              <span class="form-item-tip">天（0表示永不过期）</span>
            </el-form-item>
            <el-form-item label="登录失败限制">
              <el-input-number
                v-model="securitySettings.maxLoginAttempts"
                :min="0"
                :max="10"
                :step="1"
              />
              <span class="form-item-tip">次（0表示不限制）</span>
            </el-form-item>
            <el-form-item label="锁定时间">
              <el-input-number
                v-model="securitySettings.lockoutDuration"
                :min="0"
                :max="1440"
                :step="1"
              />
              <span class="form-item-tip">分钟（0表示不锁定）</span>
            </el-form-item>
            <el-form-item label="IP白名单">
              <el-input
                v-model="securitySettings.ipWhitelist"
                type="textarea"
                :rows="4"
                placeholder="每行一个IP地址"
              />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 通知设置 -->
        <el-tab-pane label="通知设置" name="notification">
          <el-form
            ref="notificationFormRef"
            :model="notificationSettings"
            label-width="120px"
          >
            <el-form-item label="邮件通知">
              <el-switch v-model="notificationSettings.email.enabled" />
            </el-form-item>
            <template v-if="notificationSettings.email.enabled">
              <el-form-item label="SMTP服务器">
                <el-input v-model="notificationSettings.email.smtpServer" />
              </el-form-item>
              <el-form-item label="SMTP端口">
                <el-input-number
                  v-model="notificationSettings.email.smtpPort"
                  :min="1"
                  :max="65535"
                  :step="1"
                />
              </el-form-item>
              <el-form-item label="SMTP用户名">
                <el-input v-model="notificationSettings.email.smtpUsername" />
              </el-form-item>
              <el-form-item label="SMTP密码">
                <el-input
                  v-model="notificationSettings.email.smtpPassword"
                  type="password"
                  show-password
                />
              </el-form-item>
              <el-form-item label="发件人邮箱">
                <el-input v-model="notificationSettings.email.senderEmail" />
              </el-form-item>
              <el-form-item label="发件人名称">
                <el-input v-model="notificationSettings.email.senderName" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="testEmail">测试邮件</el-button>
              </el-form-item>
            </template>

            <el-form-item label="系统通知">
              <el-checkbox-group v-model="notificationSettings.system">
                <el-checkbox label="login">登录通知</el-checkbox>
                <el-checkbox label="password">密码修改通知</el-checkbox>
                <el-checkbox label="backup">备份完成通知</el-checkbox>
                <el-checkbox label="restore">恢复完成通知</el-checkbox>
              </el-checkbox-group>
            </el-form-item>

            <el-form-item label="告警通知">
              <el-checkbox-group v-model="notificationSettings.alerts">
                <el-checkbox label="cpu">CPU使用率过高</el-checkbox>
                <el-checkbox label="memory">内存使用率过高</el-checkbox>
                <el-checkbox label="disk">磁盘使用率过高</el-checkbox>
                <el-checkbox label="network">网络异常</el-checkbox>
                <el-checkbox label="service">服务异常</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="告警阈值">
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-form-item label="CPU">
                    <el-input-number
                      v-model="notificationSettings.thresholds.cpu"
                      :min="0"
                      :max="100"
                      :step="1"
                    />
                    <span class="form-item-tip">%</span>
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="内存">
                    <el-input-number
                      v-model="notificationSettings.thresholds.memory"
                      :min="0"
                      :max="100"
                      :step="1"
                    />
                    <span class="form-item-tip">%</span>
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="磁盘">
                    <el-input-number
                      v-model="notificationSettings.thresholds.disk"
                      :min="0"
                      :max="100"
                      :step="1"
                    />
                    <span class="form-item-tip">%</span>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { settingsApi } from '@/api'

// 当前激活的标签页
const activeTab = ref('basic')

// 基本设置
const basicSettings = ref({
  language: 'zh-CN',
  timezone: 'Asia/Shanghai',
  theme: 'light',
  port: 8080,
  path: '/',
  title: 'V Panel'
})

// 安全设置
const securitySettings = ref({
  loginSecurity: ['enableCaptcha'],
  passwordPolicy: ['requireUppercase', 'requireLowercase', 'requireNumbers'],
  minPasswordLength: 8,
  passwordExpiry: 90,
  maxLoginAttempts: 5,
  lockoutDuration: 30,
  ipWhitelist: ''
})

// 通知设置
const notificationSettings = ref({
  email: {
    enabled: false,
    smtpServer: '',
    smtpPort: 587,
    smtpUsername: '',
    smtpPassword: '',
    senderEmail: '',
    senderName: ''
  },
  system: ['login', 'password'],
  alerts: ['cpu', 'memory', 'disk'],
  thresholds: {
    cpu: 80,
    memory: 80,
    disk: 80
  }
})

// 获取所有设置
const getSettings = async () => {
  try {
    const [basic, security, notification] = await Promise.all([
      settingsApi.getBasicSettings(),
      settingsApi.getSecuritySettings(),
      settingsApi.getNotificationSettings()
    ])
    basicSettings.value = basic
    securitySettings.value = security
    notificationSettings.value = notification
  } catch (error) {
    ElMessage.error('获取设置失败')
  }
}

// 保存所有设置
const saveAllSettings = async () => {
  try {
    await Promise.all([
      settingsApi.updateBasicSettings(basicSettings.value),
      settingsApi.updateSecuritySettings(securitySettings.value),
      settingsApi.updateNotificationSettings(notificationSettings.value)
    ])
    ElMessage.success('设置保存成功')
  } catch (error) {
    ElMessage.error('设置保存失败')
  }
}

// 测试邮件
const testEmail = async () => {
  try {
    await settingsApi.testEmail(notificationSettings.value.email)
    ElMessage.success('测试邮件发送成功')
  } catch (error) {
    ElMessage.error('测试邮件发送失败')
  }
}

onMounted(() => {
  getSettings()
})
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-item-tip {
  margin-left: 10px;
  color: #909399;
  font-size: 14px;
}
</style> 