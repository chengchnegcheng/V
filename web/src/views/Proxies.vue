<template>
  <div class="proxies">
    <div class="header">
      <h2>代理管理</h2>
      <el-button type="primary" @click="showAddDialog">添加代理</el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="proxies"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="port" label="端口" width="100" />
      <el-table-column prop="protocol" label="协议" width="120">
        <template #default="{ row }">
          <el-tag :type="getProtocolType(row.protocol)">
            {{ row.protocol }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="enabled" label="状态" width="100">
        <template #default="{ row }">
          <el-switch
            v-model="row.enabled"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="last_active_at" label="最后活动" width="180">
        <template #default="{ row }">
          {{ formatDate(row.last_active_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button-group>
            <el-button
              size="small"
              type="primary"
              @click="showEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              type="danger"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '添加代理' : '编辑代理'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="端口" prop="port">
          <el-input-number
            v-model="form.port"
            :min="1"
            :max="65535"
            :step="1"
          />
        </el-form-item>
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="form.protocol">
            <el-option label="VMess" value="vmess" />
            <el-option label="VLESS" value="vless" />
            <el-option label="Trojan" value="trojan" />
            <el-option label="Shadowsocks" value="shadowsocks" />
          </el-select>
        </el-form-item>
        <el-form-item
          v-if="form.protocol === 'vmess'"
          label="UUID"
          prop="settings.vmess.id"
        >
          <el-input v-model="form.settings.vmess.id" />
        </el-form-item>
        <el-form-item
          v-if="form.protocol === 'vless'"
          label="UUID"
          prop="settings.vless.id"
        >
          <el-input v-model="form.settings.vless.id" />
        </el-form-item>
        <el-form-item
          v-if="form.protocol === 'trojan'"
          label="密码"
          prop="settings.trojan.password"
        >
          <el-input v-model="form.settings.trojan.password" />
        </el-form-item>
        <el-form-item
          v-if="form.protocol === 'shadowsocks'"
          label="密码"
          prop="settings.shadowsocks.password"
        >
          <el-input v-model="form.settings.shadowsocks.password" />
        </el-form-item>
        <el-form-item
          v-if="form.protocol === 'shadowsocks'"
          label="加密方式"
          prop="settings.shadowsocks.method"
        >
          <el-select v-model="form.settings.shadowsocks.method">
            <el-option label="AES-256-GCM" value="aes-256-gcm" />
            <el-option label="AES-128-GCM" value="aes-128-gcm" />
            <el-option label="ChaCha20-Poly1305" value="chacha20-poly1305" />
          </el-select>
        </el-form-item>
        <el-form-item label="传输协议" prop="stream_settings.network">
          <el-select v-model="form.stream_settings.network">
            <el-option label="TCP" value="tcp" />
            <el-option label="WebSocket" value="ws" />
            <el-option label="HTTP/2" value="http" />
          </el-select>
        </el-form-item>
        <el-form-item
          v-if="form.stream_settings.network === 'ws'"
          label="WebSocket路径"
          prop="stream_settings.ws_settings.path"
        >
          <el-input v-model="form.stream_settings.ws_settings.path" />
        </el-form-item>
        <el-form-item
          v-if="form.stream_settings.network === 'http'"
          label="HTTP路径"
          prop="stream_settings.http_settings.path"
        >
          <el-input v-model="form.stream_settings.http_settings.path" />
        </el-form-item>
        <el-form-item label="TLS" prop="stream_settings.security">
          <el-switch v-model="form.stream_settings.security" />
        </el-form-item>
        <el-form-item
          v-if="form.stream_settings.security"
          label="TLS证书"
          prop="stream_settings.tls_settings.cert"
        >
          <el-select v-model="form.stream_settings.tls_settings.cert">
            <el-option
              v-for="cert in certificates"
              :key="cert.id"
              :label="cert.domain"
              :value="cert.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { formatDate } from '@/utils/date'
import api from '@/api'

const userStore = useUserStore()
const loading = ref(false)
const proxies = ref([])
const certificates = ref([])
const dialogVisible = ref(false)
const dialogType = ref('add')
const formRef = ref(null)
const form = ref({
  port: 443,
  protocol: 'vmess',
  settings: {
    vmess: {
      id: '',
      alter_id: 0,
      security: 'auto'
    },
    vless: {
      id: '',
      flow: '',
      security: 'none'
    },
    trojan: {
      password: ''
    },
    shadowsocks: {
      method: 'aes-256-gcm',
      password: ''
    }
  },
  stream_settings: {
    network: 'tcp',
    security: false,
    tls_settings: {
      cert: '',
      allow_insecure: false,
      server_name: ''
    },
    ws_settings: {
      path: '/ws'
    },
    http_settings: {
      path: '/http'
    }
  }
})

const rules = {
  port: [
    { required: true, message: '请输入端口', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口范围1-65535', trigger: 'blur' }
  ],
  protocol: [
    { required: true, message: '请选择协议', trigger: 'change' }
  ],
  'settings.vmess.id': [
    { required: true, message: '请输入UUID', trigger: 'blur' }
  ],
  'settings.vless.id': [
    { required: true, message: '请输入UUID', trigger: 'blur' }
  ],
  'settings.trojan.password': [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  'settings.shadowsocks.password': [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  'stream_settings.network': [
    { required: true, message: '请选择传输协议', trigger: 'change' }
  ]
}

const getProtocolType = (protocol) => {
  const types = {
    vmess: 'success',
    vless: 'warning',
    trojan: 'danger',
    shadowsocks: 'info'
  }
  return types[protocol] || 'info'
}

const fetchProxies = async () => {
  loading.value = true
  try {
    const response = await api.get('/proxies')
    proxies.value = response.data
  } catch (error) {
    ElMessage.error('获取代理列表失败')
  } finally {
    loading.value = false
  }
}

const fetchCertificates = async () => {
  try {
    const response = await api.get('/certificates')
    certificates.value = response.data
  } catch (error) {
    ElMessage.error('获取证书列表失败')
  }
}

const showAddDialog = () => {
  dialogType.value = 'add'
  form.value = {
    port: 443,
    protocol: 'vmess',
    settings: {
      vmess: {
        id: '',
        alter_id: 0,
        security: 'auto'
      },
      vless: {
        id: '',
        flow: '',
        security: 'none'
      },
      trojan: {
        password: ''
      },
      shadowsocks: {
        method: 'aes-256-gcm',
        password: ''
      }
    },
    stream_settings: {
      network: 'tcp',
      security: false,
      tls_settings: {
        cert: '',
        allow_insecure: false,
        server_name: ''
      },
      ws_settings: {
        path: '/ws'
      },
      http_settings: {
        path: '/http'
      }
    }
  }
  dialogVisible.value = true
}

const showEditDialog = (proxy) => {
  dialogType.value = 'edit'
  form.value = { ...proxy }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (dialogType.value === 'add') {
          await api.post('/proxies', form.value)
          ElMessage.success('添加代理成功')
        } else {
          await api.put(`/proxies/${form.value.id}`, form.value)
          ElMessage.success('更新代理成功')
        }
        dialogVisible.value = false
        fetchProxies()
      } catch (error) {
        ElMessage.error('操作失败')
      }
    }
  })
}

const handleDelete = async (proxy) => {
  try {
    await ElMessageBox.confirm('确定要删除该代理吗？', '提示', {
      type: 'warning'
    })
    await api.delete(`/proxies/${proxy.id}`)
    ElMessage.success('删除代理成功')
    fetchProxies()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除代理失败')
    }
  }
}

const handleStatusChange = async (proxy) => {
  try {
    if (proxy.enabled) {
      await api.post(`/proxies/${proxy.id}/start`)
      ElMessage.success('启动代理成功')
    } else {
      await api.post(`/proxies/${proxy.id}/stop`)
      ElMessage.success('停止代理成功')
    }
  } catch (error) {
    proxy.enabled = !proxy.enabled
    ElMessage.error('操作失败')
  }
}

onMounted(() => {
  fetchProxies()
  fetchCertificates()
})
</script>

<style scoped>
.proxies {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style> 