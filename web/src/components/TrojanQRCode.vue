<template>
  <div class="trojan-qrcode-container">
    <div v-if="loading" class="qrcode-loading">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>生成二维码中...</span>
    </div>
    <div v-else-if="error" class="qrcode-error">
      <el-icon><Warning /></el-icon>
      <span>{{ error }}</span>
    </div>
    <div v-else class="qrcode-content">
      <div class="protocol-header">
        <span class="protocol-badge">TROJAN</span>
        <span class="protocol-name">{{ connectionName }}</span>
      </div>
      
      <div class="qrcode-display">
        <img v-if="qrCodeImage" :src="qrCodeImage" alt="Trojan QR Code" />
        <div v-else ref="qrcodeTarget" class="qrcode-element"></div>
      </div>
      
      <div class="connection-details">
        <div class="connection-item">
          <span class="item-label">服务器</span>
          <span class="item-value">{{ server }}</span>
        </div>
        <div class="connection-item">
          <span class="item-label">端口</span>
          <span class="item-value">{{ port }}</span>
        </div>
        <div class="connection-item">
          <span class="item-label">密码</span>
          <span class="item-value">
            <el-input 
              :value="maskPassword(password)" 
              readonly 
              size="small"
              class="password-input"
            >
              <template #append>
                <el-button @click="togglePasswordVisible">
                  <el-icon v-if="passwordVisible"><View /></el-icon>
                  <el-icon v-else><Hide /></el-icon>
                </el-button>
              </template>
            </el-input>
          </span>
        </div>
        <div class="connection-item">
          <span class="item-label">传输方式</span>
          <span class="item-value">{{ transport.toUpperCase() }}</span>
        </div>
      </div>
      
      <div class="share-link">
        <el-input 
          v-model="shareLink" 
          readonly 
          size="small"
          class="share-link-input"
        >
          <template #append>
            <el-button type="primary" @click="copyLink">
              <el-icon><DocumentCopy /></el-icon>
            </el-button>
          </template>
        </el-input>
      </div>
      
      <div class="actions">
        <el-button type="primary" @click="downloadQRCode">
          <el-icon class="el-icon--left"><Download /></el-icon>
          下载二维码
        </el-button>
        <el-button @click="regenerateQRCode">
          <el-icon class="el-icon--left"><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, Warning, View, Hide, DocumentCopy, Download, Refresh } from '@element-plus/icons-vue'
import QRCode from 'qrcode'

const props = defineProps({
  // 连接名称
  connectionName: {
    type: String,
    default: 'Trojan 连接'
  },
  // 服务器地址
  server: {
    type: String,
    required: true
  },
  // 端口
  port: {
    type: [Number, String],
    default: 443
  },
  // 密码
  password: {
    type: String,
    required: true
  },
  // 安全类型 (tls/xtls)
  security: {
    type: String,
    default: 'tls'
  },
  // SNI
  sni: {
    type: String,
    default: ''
  },
  // 传输方式 (tcp/ws/http/grpc)
  transport: {
    type: String,
    default: 'tcp'
  },
  // WebSocket 路径
  wsPath: {
    type: String,
    default: ''
  },
  // WebSocket 主机
  wsHost: {
    type: String,
    default: ''
  },
  // HTTP/2 路径
  httpPath: {
    type: String,
    default: ''
  },
  // HTTP/2 主机
  httpHost: {
    type: String,
    default: ''
  },
  // gRPC 服务名称
  grpcServiceName: {
    type: String,
    default: ''
  },
  // 跳过证书验证
  allowInsecure: {
    type: Boolean,
    default: false
  },
  // 自定义备注
  remark: {
    type: String,
    default: ''
  }
})

// 状态
const loading = ref(false)
const error = ref('')
const qrCodeImage = ref('')
const qrcodeTarget = ref(null)
const passwordVisible = ref(false)

// 计算分享链接
const shareLink = computed(() => {
  // 基础URL格式: trojan://password@server:port
  let link = `trojan://${encodeURIComponent(props.password)}@${props.server}:${props.port}`
  
  // 参数对象
  const params = new URLSearchParams()
  
  // 添加安全类型 - Trojan 默认使用 TLS
  if (props.security && props.security !== 'tls') {
    params.append('security', props.security)
  }
  
  // 添加SNI
  if (props.sni) {
    params.append('sni', props.sni)
  }
  
  // 添加传输方式 (非TCP需要添加参数)
  if (props.transport !== 'tcp') {
    params.append('type', props.transport)
    
    // WebSocket特有参数
    if (props.transport === 'ws') {
      if (props.wsPath) params.append('path', props.wsPath)
      if (props.wsHost) params.append('host', props.wsHost)
    }
    
    // HTTP/2特有参数
    else if (props.transport === 'http') {
      if (props.httpPath) params.append('path', props.httpPath)
      if (props.httpHost) params.append('host', props.httpHost)
    }
    
    // gRPC特有参数
    else if (props.transport === 'grpc') {
      if (props.grpcServiceName) params.append('serviceName', props.grpcServiceName)
    }
  }
  
  // 添加跳过证书验证
  if (props.allowInsecure) {
    params.append('allowInsecure', '1')
  }
  
  // 构建URL参数部分
  const queryString = params.toString()
  if (queryString) {
    link += `?${queryString}`
  }
  
  // 添加备注
  const commentText = props.remark || props.connectionName
  if (commentText) {
    link += `#${encodeURIComponent(commentText)}`
  }
  
  return link
})

// 密码掩码
const maskPassword = (password) => {
  if (!password) return ''
  if (passwordVisible.value) return password
  
  if (password.length <= 4) {
    return '•'.repeat(password.length)
  } else {
    return password.substring(0, 2) + '•'.repeat(password.length - 4) + password.substring(password.length - 2)
  }
}

// 切换密码可见性
const togglePasswordVisible = () => {
  passwordVisible.value = !passwordVisible.value
}

// 生成二维码
const generateQRCode = async () => {
  if (!shareLink.value) {
    error.value = '无法生成Trojan链接，请检查必填参数'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    // 生成二维码
    const dataUrl = await QRCode.toDataURL(shareLink.value, {
      errorCorrectionLevel: 'M',
      margin: 4,
      width: 256,
      color: {
        dark: '#000000',
        light: '#ffffff'
      }
    })
    
    qrCodeImage.value = dataUrl
  } catch (err) {
    console.error('生成二维码失败:', err)
    error.value = '生成二维码失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

// 复制链接
const copyLink = async () => {
  if (!shareLink.value) {
    ElMessage.error('无可用的分享链接')
    return
  }
  
  try {
    await navigator.clipboard.writeText(shareLink.value)
    ElMessage.success('链接已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
    ElMessage.error('复制失败，请手动复制')
  }
}

// 下载二维码
const downloadQRCode = () => {
  if (!qrCodeImage.value) {
    ElMessage.error('二维码图片未生成')
    return
  }
  
  const link = document.createElement('a')
  link.href = qrCodeImage.value
  link.download = `trojan_${props.server}_${props.port}.png`
  link.click()
}

// 重新生成二维码
const regenerateQRCode = () => {
  generateQRCode()
}

// 监听属性变化
watch(
  () => [
    props.server,
    props.port,
    props.password,
    props.security,
    props.sni,
    props.transport,
    props.wsPath,
    props.wsHost,
    props.httpPath,
    props.httpHost,
    props.grpcServiceName,
    props.allowInsecure,
    props.remark,
    props.connectionName
  ],
  () => {
    generateQRCode()
  }
)

// 组件挂载时生成二维码
onMounted(() => {
  generateQRCode()
})
</script>

<style scoped>
.trojan-qrcode-container {
  width: 100%;
  padding: 16px;
  border-radius: 8px;
  background-color: #fff;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.qrcode-loading,
.qrcode-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 250px;
  gap: 16px;
  color: #909399;
}

.qrcode-error {
  color: #F56C6C;
}

.qrcode-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.protocol-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  width: 100%;
  justify-content: center;
}

.protocol-badge {
  background-color: #e6a23c;
  color: white;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.protocol-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.qrcode-display {
  padding: 16px;
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  background-color: #fff;
}

.qrcode-display img {
  max-width: 100%;
  height: auto;
}

.qrcode-element {
  min-width: 240px;
  min-height: 240px;
}

.connection-details {
  width: 100%;
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.connection-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-label {
  font-weight: 500;
  color: #606266;
}

.item-value {
  font-family: monospace;
  color: #303133;
}

.password-input {
  width: 220px;
  font-family: monospace;
}

.share-link {
  width: 100%;
}

.share-link-input {
  font-family: monospace;
  font-size: 12px;
}

.actions {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 8px;
}

@media (max-width: 480px) {
  .connection-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  
  .item-value {
    width: 100%;
  }
  
  .password-input {
    width: 100%;
  }
  
  .actions {
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }
  
  .actions .el-button {
    width: 100%;
  }
}
</style> 