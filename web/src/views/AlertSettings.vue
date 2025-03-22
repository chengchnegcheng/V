<template>
  <div class="alert-settings">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>告警配置</span>
        </div>
      </template>

      <el-form :model="form" label-width="120px">
        <!-- CPU告警配置 -->
        <el-form-item label="CPU告警">
          <el-switch v-model="form.cpu.enabled" />
          <template v-if="form.cpu.enabled">
            <el-input-number 
              v-model="form.cpu.threshold" 
              :min="0" 
              :max="100" 
              :step="5"
              placeholder="告警阈值"
            />
            <span class="unit">%</span>
            <el-input-number 
              v-model="form.cpu.duration" 
              :min="1" 
              :max="60" 
              :step="1"
              placeholder="持续时间"
            />
            <span class="unit">分钟</span>
          </template>
        </el-form-item>

        <!-- 内存告警配置 -->
        <el-form-item label="内存告警">
          <el-switch v-model="form.memory.enabled" />
          <template v-if="form.memory.enabled">
            <el-input-number 
              v-model="form.memory.threshold" 
              :min="0" 
              :max="100" 
              :step="5"
              placeholder="告警阈值"
            />
            <span class="unit">%</span>
            <el-input-number 
              v-model="form.memory.duration" 
              :min="1" 
              :max="60" 
              :step="1"
              placeholder="持续时间"
            />
            <span class="unit">分钟</span>
          </template>
        </el-form-item>

        <!-- 磁盘告警配置 -->
        <el-form-item label="磁盘告警">
          <el-switch v-model="form.disk.enabled" />
          <template v-if="form.disk.enabled">
            <el-input-number 
              v-model="form.disk.threshold" 
              :min="0" 
              :max="100" 
              :step="5"
              placeholder="告警阈值"
            />
            <span class="unit">%</span>
            <el-input-number 
              v-model="form.disk.duration" 
              :min="1" 
              :max="60" 
              :step="1"
              placeholder="持续时间"
            />
            <span class="unit">分钟</span>
          </template>
        </el-form-item>

        <!-- 网络流量告警配置 -->
        <el-form-item label="网络流量告警">
          <el-switch v-model="form.network.enabled" />
          <template v-if="form.network.enabled">
            <el-input-number 
              v-model="form.network.threshold" 
              :min="0" 
              :step="100"
              placeholder="告警阈值"
            />
            <span class="unit">MB/s</span>
            <el-input-number 
              v-model="form.network.duration" 
              :min="1" 
              :max="60" 
              :step="1"
              placeholder="持续时间"
            />
            <span class="unit">分钟</span>
          </template>
        </el-form-item>

        <!-- 告警通知配置 -->
        <el-divider>告警通知</el-divider>
        
        <!-- 邮件通知 -->
        <el-form-item label="邮件通知">
          <el-switch v-model="form.notification.email.enabled" />
          <template v-if="form.notification.email.enabled">
            <el-input 
              v-model="form.notification.email.smtpServer" 
              placeholder="SMTP服务器"
            />
            <el-input 
              v-model="form.notification.email.smtpPort" 
              placeholder="SMTP端口"
            />
            <el-input 
              v-model="form.notification.email.username" 
              placeholder="用户名"
            />
            <el-input 
              v-model="form.notification.email.password" 
              type="password" 
              placeholder="密码"
            />
            <el-input 
              v-model="form.notification.email.from" 
              placeholder="发件人邮箱"
            />
            <el-input 
              v-model="form.notification.email.to" 
              placeholder="收件人邮箱"
            />
          </template>
        </el-form-item>

        <!-- Webhook通知 -->
        <el-form-item label="Webhook通知">
          <el-switch v-model="form.notification.webhook.enabled" />
          <template v-if="form.notification.webhook.enabled">
            <el-input 
              v-model="form.notification.webhook.url" 
              placeholder="Webhook URL"
            />
            <el-input 
              v-model="form.notification.webhook.secret" 
              type="password" 
              placeholder="Webhook密钥"
            />
          </template>
        </el-form-item>

        <!-- 告警历史 -->
        <el-divider>告警历史</el-divider>
        <el-table :data="alertHistory" style="width: 100%">
          <el-table-column prop="time" label="时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.time) }}
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="120" />
          <el-table-column prop="value" label="值" width="120">
            <template #default="scope">
              {{ scope.row.value }}{{ scope.row.unit }}
            </template>
          </el-table-column>
          <el-table-column prop="threshold" label="阈值" width="120">
            <template #default="scope">
              {{ scope.row.threshold }}{{ scope.row.unit }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="120">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'resolved' ? 'success' : 'danger'">
                {{ scope.row.status === 'resolved' ? '已恢复' : '告警中' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <el-form-item>
          <el-button type="primary" @click="saveSettings">保存配置</el-button>
          <el-button @click="resetSettings">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { monitorApi } from '@/api'

const form = ref({
  cpu: {
    enabled: false,
    threshold: 80,
    duration: 5
  },
  memory: {
    enabled: false,
    threshold: 80,
    duration: 5
  },
  disk: {
    enabled: false,
    threshold: 80,
    duration: 5
  },
  network: {
    enabled: false,
    threshold: 100,
    duration: 5
  },
  notification: {
    email: {
      enabled: false,
      smtpServer: '',
      smtpPort: '',
      username: '',
      password: '',
      from: '',
      to: ''
    },
    webhook: {
      enabled: false,
      url: '',
      secret: ''
    }
  }
})

const alertHistory = ref([])

// 格式化时间
const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 获取告警配置
const getAlertSettings = async () => {
  try {
    // 模拟获取告警配置
    // const res = await monitorApi.getAlertSettings()
    // form.value = res.data
    console.log('加载告警配置')
  } catch (error) {
    ElMessage.error('获取告警配置失败')
  }
}

// 获取告警历史
const getAlertHistory = async () => {
  try {
    // 模拟获取告警历史
    // const res = await monitorApi.getAlertHistory()
    // alertHistory.value = res.data
    
    // 使用模拟数据
    alertHistory.value = [
      {
        id: 1,
        time: new Date(Date.now() - 3600000),
        type: 'CPU',
        value: 92,
        threshold: 80,
        status: 'resolved',
        unit: '%'
      },
      {
        id: 2,
        time: new Date(Date.now() - 7200000),
        type: '内存',
        value: 85,
        threshold: 80,
        status: 'resolved',
        unit: '%'
      },
      {
        id: 3,
        time: new Date(),
        type: '磁盘',
        value: 95,
        threshold: 80,
        status: 'alerting',
        unit: '%'
      }
    ]
  } catch (error) {
    ElMessage.error('获取告警历史失败')
  }
}

// 保存配置
const saveSettings = async () => {
  try {
    // await monitorApi.updateAlertSettings(form.value)
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 重置配置
const resetSettings = () => {
  getAlertSettings()
}

// 初始化数据
onMounted(() => {
  getAlertSettings()
  getAlertHistory()
})
</script>

<style scoped>
.alert-settings {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.unit {
  margin-left: 8px;
  color: #666;
}

.el-input-number {
  width: 120px;
}

.el-input {
  width: 300px;
  margin-bottom: 10px;
}
</style> 