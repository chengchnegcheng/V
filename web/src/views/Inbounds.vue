<template>
  <div>
    <div class="page-header">
      <div class="title">协议管理</div>
      <el-button type="primary" class="add-btn" @click="openAddInboundDialog">
        <el-icon class="el-icon--left"><Plus /></el-icon> 添加协议
      </el-button>
    </div>
    
    <el-table
      v-loading="loading"
      :data="inbounds"
      border
      style="width: 100%"
      stripe
    >
      <el-table-column prop="remark" label="名称" min-width="120">
        <template #default="{ row }">
          <el-tooltip
            effect="dark"
            :content="row.protocol"
            placement="top"
          >
            <span>{{ row.remark }}</span>
          </el-tooltip>
        </template>
      </el-table-column>
      
      <el-table-column label="协议类型" width="120" align="center">
        <template #default="{ row }">
          <span :class="['protocol-tag', row.protocol]">
            {{ row.protocol.toUpperCase() }}
          </span>
        </template>
      </el-table-column>
      
      <el-table-column prop="port" label="端口" width="100" align="center" />
      
      <el-table-column prop="clientCount" label="用户数" width="100" align="center" />
      
      <el-table-column prop="created_at" label="创建时间" min-width="150" align="center" />
      
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <span :class="['status-tag', row.enable ? 'running' : 'stopped']">
            {{ row.enable ? '运行中' : '已停止' }}
          </span>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" min-width="320" fixed="right">
        <template #default="{ row }">
          <div class="operation-btns">
            <el-button
              size="small"
              type="primary"
              @click="copyLink(row)"
            >
              链接
            </el-button>
            
            <el-button
              size="small"
              :type="row.enable ? 'warning' : 'success'"
              @click="toggleStatus(row)"
            >
              {{ row.enable ? '停止' : '启动' }}
            </el-button>
            
            <el-button
              size="small"
              type="info"
              @click="editInbound(row)"
            >
              编辑
            </el-button>
            
            <el-button
              size="small"
              type="danger"
              @click="deleteInbound(row)"
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

    <!-- 分页控件 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 添加入站对话框 -->
    <el-dialog
      v-model="addInboundDialogVisible"
      title="添加协议"
      width="500px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form
        ref="inboundFormRef"
        :model="inboundForm"
        :rules="rules"
        label-width="100px"
        label-position="left"
      >
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="inboundForm.protocol" style="width: 100%">
            <el-option label="VMess" value="vmess" />
            <el-option label="VLESS" value="vless" />
            <el-option label="Trojan" value="trojan" />
            <el-option label="Shadowsocks" value="shadowsocks" />
            <el-option label="Dokodemo-Door" value="dokodemo-door" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="备注" prop="remark">
          <el-input v-model="inboundForm.remark" placeholder="请输入备注" />
        </el-form-item>
        
        <el-form-item label="IP监听" prop="listen">
          <el-input v-model="inboundForm.listen" placeholder="填空默认使用0.0.0.0" />
        </el-form-item>
        
        <el-form-item label="端口" prop="port">
          <el-input-number 
            v-model="inboundForm.port" 
            :min="1" 
            :max="65535" 
            style="width: 100%" 
            controls-position="right"
          />
          <el-button 
            size="small" 
            type="primary" 
            style="margin-left: 10px" 
            @click="inboundForm.port = generateRandomPort()"
          >
            随机端口
          </el-button>
        </el-form-item>
        
        <!-- VMess 特有设置 -->
        <template v-if="inboundForm.protocol === 'vmess'">
          <el-form-item label="用户ID" prop="vmess_id">
            <el-input v-model="inboundForm.vmess_id" placeholder="UUID格式" />
            <el-button 
              size="small" 
              type="primary" 
              style="margin-left: 10px" 
              @click="inboundForm.vmess_id = generateUUID()"
            >
              随机UUID
            </el-button>
          </el-form-item>
          
          <el-form-item label="额外ID" prop="vmess_aid">
            <el-input-number 
              v-model="inboundForm.vmess_aid" 
              :min="0" 
              :max="65535" 
              style="width: 100%" 
              controls-position="right"
            />
          </el-form-item>
        </template>
        
        <!-- VLESS 特有设置 -->
        <template v-if="inboundForm.protocol === 'vless'">
          <el-form-item label="用户ID" prop="vless_id">
            <el-input v-model="inboundForm.vless_id" placeholder="UUID格式" />
            <el-button 
              size="small" 
              type="primary" 
              style="margin-left: 10px" 
              @click="inboundForm.vless_id = generateUUID()"
            >
              随机UUID
            </el-button>
          </el-form-item>
          
          <el-form-item label="流控" prop="vless_flow">
            <el-select v-model="inboundForm.vless_flow" style="width: 100%">
              <el-option label="无流控" value="none" />
              <el-option label="xtls-rprx-vision" value="xtls-rprx-vision" />
              <el-option label="xtls-rprx-vision-udp443" value="xtls-rprx-vision-udp443" />
            </el-select>
          </el-form-item>
        </template>
        
        <!-- Trojan 特有设置 -->
        <template v-if="inboundForm.protocol === 'trojan'">
          <el-form-item label="密码" prop="trojan_password">
            <el-input v-model="inboundForm.trojan_password" placeholder="请输入密码" />
            <el-button 
              size="small" 
              type="primary" 
              style="margin-left: 10px" 
              @click="generateRandomPassword()"
            >
              随机密码
            </el-button>
          </el-form-item>
          
          <el-form-item label="流控" prop="trojan_flow">
            <el-select v-model="inboundForm.trojan_flow" style="width: 100%">
              <el-option label="无流控" value="none" />
              <el-option label="xtls-rprx-direct" value="xtls-rprx-direct" />
              <el-option label="xtls-rprx-direct-udp443" value="xtls-rprx-direct-udp443" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="回落" prop="trojan_fallbacks">
            <el-button type="primary" @click="addFallback">添加回落</el-button>
            
            <div v-for="(fallback, index) in inboundForm.trojan_fallbacks" :key="index" class="fallback-item">
              <el-form-item label="地址" style="margin-bottom: 0; margin-right: 10px; flex: 1;">
                <el-input v-model="fallback.dest" placeholder="回落地址，例如: 127.0.0.1" />
              </el-form-item>
              
              <el-form-item label="端口" style="margin-bottom: 0; margin-right: 10px; width: 150px;">
                <el-input-number v-model="fallback.port" :min="1" :max="65535" style="width: 100%" />
              </el-form-item>
              
              <el-button type="danger" @click="removeFallback(index)" circle>
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </el-form-item>
        </template>
        
        <!-- Shadowsocks 特有设置 -->
        <template v-if="inboundForm.protocol === 'shadowsocks'">
          <el-form-item label="加密方式" prop="ss_method">
            <el-select v-model="inboundForm.ss_method" style="width: 100%">
              <el-option label="aes-256-gcm" value="aes-256-gcm" />
              <el-option label="aes-128-gcm" value="aes-128-gcm" />
              <el-option label="chacha20-poly1305" value="chacha20-poly1305" />
              <el-option label="chacha20-ietf-poly1305" value="chacha20-ietf-poly1305" />
              <el-option label="none" value="none" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="密码" prop="ss_password">
            <el-input v-model="inboundForm.ss_password" placeholder="请输入密码" />
            <el-button 
              size="small" 
              type="primary" 
              style="margin-left: 10px" 
              @click="generateRandomPassword()"
            >
              随机密码
            </el-button>
          </el-form-item>
        </template>
        
        <!-- Dokodemo-Door 特有设置 -->
        <template v-if="inboundForm.protocol === 'dokodemo-door'">
          <el-form-item label="目标地址" prop="dokodemo_address">
            <el-input v-model="inboundForm.dokodemo_address" placeholder="请输入目标地址" />
          </el-form-item>
          
          <el-form-item label="目标端口" prop="dokodemo_port">
            <el-input-number 
              v-model="inboundForm.dokodemo_port" 
              :min="1" 
              :max="65535" 
              style="width: 100%" 
              controls-position="right"
            />
          </el-form-item>
        </template>
        
        <el-form-item label="网络" prop="network">
          <el-select v-model="inboundForm.network" style="width: 100%">
            <el-option label="TCP+UDP" value="tcp+udp" />
            <el-option label="TCP" value="tcp" />
            <el-option label="UDP" value="udp" />
          </el-select>
        </el-form-item>
        
        <el-divider content-position="left">传输设置</el-divider>
        
        <el-form-item label="传输协议">
          <el-select v-model="inboundForm.stream_settings.network" style="width: 100%">
            <el-option label="TCP" value="tcp" />
            <el-option label="WebSocket" value="ws" />
            <el-option label="HTTP/2" value="http" />
            <el-option label="gRPC" value="grpc" />
          </el-select>
        </el-form-item>
        
        <!-- TCP 设置 -->
        <template v-if="inboundForm.stream_settings.network === 'tcp'">
          <el-form-item label="伪装">
            <el-switch
              v-model="inboundForm.stream_settings.tcp_settings.is_http"
              active-text="HTTP伪装"
              inactive-text="不伪装"
            />
          </el-form-item>
          
          <template v-if="inboundForm.stream_settings.tcp_settings.is_http">
            <el-form-item label="域名">
              <el-input
                v-model="inboundForm.stream_settings.tcp_settings.http_settings.host[0]"
                placeholder="请输入域名，多个域名用逗号分隔"
              />
            </el-form-item>
            
            <el-form-item label="路径">
              <el-input
                v-model="inboundForm.stream_settings.tcp_settings.http_settings.path"
                placeholder="请输入路径，例如: /api"
              />
            </el-form-item>
          </template>
        </template>
        
        <!-- WebSocket 设置 -->
        <template v-if="inboundForm.stream_settings.network === 'ws'">
          <el-form-item label="路径">
            <el-input
              v-model="inboundForm.stream_settings.ws_settings.path"
              placeholder="请输入路径，例如: /ws"
            />
          </el-form-item>
          
          <el-form-item label="域名">
            <el-input
              v-model="inboundForm.stream_settings.ws_settings.host"
              placeholder="请输入域名"
            />
          </el-form-item>
        </template>
        
        <!-- HTTP/2 设置 -->
        <template v-if="inboundForm.stream_settings.network === 'http'">
          <el-form-item label="域名">
            <el-input
              v-model="inboundForm.stream_settings.http_settings.host[0]"
              placeholder="请输入域名，多个域名用逗号分隔"
            />
          </el-form-item>
          
          <el-form-item label="路径">
            <el-input
              v-model="inboundForm.stream_settings.http_settings.path"
              placeholder="请输入路径，例如: /h2"
            />
          </el-form-item>
        </template>
        
        <!-- gRPC 设置 -->
        <template v-if="inboundForm.stream_settings.network === 'grpc'">
          <el-form-item label="服务名称">
            <el-input
              v-model="inboundForm.stream_settings.grpc_settings.service_name"
              placeholder="请输入服务名称"
            />
          </el-form-item>
        </template>
        
        <el-divider content-position="left">TLS 设置</el-divider>
        
        <el-form-item label="TLS">
          <el-switch v-model="tlsEnabled" />
        </el-form-item>
        
        <template v-if="tlsEnabled">
          <el-form-item label="XTLS">
            <el-switch v-model="xtlsEnabled" />
          </el-form-item>
          
          <el-form-item label="域名">
            <el-input
              v-model="inboundForm.stream_settings.tls_settings.server_name"
              placeholder="请输入域名"
            />
          </el-form-item>
          
          <el-form-item label="证书配置">
            <el-radio-group v-model="inboundForm.stream_settings.tls_settings.cert_mode">
              <el-radio value="certificate file path">证书文件路径</el-radio>
              <el-radio value="certificate file content">证书文件内容</el-radio>
            </el-radio-group>
          </el-form-item>
          
          <template v-if="inboundForm.stream_settings.tls_settings.cert_mode === 'certificate file path'">
            <el-form-item label="公钥文件路径">
              <el-input
                v-model="inboundForm.stream_settings.tls_settings.cert_file"
                placeholder="请输入公钥文件路径"
              />
            </el-form-item>
            
            <el-form-item label="私钥文件路径">
              <el-input
                v-model="inboundForm.stream_settings.tls_settings.key_file"
                placeholder="请输入私钥文件路径"
              />
            </el-form-item>
          </template>
          
          <template v-else>
            <el-form-item label="公钥文件内容">
              <el-input
                v-model="inboundForm.stream_settings.tls_settings.cert_content"
                type="textarea"
                rows="4"
                placeholder="请输入公钥文件内容"
              />
            </el-form-item>
            
            <el-form-item label="私钥文件内容">
              <el-input
                v-model="inboundForm.stream_settings.tls_settings.key_content"
                type="textarea"
                rows="4"
                placeholder="请输入私钥文件内容"
              />
            </el-form-item>
          </template>
        </template>
        
        <el-divider content-position="left">高级设置</el-divider>
        
        <el-form-item label="流量">
          <el-input-number
            v-model="inboundForm.total_traffic"
            :min="0"
            :step="1"
            style="width: 100%"
            controls-position="right"
          >
            <template #append>GB</template>
          </el-input-number>
          <div style="font-size: 12px; color: #909399; margin-top: 5px;">
            0表示不限制
          </div>
        </el-form-item>
        
        <el-form-item label="过期时间">
          <el-date-picker
            v-model="inboundForm.expiry_time"
            type="datetime"
            placeholder="请选择过期时间"
            style="width: 100%"
          />
          <div style="font-size: 12px; color: #909399; margin-top: 5px;">
            留空表示不限制
          </div>
        </el-form-item>
        
        <el-form-item label="嗅探">
          <el-switch v-model="inboundForm.sniffing.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div style="text-align: right">
          <el-button @click="addInboundDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveInbound" :loading="submitting">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 二维码对话框 -->
    <el-dialog
      v-model="qrCodeDialogVisible"
      title="分享二维码"
      width="350px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <div class="qrcode-container">
        <div id="qrcode-display" class="qrcode"></div>
        <div class="protocol-name">{{ currentQrCodeInfo?.protocol.toUpperCase() }}</div>
        <div class="remark">{{ currentQrCodeInfo?.remark }}</div>
      </div>
      
      <template #footer>
        <div style="text-align: center">
          <el-button type="primary" @click="downloadQrCode">下载二维码</el-button>
          <el-button @click="qrCodeDialogVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { InfoFilled, Plus, Delete } from '@element-plus/icons-vue'
import axios from 'axios'
import QRCode from 'qrcode'

// 数据表格
const loading = ref(false)
const inbounds = ref([])

// 表单对话框
const addInboundDialogVisible = ref(false)
const inboundFormRef = ref(null)
const submitting = ref(false)

// 二维码相关
const qrCodeDialogVisible = ref(false)
const currentQrCodeInfo = ref(null)
const currentQrCodeLink = ref('')

// 默认表单
const defaultInboundForm = {
  remark: '',
  enable: true,
  protocol: 'vmess',
  listen: '',
  port: null,
  total_traffic: 0,
  expiry_time: '',
  vmess_id: '',  // vmess 特有
  vmess_aid: 0,  // vmess 特有
  vless_id: '',  // vless 特有
  vless_flow: 'none', // vless 特有
  trojan_password: '', // trojan 特有
  trojan_flow: 'none',  // trojan 特有
  trojan_fallbacks: [], // trojan 特有
  ss_method: 'aes-256-gcm', // shadowsocks 特有
  ss_password: '', // shadowsocks 特有
  dokodemo_address: '', // dokodemo-door 特有
  dokodemo_port: null, // dokodemo-door 特有
  network: 'tcp+udp',
  stream_settings: {
    network: 'tcp',
    security: '',
    tcp_settings: {
      is_http: false,
      http_settings: {
        host: [],
        path: '/'
      }
    },
    ws_settings: {
      path: '/',
      host: ''
    },
    http_settings: {
      host: [],
      path: '/'
    },
    grpc_settings: {
      service_name: ''
    },
    tls_settings: {
      server_name: '',
      cert_mode: 'certificate file path',
      cert_file: '',
      key_file: '',
      cert_content: '',
      key_content: ''
    }
  },
  sniffing: {
    enabled: true,
    dest_override: ['http', 'tls', 'quic']
  }
}

// 当前表单
const inboundForm = reactive({...defaultInboundForm})

// TLS开关
const tlsEnabled = computed({
  get: () => inboundForm.stream_settings.security === 'tls',
  set: (value) => {
    inboundForm.stream_settings.security = value ? 'tls' : ''
  }
})

// XTLS开关
const xtlsEnabled = computed({
  get: () => inboundForm.stream_settings.security === 'xtls',
  set: (value) => {
    inboundForm.stream_settings.security = value ? 'xtls' : 'tls'
  }
})

// 表单验证规则
const rules = {
  remark: [
    { required: true, message: '请输入备注', trigger: 'blur' }
  ],
  protocol: [
    { required: true, message: '请选择协议', trigger: 'change' }
  ],
  port: [
    { required: true, message: '请输入端口', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口范围 1-65535', trigger: 'blur' }
  ],
  ss_method: [
    { required: true, message: '请选择加密方式', trigger: 'change' }
  ],
  ss_password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  trojan_password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  vmess_id: [
    { required: true, message: '请输入ID', trigger: 'blur' }
  ],
  vless_id: [
    { required: true, message: '请输入ID', trigger: 'blur' }
  ],
  dokodemo_address: [
    { required: true, message: '请输入目标地址', trigger: 'blur' }
  ],
  dokodemo_port: [
    { required: true, message: '请输入目标端口', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口范围 1-65535', trigger: 'blur' }
  ]
}

// 添加运行时验证
const validateTrojanForm = () => {
  if (inboundForm.protocol === 'trojan') {
    // Trojan协议必须启用TLS
    tlsEnabled.value = true;
    
    // 如果没有设置服务器名称，使用默认值
    if (!inboundForm.stream_settings.tls_settings.server_name) {
      inboundForm.stream_settings.tls_settings.server_name = 'example.com';
    }
  }
}

// 监听协议变化
watch(() => inboundForm.protocol, (newVal) => {
  validateTrojanForm();
})

// 分页相关
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 处理分页
const handleSizeChange = (size) => {
  pageSize.value = size
  loadInbounds()
}

const handleCurrentChange = (page) => {
  currentPage.value = page
  loadInbounds()
}

// 初始化
onMounted(() => {
  loadInbounds()
})

// 加载入站列表
const loadInbounds = async () => {
  loading.value = true
  try {
    // 模拟API响应
    const mockData = {
      total: 3,
      list: [
        {
          id: 1,
          remark: 'VIP用户',
          protocol: 'vmess',
          port: 10086,
          clientCount: 5,
          created_at: '2023-03-15 08:30:00',
          enable: true
        },
        {
          id: 2,
          remark: '免费用户',
          protocol: 'vless',
          port: 20086,
          clientCount: 10,
          created_at: '2023-03-10 14:20:00',
          enable: false
        },
        {
          id: 3,
          remark: '测试节点',
          protocol: 'trojan',
          port: 30086,
          clientCount: 2,
          created_at: '2023-03-05 09:15:00',
          enable: true
        }
      ]
    }
    
    // 更新数据
    inbounds.value = mockData.list
    total.value = mockData.total
    
    // 注释掉实际的API调用
    /*
    const response = await axios.get('/api/inbounds', {
      params: {
        page: currentPage.value,
        pageSize: pageSize.value
      }
    })
    
    if (response.data) {
      inbounds.value = response.data.list.map(item => ({
        ...item,
        clientCount: Math.floor(Math.random() * 10) + 1 // 模拟用户数
      }))
      total.value = response.data.total
    }
    */
  } catch (error) {
    console.error('Failed to load inbounds:', error)
    ElMessage.error('加载入站列表失败')
  } finally {
    loading.value = false
  }
}

// 流量格式化
const formatTraffic = (bytes) => {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 打开添加入站对话框
const openAddInboundDialog = () => {
  // 重置表单
  Object.assign(inboundForm, defaultInboundForm)
  // 设置默认端口
  inboundForm.port = generateRandomPort()
  
  // 根据协议类型初始化特定字段
  if (inboundForm.protocol === 'vmess') {
    inboundForm.vmess_id = generateUUID()
  } else if (inboundForm.protocol === 'vless') {
    inboundForm.vless_id = generateUUID()
  } else if (inboundForm.protocol === 'trojan') {
    inboundForm.trojan_password = generateRandomPassword();
    // 确保TLS启用
    tlsEnabled.value = true;
    inboundForm.stream_settings.tls_settings.server_name = 'example.com';
  }
  
  // 验证表单
  validateTrojanForm();
  addInboundDialogVisible.value = true
}

// 添加fallback
const addFallback = () => {
  if (!inboundForm.trojan_fallbacks) {
    inboundForm.trojan_fallbacks = []
  }
  
  inboundForm.trojan_fallbacks.push({
    dest: '',
    port: null
  })
}

// 删除fallback
const removeFallback = (index) => {
  inboundForm.trojan_fallbacks.splice(index, 1)
}

// 保存入站
const saveInbound = async () => {
  if (!inboundFormRef.value) return
  
  await inboundFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      // 准备提交数据
      const submittingData = { ...inboundForm }
      
      // 特殊处理不同协议的配置
      if (inboundForm.protocol === 'trojan') {
        submittingData.settings = {
          password: inboundForm.trojan_password,
          flow: inboundForm.trojan_flow,
          sni: inboundForm.stream_settings.tls_settings.server_name || 'example.com',
          network: inboundForm.stream_settings.network,
          tls: true, // Trojan必须启用TLS
          fallbacks: inboundForm.trojan_fallbacks
        }
        
        // 确保启用TLS
        if (!tlsEnabled.value) {
          tlsEnabled.value = true
        }
      } else if (inboundForm.protocol === 'vmess') {
        // 其他协议配置...
      }
      
      // 模拟提交成功
      setTimeout(() => {
        // 添加到本地数据
        const newInbound = {
          id: inbounds.value.length + 1,
          ...submittingData,
          created_at: new Date().toLocaleString(),
          clientCount: 0
        }
        
        inbounds.value.unshift(newInbound)
        total.value++
        
        ElMessage.success('添加入站成功')
        addInboundDialogVisible.value = false
        submitting.value = false
      }, 800)
      
      // 注释掉实际的API调用
      /*
      const response = await axios.post('/api/inbounds', submittingData)
      
      if (response.data && response.data.success) {
        ElMessage.success('添加入站成功')
        addInboundDialogVisible.value = false
        loadInbounds()
      } else {
        ElMessage.error(response.data.message || '添加入站失败')
      }
      */
    } catch (error) {
      console.error('Failed to save inbound:', error)
      ElMessage.error('添加入站失败: ' + error.message)
      submitting.value = false
    }
  })
}

// 编辑入站
const editInbound = (row) => {
  // 直接赋值
  Object.assign(inboundForm, row)
  addInboundDialogVisible.value = true
}

// 切换状态
const toggleStatus = (row) => {
  const action = row.enable ? '停止' : '启动'
  ElMessageBox.confirm(`确定要${action}入站 "${row.remark}" 吗?`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        // 直接修改本地数据
        row.enable = !row.enable
        ElMessage.success(`${action}入站成功`)
      }, 300)
    } catch (error) {
      console.error(`Failed to ${action} inbound:`, error)
      ElMessage.error(`${action}入站失败: ` + error.message)
    }
  }).catch(() => {
    // 取消操作
  })
}

// 删除入站
const deleteInbound = (row) => {
  ElMessageBox.confirm(`确定要删除入站 "${row.remark}" 吗?`, '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        // 从本地数据中删除
        const index = inbounds.value.findIndex(item => item.id === row.id)
        if (index !== -1) {
          inbounds.value.splice(index, 1)
          total.value--
        }
        ElMessage.success('删除入站成功')
      }, 300)
      
      // 注释掉实际的API调用
      /*
      const response = await axios.delete(`/api/inbounds/${row.id}`)
      if (response.data && response.data.success) {
        ElMessage.success('删除入站成功')
        loadInbounds()
      } else {
        ElMessage.error(response.data.message || '删除入站失败')
      }
      */
    } catch (error) {
      console.error('Failed to delete inbound:', error)
      ElMessage.error('删除入站失败: ' + error.message)
    }
  }).catch(() => {
    // 取消删除
  })
}

// 复制链接
const copyLink = async (row) => {
  try {
    // 获取链接
    let link = '';
    // 使用API获取实际链接
    try {
      const response = await axios.get(`/api/inbounds/${row.id}/link`)
      if (response.data && response.data.link) {
        link = response.data.link;
      } else {
        throw new Error('API返回链接为空');
      }
    } catch (apiError) {
      console.error('API获取链接失败:', apiError);
      // 使用本地生成备用链接
      link = getLocalGeneratedLink(row);
      ElMessage.warning('使用本地生成的链接');
    }
    
    // 确保链接有效
    if (!link) {
      throw new Error('无法生成有效链接');
    }
    
    // 使用更可靠的剪贴板复制方法
    try {
      // 先尝试使用navigator.clipboard
      await navigator.clipboard.writeText(link);
    } catch (clipError) {
      console.error('剪贴板API失败，使用备用方法:', clipError);
      // 备用复制方法
      const textarea = document.createElement('textarea');
      textarea.value = link;
      textarea.style.position = 'fixed';
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
    }
    
    ElMessage.success('链接已复制到剪贴板');
  } catch (error) {
    console.error('复制链接失败:', error);
    ElMessage.error('复制链接失败: ' + error.message);
  }
}

// 生成本地链接(备用)
const getLocalGeneratedLink = (row) => {
  const protocol = row.protocol;
  let link = '';
  
  switch (protocol) {
    case 'vmess':
      link = `vmess://eyJhZGQiOiJleGFtcGxlLmNvbSIsImFpZCI6IjAiLCJpZCI6IjhhZDM4OGZmLThkODItNDE4Yy05YzQ0LWZiYjNhNTgwYzFmYiIsIm5ldCI6InRjcCIsInBhdGgiOiIvIiwicG9ydCI6IiR7cm93LnBvcnR9IiwicHMiOiIke3Jvdy5yZW1hcmt9IiwidGxzIjoiIiwidHlwZSI6Im5vbmUiLCJ2IjoiMiJ9`
      break;
    case 'vless':
      link = `vless://8ad388ff-8d82-418c-9c44-fbb3a580c1fb@example.com:${row.port}?encryption=none&security=tls&type=tcp#${encodeURIComponent(row.remark)}`
      break;
    case 'trojan':
      // 获取设置，如果有的话
      let password = 'password123';
      let sni = 'example.com';
      
      if (row.settings) {
        if (row.settings.password) {
          password = row.settings.password;
        }
        if (row.settings.sni) {
          sni = row.settings.sni;
        }
      }
      
      // 标准Trojan链接格式
      link = `trojan://${encodeURIComponent(password)}@example.com:${row.port}?security=tls&sni=${encodeURIComponent(sni)}#${encodeURIComponent(row.remark)}`
      break;
    default:
      link = `${protocol}://example.com:${row.port}#${encodeURIComponent(row.remark)}`
  }
  
  return link;
}

// 生成随机密码
const generateRandomPassword = () => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = ''
  const length = 16
  
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  
  if (inboundForm.protocol === 'shadowsocks') {
    inboundForm.ss_password = result
  } else if (inboundForm.protocol === 'trojan') {
    inboundForm.trojan_password = result
  }
  
  return result;
}

// 生成随机端口
const generateRandomPort = () => {
  // 生成10000-60000之间的随机端口
  return Math.floor(Math.random() * 50000) + 10000
}

// 生成UUID
const generateUUID = () => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

// 获取分享链接
const getShareLink = async (row) => {
  // 模拟获取链接
  const protocol = row.protocol
  let link = ''
  
  switch (protocol) {
    case 'vmess':
      link = `vmess://eyJhZGQiOiIxMjcuMC4wLjEiLCJhaWQiOiIwIiwiaG9zdCI6ImV4YW1wbGUuY29tIiwiaWQiOiI4YWQzODhmZi04ZDgyLTQxOGMtOWM0NC1mYmIzYTU4MGMxZmIiLCJuZXQiOiJ0Y3AiLCJwYXRoIjoiLyIsInBvcnQiOiIke3Jvdy5wb3J0fSIsInBzIjoiJHtyb3cucmVtYXJrfSIsInRscyI6IiIsInR5cGUiOiJub25lIiwidiI6IjIifQ==`
      break
    case 'vless':
      link = `vless://8ad388ff-8d82-418c-9c44-fbb3a580c1fb@127.0.0.1:${row.port}?encryption=none&security=none&type=tcp&headerType=none#${encodeURIComponent(row.remark)}`
      break
    case 'trojan':
      link = `trojan://password123@example.com:${row.port}?security=tls&sni=example.com&type=tcp#${encodeURIComponent(row.remark)}`
      break
    default:
      link = `${protocol}://127.0.0.1:${row.port}#${encodeURIComponent(row.remark)}`
  }
  
  return link
  
  // 实际的API调用替代方案
  /*
  const response = await axios.get(`/api/inbounds/${row.id}/link`)
  if (response.data && response.data.link) {
    return response.data.link
  } else {
    throw new Error('获取链接失败')
  }
  */
}

// 显示二维码
const showQrCode = async (row) => {
  try {
    // 获取链接
    const link = await getShareLink(row)
    currentQrCodeLink.value = link
    currentQrCodeInfo.value = row
    qrCodeDialogVisible.value = true
    
    // 等待DOM更新
    await nextTick()
    
    // 生成二维码
    const qrElement = document.getElementById('qrcode-display')
    if (qrElement) {
      // 清空已有内容
      qrElement.innerHTML = ''
      
      QRCode.toCanvas(link, {
        width: 256,
        margin: 1,
        color: {
          dark: '#000000',
          light: '#ffffff'
        }
      }).then(canvas => {
        qrElement.appendChild(canvas)
      }).catch(err => {
        console.error('QRCode generation error:', err)
        ElMessage.error('生成二维码失败: ' + err.message)
      })
    }
  } catch (error) {
    console.error('Failed to show QR code:', error)
    ElMessage.error('生成二维码失败: ' + error.message)
  }
}

// 下载二维码
const downloadQrCode = async () => {
  try {
    if (!currentQrCodeInfo.value) return
    
    // 获取QR码 canvas
    const qrCanvas = document.getElementById('qrcode-display')?.querySelector('canvas')
    if (!qrCanvas) {
      ElMessage.error('未找到二维码画布')
      return
    }
    
    // 创建临时canvas
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('无法获取canvas上下文')
    
    // 设置画布大小
    canvas.width = 300
    canvas.height = 350
    
    // 填充白色背景
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, canvas.width, canvas.height)
    
    // 绘制二维码
    ctx.drawImage(qrCanvas, 22, 20, 256, 256)
    
    // 绘制协议名称
    ctx.font = 'bold 16px Arial'
    ctx.fillStyle = '#333333'
    ctx.textAlign = 'center'
    ctx.fillText(currentQrCodeInfo.value.protocol.toUpperCase(), canvas.width / 2, 296)
    
    // 绘制备注
    ctx.font = '14px Arial'
    ctx.fillStyle = '#666666'
    ctx.fillText(currentQrCodeInfo.value.remark, canvas.width / 2, 320)
    
    // 转换为图片并下载
    const link = document.createElement('a')
    link.download = `${currentQrCodeInfo.value.protocol}-${currentQrCodeInfo.value.remark}.png`
    link.href = canvas.toDataURL('image/png')
    link.click()
    
    ElMessage.success('二维码已下载')
  } catch (error) {
    console.error('Failed to download QR code:', error)
    ElMessage.error('下载二维码失败: ' + error.message)
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 10px 16px;
  background-color: white;
  border-radius: 4px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.add-btn {
  font-size: 13px;
  padding: 8px 16px;
}

/* 表格调整 */
:deep(.el-table) {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  border-radius: 4px;
  font-size: 13px;
}

:deep(.el-table th) {
  background-color: #f5f7fa;
  padding: 10px 0;
  font-weight: 500;
  color: #606266;
  font-size: 13px;
}

:deep(.el-table td) {
  padding: 8px 0;
}

:deep(.el-table--border) {
  border: 1px solid #ebeef5;
}

:deep(.el-button--small) {
  padding: 5px 11px;
  font-size: 12px;
}

.protocol-tag {
  display: inline-block;
  padding: 2px 8px;
  font-size: 12px;
  border-radius: 2px;
  color: #fff;
}

.protocol-tag.vmess {
  background-color: #409eff;
}

.protocol-tag.vless {
  background-color: #67c23a;
}

.protocol-tag.trojan {
  background-color: #e6a23c;
}

.protocol-tag.shadowsocks {
  background-color: #f56c6c;
}

.protocol-tag.socks {
  background-color: #909399;
}

.protocol-tag.http {
  background-color: #9254de;
}

.status-tag {
  display: inline-block;
  padding: 2px 8px;
  font-size: 12px;
  border-radius: 2px;
  color: #fff;
}

.status-tag.running {
  background-color: #67c23a;
}

.status-tag.stopped {
  background-color: #f56c6c;
}

.operation-btns {
  display: flex;
  justify-content: space-between;
  flex-wrap: nowrap;
  width: 100%;
}

.operation-btns .el-button {
  margin: 0 2px !important;
  padding: 4px 8px;
  font-size: 12px;
}

:deep(.el-table .operation-btns .el-button) {
  color: #fff;
}

:deep(.el-table .operation-btns .el-button--primary) {
  background-color: #409eff;
  border-color: #409eff;
}

:deep(.el-table .operation-btns .el-button--success) {
  background-color: #67c23a;
  border-color: #67c23a;
}

:deep(.el-table .operation-btns .el-button--warning) {
  background-color: #e6a23c;
  border-color: #e6a23c;
}

:deep(.el-table .operation-btns .el-button--danger) {
  background-color: #f56c6c;
  border-color: #f56c6c;
}

:deep(.el-table .operation-btns .el-button--info) {
  background-color: #909399;
  border-color: #909399;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-pagination) {
  padding: 10px 0;
  margin-right: 10px;
  font-weight: normal;
}

:deep(.el-pagination button) {
  min-width: 28px;
  height: 28px;
}

:deep(.el-pagination .el-select .el-input) {
  width: 100px;
}

:deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background-color: #fafafa;
}

:deep(.el-table .cell) {
  padding: 0 6px;
  line-height: 20px;
  white-space: nowrap;
  overflow: visible;
}

.fallback-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  border: 1px dashed #dcdfe6;
  padding: 10px;
  border-radius: 4px;
}

.qrcode-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 0 15px;
}

.qrcode {
  margin-bottom: 15px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: center;
  align-items: center;
}

.protocol-name {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 5px;
  color: #333;
}

.remark {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

/* 确保表格不会压缩内容 */
:deep(.el-table) {
  width: 100%;
  table-layout: fixed;
}

/* 特别优化操作列样式 */
:deep(.el-table__fixed-right) {
  box-shadow: none;
  height: 100% !important;
}

:deep(.el-table__fixed-right-patch) {
  background-color: #f5f7fa;
}
</style>