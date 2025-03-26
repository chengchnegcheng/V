<template>
  <div class="protocol-manager-container">
    <div class="page-header">
      <div class="title">协议管理</div>
      <el-button type="primary" class="add-btn" @click="openAddProtocolDialog">
        <el-icon class="el-icon--left"><Plus /></el-icon> 添加协议
      </el-button>
    </div>

    <!-- Protocol List Table -->
    <el-table 
      v-loading="loading"
      :data="protocols"
      border
      style="width: 100%"
      stripe
    >
      <el-table-column prop="name" label="协议名称" min-width="150">
        <template #default="{ row }">
          <el-tooltip
            effect="dark"
            :content="row.description"
            placement="top"
          >
            <span class="protocol-name">{{ row.name }}</span>
          </el-tooltip>
        </template>
      </el-table-column>
      
      <el-table-column label="协议类型" width="120" align="center">
        <template #default="{ row }">
          <span :class="['protocol-tag', row.type]">
            {{ row.type.toUpperCase() }}
          </span>
        </template>
      </el-table-column>
      
      <el-table-column prop="server" label="服务器" min-width="180" />
      
      <el-table-column prop="port" label="端口" width="100" align="center" />
      
      <el-table-column label="传输方式" width="120" align="center">
        <template #default="{ row }">
          <span class="transport-tag">{{ row.transport.toUpperCase() }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="TLS" width="80" align="center">
        <template #default="{ row }">
          <el-tag 
            :type="row.security === 'tls' ? 'success' : 'info'"
            size="small"
          >
            {{ row.security.toUpperCase() }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <span :class="['status-tag', row.enabled ? 'running' : 'stopped']">
            {{ row.enabled ? '已启用' : '已禁用' }}
          </span>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" min-width="320" fixed="right">
        <template #default="{ row }">
          <div class="operation-btns">
            <el-button
              size="small"
              type="primary"
              @click="copyShareLink(row)"
            >
              分享链接
            </el-button>
            
            <el-button
              size="small"
              :type="row.enabled ? 'warning' : 'success'"
              @click="toggleProtocolStatus(row)"
            >
              {{ row.enabled ? '禁用' : '启用' }}
            </el-button>
            
            <el-button
              size="small"
              type="info"
              @click="editProtocol(row)"
            >
              编辑
            </el-button>
            
            <el-button
              size="small"
              type="danger"
              @click="deleteProtocol(row)"
            >
              删除
            </el-button>
            
            <el-button
              size="small"
              type="primary"
              @click="showQrCode(row)"
            >
              二维码
            </el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- Add/Edit Protocol Dialog -->
    <el-dialog
      v-model="protocolDialogVisible"
      :title="isEditing ? '编辑协议' : '添加协议'"
      width="600px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form
        ref="protocolFormRef"
        :model="protocolForm"
        :rules="formRules"
        label-width="120px"
        label-position="left"
      >
        <el-form-item label="协议类型" prop="type">
          <el-select v-model="protocolForm.type" style="width: 100%">
            <el-option label="Trojan" value="trojan" />
            <el-option label="VMess" value="vmess" />
            <el-option label="VLESS" value="vless" />
            <el-option label="Shadowsocks" value="shadowsocks" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="名称" prop="name">
          <el-input v-model="protocolForm.name" placeholder="请输入协议名称" />
        </el-form-item>
        
        <el-form-item label="服务器" prop="server">
          <el-input v-model="protocolForm.server" placeholder="服务器地址" />
        </el-form-item>
        
        <el-form-item label="端口" prop="port">
          <el-input-number 
            v-model="protocolForm.port" 
            :min="1" 
            :max="65535" 
            style="width: 100%" 
            controls-position="right"
          />
        </el-form-item>
        
        <!-- Trojan Specific Settings -->
        <template v-if="protocolForm.type === 'trojan'">
          <el-divider content-position="left">Trojan 协议设置</el-divider>
          
          <el-form-item label="密码" prop="password">
            <el-input v-model="protocolForm.password" placeholder="请输入密码" type="password" show-password>
              <template #append>
                <el-button @click="generateRandomPassword">随机</el-button>
              </template>
            </el-input>
          </el-form-item>
          
          <el-form-item label="TLS 安全" prop="security">
            <el-select v-model="protocolForm.security" style="width: 100%">
              <el-option label="TLS" value="tls" />
              <el-option label="XTLS" value="xtls" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="SNI" prop="sni">
            <el-input v-model="protocolForm.sni" placeholder="服务器名称指示" />
          </el-form-item>
          
          <el-form-item label="传输方式" prop="transport">
            <el-select v-model="protocolForm.transport" style="width: 100%">
              <el-option label="TCP" value="tcp" />
              <el-option label="WebSocket" value="ws" />
              <el-option label="HTTP/2" value="http" />
              <el-option label="gRPC" value="grpc" />
            </el-select>
          </el-form-item>
          
          <!-- WebSocket 特有设置 -->
          <template v-if="protocolForm.transport === 'ws'">
            <el-form-item label="WebSocket 路径" prop="wsPath">
              <el-input v-model="protocolForm.wsPath" placeholder="/trojan-ws" />
            </el-form-item>
            
            <el-form-item label="WebSocket 主机" prop="wsHost">
              <el-input v-model="protocolForm.wsHost" placeholder="example.com" />
            </el-form-item>
          </template>
          
          <!-- HTTP/2 特有设置 -->
          <template v-if="protocolForm.transport === 'http'">
            <el-form-item label="HTTP 主机" prop="httpHost">
              <el-input v-model="protocolForm.httpHost" placeholder="example.com" />
            </el-form-item>
            
            <el-form-item label="HTTP 路径" prop="httpPath">
              <el-input v-model="protocolForm.httpPath" placeholder="/trojan-h2" />
            </el-form-item>
          </template>
          
          <!-- gRPC 特有设置 -->
          <template v-if="protocolForm.transport === 'grpc'">
            <el-form-item label="服务名称" prop="grpcServiceName">
              <el-input v-model="protocolForm.grpcServiceName" placeholder="trojan-grpc-service" />
            </el-form-item>
          </template>
          
          <el-form-item label="跳过证书验证" prop="allowInsecure">
            <el-switch v-model="protocolForm.allowInsecure" />
          </el-form-item>
        </template>
        
        <!-- VMess Specific Settings -->
        <template v-if="protocolForm.type === 'vmess'">
          <el-divider content-position="left">VMess 协议设置</el-divider>
          <!-- VMess specific settings would go here -->
        </template>
        
        <!-- VLESS Specific Settings -->
        <template v-if="protocolForm.type === 'vless'">
          <el-divider content-position="left">VLESS 协议设置</el-divider>
          <!-- VLESS specific settings would go here -->
        </template>
        
        <!-- Shadowsocks Specific Settings -->
        <template v-if="protocolForm.type === 'shadowsocks'">
          <el-divider content-position="left">Shadowsocks 协议设置</el-divider>
          <!-- Shadowsocks specific settings would go here -->
        </template>
        
        <el-divider content-position="left">高级设置</el-divider>
        
        <el-form-item label="备注" prop="remark">
          <el-input v-model="protocolForm.remark" placeholder="备注信息" />
        </el-form-item>
        
        <el-form-item label="启用协议">
          <el-switch v-model="protocolForm.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div style="text-align: right">
          <el-button @click="protocolDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveProtocol" :loading="saving">保存</el-button>
        </div>
      </template>
    </el-dialog>
    
    <!-- QR Code Dialog -->
    <el-dialog
      v-model="qrCodeDialogVisible"
      title="协议二维码"
      width="350px"
      destroy-on-close
      :close-on-click-modal="true"
    >
      <div class="qrcode-container">
        <div v-if="selectedProtocol" class="qrcode-header">
          <span class="protocol-type">{{ selectedProtocol.type.toUpperCase() }}</span>
          <span class="protocol-name">{{ selectedProtocol.name }}</span>
        </div>
        <div id="qrcode-display" class="qrcode"></div>
        <div class="share-link">
          <el-input 
            v-model="shareLink" 
            readonly 
            size="small"
          >
            <template #append>
              <el-button type="primary" size="small" @click="copyShareLink">
                <el-icon><Document /></el-icon>
              </el-button>
            </template>
          </el-input>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Document } from '@element-plus/icons-vue'
import QRCode from 'qrcode'

// 协议列表
const protocols = ref([])
const loading = ref(false)
const saving = ref(false)

// 对话框控制
const protocolDialogVisible = ref(false)
const qrCodeDialogVisible = ref(false)
const isEditing = ref(false)

// 表单引用
const protocolFormRef = ref(null)

// 选中的协议和分享链接
const selectedProtocol = ref(null)
const shareLink = ref('')

// 协议表单数据
const protocolForm = reactive({
  id: null,
  type: 'trojan',
  name: '',
  server: '',
  port: 443,
  password: '',
  security: 'tls',
  sni: '',
  transport: 'tcp',
  wsPath: '',
  wsHost: '',
  httpPath: '',
  httpHost: '',
  grpcServiceName: '',
  allowInsecure: false,
  remark: '',
  enabled: true
})

// 表单验证规则
const formRules = {
  type: [{ required: true, message: '请选择协议类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入协议名称', trigger: 'blur' }],
  server: [{ required: true, message: '请输入服务器地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口号', trigger: 'change' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  security: [{ required: true, message: '请选择安全类型', trigger: 'change' }],
  sni: [{ required: true, message: '请输入SNI', trigger: 'blur' }],
  transport: [{ required: true, message: '请选择传输方式', trigger: 'change' }]
}

// 加载协议列表
const loadProtocols = async () => {
  loading.value = true
  
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    
    // 模拟数据
    protocols.value = [
      {
        id: 1,
        type: 'trojan',
        name: 'Trojan-TCP',
        server: 'trojan1.sh-crystalcg.com',
        port: 60506,
        password: '123456',
        security: 'tls',
        sni: 'trojan1.sh-crystalcg.com',
        transport: 'tcp',
        headerType: 'none',
        wsPath: '',
        wsHost: '',
        remark: 'Trojan TCP连接',
        enabled: true
      },
      {
        id: 2,
        type: 'trojan',
        name: 'Trojan-WebSocket',
        server: 'trojan2.sh-crystalcg.com',
        port: 443,
        password: 'password123',
        security: 'tls',
        sni: 'trojan2.sh-crystalcg.com',
        transport: 'ws',
        wsPath: '/trojan-ws',
        wsHost: 'trojan2.sh-crystalcg.com',
        remark: 'Trojan WebSocket连接',
        enabled: true
      },
      {
        id: 3,
        type: 'vmess',
        name: 'VMess-TCP',
        server: 'vmess.example.com',
        port: 8443,
        security: 'tls',
        transport: 'tcp',
        remark: 'VMess TCP连接',
        enabled: false
      }
    ]
  } catch (error) {
    console.error('加载协议列表失败:', error)
    ElMessage.error('加载协议列表失败')
  } finally {
    loading.value = false
  }
}

// 打开添加协议对话框
const openAddProtocolDialog = () => {
  isEditing.value = false
  resetForm()
  protocolDialogVisible.value = true
}

// 编辑协议
const editProtocol = (row) => {
  isEditing.value = true
  Object.assign(protocolForm, row)
  protocolDialogVisible.value = true
}

// 删除协议
const deleteProtocol = (row) => {
  ElMessageBox.confirm(
    `确定要删除协议 "${row.name}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(async () => {
      // 模拟删除操作
      await new Promise(resolve => setTimeout(resolve, 500))
      protocols.value = protocols.value.filter(p => p.id !== row.id)
      ElMessage.success('删除成功')
    })
    .catch(() => {
      ElMessage.info('已取消删除')
    })
}

// 重置表单
const resetForm = () => {
  // 重置为默认值
  protocolForm.id = null
  protocolForm.type = 'trojan'
  protocolForm.name = ''
  protocolForm.server = ''
  protocolForm.port = 443
  protocolForm.password = ''
  protocolForm.security = 'tls'
  protocolForm.sni = ''
  protocolForm.transport = 'tcp'
  protocolForm.wsPath = ''
  protocolForm.wsHost = ''
  protocolForm.httpPath = ''
  protocolForm.httpHost = ''
  protocolForm.grpcServiceName = ''
  protocolForm.allowInsecure = false
  protocolForm.remark = ''
  protocolForm.enabled = true
  
  // 如果表单引用存在，重置表单验证
  if (protocolFormRef.value) {
    protocolFormRef.value.resetFields()
  }
}

// 生成随机密码
const generateRandomPassword = () => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let password = ''
  for (let i = 0; i < 16; i++) {
    password += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  protocolForm.password = password
}

// 保存协议
const saveProtocol = async () => {
  if (!protocolFormRef.value) return
  
  await protocolFormRef.value.validate(async (valid) => {
    if (!valid) {
      ElMessage.error('请正确填写表单')
      return
    }
    
    saving.value = true
    
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 800))
      
      // 更新或添加协议
      if (isEditing.value) {
        // 编辑现有协议
        const index = protocols.value.findIndex(p => p.id === protocolForm.id)
        if (index !== -1) {
          protocols.value[index] = { ...protocolForm }
        }
        ElMessage.success('协议更新成功')
      } else {
        // 添加新协议
        const newId = protocols.value.length > 0 
          ? Math.max(...protocols.value.map(p => p.id)) + 1 
          : 1
        
        protocols.value.push({
          ...protocolForm,
          id: newId
        })
        ElMessage.success('协议添加成功')
      }
      
      // 关闭对话框
      protocolDialogVisible.value = false
    } catch (error) {
      console.error('保存协议失败:', error)
      ElMessage.error('保存协议失败')
    } finally {
      saving.value = false
    }
  })
}

// 切换协议状态
const toggleProtocolStatus = async (row) => {
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 300))
    
    // 更新状态
    const protocol = protocols.value.find(p => p.id === row.id)
    if (protocol) {
      protocol.enabled = !protocol.enabled
      ElMessage.success(`协议已${protocol.enabled ? '启用' : '禁用'}`)
    }
  } catch (error) {
    console.error('切换协议状态失败:', error)
    ElMessage.error('切换协议状态失败')
  }
}

// 生成协议分享链接
const generateShareLink = (protocol) => {
  if (!protocol) return ''
  
  let link = ''
  
  switch (protocol.type) {
    case 'trojan':
      link = `trojan://${encodeURIComponent(protocol.password)}@${protocol.server}:${protocol.port}`
      
      // 添加参数
      const params = new URLSearchParams()
      
      // Trojan 默认使用 TLS
      if (protocol.security && protocol.security !== 'tls') {
        params.append('security', protocol.security)
      }
      
      if (protocol.sni) {
        params.append('sni', protocol.sni)
      }
      
      if (protocol.transport !== 'tcp') {
        params.append('type', protocol.transport)
        
        if (protocol.transport === 'ws') {
          if (protocol.wsPath) params.append('path', protocol.wsPath)
          if (protocol.wsHost) params.append('host', protocol.wsHost)
        } else if (protocol.transport === 'http') {
          if (protocol.httpPath) params.append('path', protocol.httpPath)
          if (protocol.httpHost) params.append('host', protocol.httpHost)
        } else if (protocol.transport === 'grpc') {
          if (protocol.grpcServiceName) params.append('serviceName', protocol.grpcServiceName)
        }
      }
      
      if (protocol.allowInsecure) {
        params.append('allowInsecure', '1')
      }
      
      // 添加参数和备注
      const queryString = params.toString()
      if (queryString) {
        link += `?${queryString}`
      }
      
      if (protocol.remark || protocol.name) {
        link += `#${encodeURIComponent(protocol.remark || protocol.name)}`
      }
      break
      
    case 'vmess':
      // VMess 使用 JSON 对象，然后 Base64 编码
      const vmessConfig = {
        v: "2",
        ps: protocol.name || protocol.remark || "VMess",
        add: protocol.server,
        port: protocol.port,
        id: protocol.uuid || "",
        aid: protocol.alterId || 0,
        net: protocol.transport || "tcp",
        type: protocol.headerType || "none",
        host: (protocol.transport === 'ws' ? protocol.wsHost : 
              protocol.transport === 'http' ? protocol.httpHost : ""),
        path: (protocol.transport === 'ws' ? protocol.wsPath : 
              protocol.transport === 'http' ? protocol.httpPath : 
              protocol.transport === 'grpc' ? protocol.grpcServiceName : ""),
        tls: protocol.security === 'tls' || protocol.security === 'xtls' ? "tls" : "none",
        sni: protocol.sni || ""
      }
      
      // Base64 编码 JSON 字符串
      link = `vmess://${btoa(JSON.stringify(vmessConfig))}`
      break
      
    case 'vless':
      // VLESS 格式: vless://uuid@server:port?encryption=none&security=tls&type=tcp#remark
      link = `vless://${protocol.uuid || ""}@${protocol.server}:${protocol.port}`
      
      const vlessParams = new URLSearchParams()
      vlessParams.append('encryption', 'none') // VLESS 需要指定 encryption=none
      
      if (protocol.security) {
        vlessParams.append('security', protocol.security)
      }
      
      if (protocol.sni) {
        vlessParams.append('sni', protocol.sni)
      }
      
      if (protocol.transport) {
        vlessParams.append('type', protocol.transport)
        
        if (protocol.transport === 'ws') {
          if (protocol.wsPath) vlessParams.append('path', protocol.wsPath)
          if (protocol.wsHost) vlessParams.append('host', protocol.wsHost)
        } else if (protocol.transport === 'http') {
          if (protocol.httpPath) vlessParams.append('path', protocol.httpPath)
          if (protocol.httpHost) vlessParams.append('host', protocol.httpHost)
        } else if (protocol.transport === 'grpc') {
          if (protocol.grpcServiceName) vlessParams.append('serviceName', protocol.grpcServiceName)
        }
      }
      
      if (protocol.flow) {
        vlessParams.append('flow', protocol.flow)
      }
      
      // 添加参数和备注
      const vlessQueryString = vlessParams.toString()
      if (vlessQueryString) {
        link += `?${vlessQueryString}`
      }
      
      if (protocol.remark || protocol.name) {
        link += `#${encodeURIComponent(protocol.remark || protocol.name)}`
      }
      break
      
    // 其他协议类型的链接生成逻辑
    default:
      link = ''
  }
  
  return link
}

// 显示二维码
const showQrCode = async (row) => {
  selectedProtocol.value = row
  shareLink.value = generateShareLink(row)
  qrCodeDialogVisible.value = true
  
  // 生成二维码
  await nextTick()
  const qrcodeElement = document.getElementById('qrcode-display')
  if (qrcodeElement && shareLink.value) {
    qrcodeElement.innerHTML = ''
    QRCode.toCanvas(qrcodeElement, shareLink.value, {
      width: 250,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#ffffff'
      }
    }).catch(err => {
      console.error('生成二维码失败:', err)
      ElMessage.error('生成二维码失败')
    })
  }
}

// 复制分享链接
const copyShareLink = async (protocol) => {
  const link = protocol ? generateShareLink(protocol) : shareLink.value
  
  if (!link) {
    ElMessage.error('无效的分享链接')
    return
  }
  
  try {
    await navigator.clipboard.writeText(link)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败')
  }
}

// 组件挂载时加载协议列表
onMounted(() => {
  loadProtocols()
})
</script>

<style scoped>
.protocol-manager-container {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.protocol-tag {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  color: white;
}

.protocol-tag.trojan {
  background-color: #e6a23c;
}

.protocol-tag.vmess {
  background-color: #409eff;
}

.protocol-tag.vless {
  background-color: #67c23a;
}

.protocol-tag.shadowsocks {
  background-color: #909399;
}

.status-tag {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  color: white;
}

.status-tag.running {
  background-color: #67c23a;
}

.status-tag.stopped {
  background-color: #909399;
}

.transport-tag {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  background-color: #ecf5ff;
  color: #409eff;
  border: 1px solid #d9ecff;
}

.operation-btns {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.qrcode-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px;
}

.qrcode-header {
  margin-bottom: 16px;
  text-align: center;
}

.qrcode-header .protocol-type {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  color: white;
  background-color: #e6a23c;
  margin-right: 8px;
}

.qrcode-header .protocol-name {
  font-size: 16px;
  font-weight: 600;
}

.qrcode {
  margin: 16px 0;
  padding: 16px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background-color: white;
}

.share-link {
  width: 100%;
  margin-top: 16px;
}

.protocol-name {
  font-weight: 500;
}
</style> 