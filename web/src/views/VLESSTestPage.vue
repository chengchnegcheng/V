<template>
  <div class="test-page-container">
    <div class="page-header">
      <div class="title">VLESS 协议测试</div>
    </div>
    
    <el-card shadow="hover" class="test-card">
      <template #header>
        <div class="card-header">
          <span>测试 VLESS 配置</span>
        </div>
      </template>
      
      <div class="form-container">
        <el-form 
          :model="configForm" 
          label-position="top"
          label-width="100px"
        >
          <el-row :gutter="20">
            <el-col :xs="24" :md="12">
              <el-form-item label="连接名称">
                <el-input v-model="configForm.name" placeholder="请输入连接名称" />
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12">
              <el-form-item label="服务器地址">
                <el-input v-model="configForm.server" placeholder="请输入服务器地址" />
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12">
              <el-form-item label="端口">
                <el-input-number 
                  v-model="configForm.port" 
                  :min="1" 
                  :max="65535" 
                  style="width: 100%" 
                  controls-position="right" 
                />
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12">
              <el-form-item label="UUID">
                <el-input 
                  v-model="configForm.uuid" 
                  placeholder="请输入UUID" 
                  show-password 
                >
                  <template #append>
                    <el-button @click="generateRandomUUID">随机</el-button>
                  </template>
                </el-input>
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12">
              <el-form-item label="安全类型">
                <el-select v-model="configForm.security" style="width: 100%">
                  <el-option label="无加密" value="none" />
                  <el-option label="TLS" value="tls" />
                  <el-option label="XTLS" value="xtls" />
                </el-select>
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12" v-if="configForm.security !== 'none'">
              <el-form-item label="SNI">
                <el-input v-model="configForm.sni" placeholder="请输入SNI" />
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12" v-if="configForm.security === 'xtls'">
              <el-form-item label="流控">
                <el-select v-model="configForm.flow" style="width: 100%">
                  <el-option label="无流控" value="" />
                  <el-option label="xtls-rprx-vision" value="xtls-rprx-vision" />
                  <el-option label="xtls-rprx-vision-udp443" value="xtls-rprx-vision-udp443" />
                </el-select>
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12">
              <el-form-item label="传输方式">
                <el-select v-model="configForm.transport" style="width: 100%">
                  <el-option label="TCP" value="tcp" />
                  <el-option label="WebSocket" value="ws" />
                  <el-option label="HTTP/2" value="http" />
                  <el-option label="gRPC" value="grpc" />
                </el-select>
              </el-form-item>
            </el-col>
            
            <!-- WebSocket 特有设置 -->
            <template v-if="configForm.transport === 'ws'">
              <el-col :xs="24" :md="12">
                <el-form-item label="WebSocket 路径">
                  <el-input v-model="configForm.wsPath" placeholder="/vless-ws" />
                </el-form-item>
              </el-col>
              
              <el-col :xs="24" :md="12">
                <el-form-item label="WebSocket 主机">
                  <el-input v-model="configForm.wsHost" placeholder="example.com" />
                </el-form-item>
              </el-col>
            </template>
            
            <!-- HTTP/2 特有设置 -->
            <template v-if="configForm.transport === 'http'">
              <el-col :xs="24" :md="12">
                <el-form-item label="HTTP/2 路径">
                  <el-input v-model="configForm.httpPath" placeholder="/vless-h2" />
                </el-form-item>
              </el-col>
              
              <el-col :xs="24" :md="12">
                <el-form-item label="HTTP/2 主机">
                  <el-input v-model="configForm.httpHost" placeholder="example.com" />
                </el-form-item>
              </el-col>
            </template>
            
            <!-- gRPC 特有设置 -->
            <template v-if="configForm.transport === 'grpc'">
              <el-col :xs="24" :md="12">
                <el-form-item label="gRPC 服务名称">
                  <el-input v-model="configForm.grpcServiceName" placeholder="vless-grpc-service" />
                </el-form-item>
              </el-col>
            </template>
            
            <el-col :xs="24" :md="12">
              <el-form-item label="跳过证书验证">
                <el-switch v-model="configForm.allowInsecure" />
              </el-form-item>
            </el-col>
            
            <el-col :xs="24" :md="12">
              <el-form-item label="备注">
                <el-input v-model="configForm.remark" placeholder="自定义备注" />
              </el-form-item>
            </el-col>
          </el-row>
          
          <div class="actions">
            <el-button type="primary" @click="applyConfig">应用配置</el-button>
            <el-button @click="resetConfig">重置</el-button>
          </div>
        </el-form>
      </div>
    </el-card>
    
    <el-card shadow="hover" class="test-card">
      <template #header>
        <div class="card-header">
          <span>VLESS 二维码</span>
        </div>
      </template>
      
      <div class="qrcode-container">
        <VLESSQRCode 
          :connection-name="configForm.name"
          :server="configForm.server"
          :port="configForm.port"
          :uuid="configForm.uuid"
          :security="configForm.security"
          :sni="configForm.sni"
          :flow="configForm.flow"
          :transport="configForm.transport"
          :ws-path="configForm.wsPath"
          :ws-host="configForm.wsHost"
          :http-path="configForm.httpPath"
          :http-host="configForm.httpHost"
          :grpc-service-name="configForm.grpcServiceName"
          :allow-insecure="configForm.allowInsecure"
          :remark="configForm.remark"
        />
      </div>
    </el-card>
    
    <el-card shadow="hover" class="test-card">
      <template #header>
        <div class="card-header">
          <span>预设配置</span>
        </div>
      </template>
      
      <div class="presets">
        <el-button @click="loadPreset('tcp')" type="primary">TCP 预设</el-button>
        <el-button @click="loadPreset('ws')" type="success">WebSocket 预设</el-button>
        <el-button @click="loadPreset('http')" type="warning">HTTP/2 预设</el-button>
        <el-button @click="loadPreset('grpc')" type="info">gRPC 预设</el-button>
        <el-button @click="loadPreset('xtls')" type="danger">XTLS 预设</el-button>
        <el-button @click="loadRealConfig">加载真实配置</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import VLESSQRCode from '../components/VLESSQRCode.vue'

// 配置表单
const configForm = reactive({
  name: 'VLESS 测试连接',
  server: 'vless.example.com',
  port: 443,
  uuid: '61c2f4b7-2ecf-47f5-9192-3da14976d5e2',
  security: 'tls',
  sni: 'vless.example.com',
  flow: '',
  transport: 'tcp',
  wsPath: '',
  wsHost: '',
  httpPath: '',
  httpHost: '',
  grpcServiceName: '',
  allowInsecure: false,
  remark: ''
})

// 生成随机UUID
const generateRandomUUID = () => {
  const uuid = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
  configForm.uuid = uuid
}

// 应用配置
const applyConfig = () => {
  // 在这里可以添加配置验证逻辑
  if (!configForm.server) {
    ElMessage.error('请输入服务器地址')
    return
  }
  
  if (!configForm.uuid) {
    ElMessage.error('请输入UUID')
    return
  }
  
  ElMessage.success('配置已应用')
}

// 重置配置
const resetConfig = () => {
  configForm.name = 'VLESS 测试连接'
  configForm.server = 'vless.example.com'
  configForm.port = 443
  configForm.uuid = '61c2f4b7-2ecf-47f5-9192-3da14976d5e2'
  configForm.security = 'tls'
  configForm.sni = 'vless.example.com'
  configForm.flow = ''
  configForm.transport = 'tcp'
  configForm.wsPath = ''
  configForm.wsHost = ''
  configForm.httpPath = ''
  configForm.httpHost = ''
  configForm.grpcServiceName = ''
  configForm.allowInsecure = false
  configForm.remark = ''
  
  ElMessage.info('配置已重置')
}

// 加载预设配置
const loadPreset = (type) => {
  // 基础配置
  configForm.name = `VLESS ${type.toUpperCase()} 测试`
  configForm.server = 'vless.example.com'
  configForm.port = 443
  configForm.uuid = '61c2f4b7-2ecf-47f5-9192-3da14976d5e2'
  configForm.sni = 'vless.example.com'
  configForm.remark = `VLESS ${type.toUpperCase()} 配置`
  configForm.allowInsecure = false
  
  // 传输方式特有配置
  switch (type) {
    case 'tcp':
      configForm.security = 'tls'
      configForm.flow = ''
      configForm.transport = 'tcp'
      configForm.wsPath = ''
      configForm.wsHost = ''
      configForm.httpPath = ''
      configForm.httpHost = ''
      configForm.grpcServiceName = ''
      break
      
    case 'ws':
      configForm.security = 'tls'
      configForm.flow = ''
      configForm.transport = 'ws'
      configForm.wsPath = '/vless-ws'
      configForm.wsHost = 'vless.example.com'
      configForm.httpPath = ''
      configForm.httpHost = ''
      configForm.grpcServiceName = ''
      break
      
    case 'http':
      configForm.security = 'tls'
      configForm.flow = ''
      configForm.transport = 'http'
      configForm.wsPath = ''
      configForm.wsHost = ''
      configForm.httpPath = '/vless-h2'
      configForm.httpHost = 'vless.example.com'
      configForm.grpcServiceName = ''
      break
      
    case 'grpc':
      configForm.security = 'tls'
      configForm.flow = ''
      configForm.transport = 'grpc'
      configForm.wsPath = ''
      configForm.wsHost = ''
      configForm.httpPath = ''
      configForm.httpHost = ''
      configForm.grpcServiceName = 'vless-grpc-service'
      break
      
    case 'xtls':
      configForm.security = 'xtls'
      configForm.flow = 'xtls-rprx-vision'
      configForm.transport = 'tcp'
      configForm.wsPath = ''
      configForm.wsHost = ''
      configForm.httpPath = ''
      configForm.httpHost = ''
      configForm.grpcServiceName = ''
      break
      
    default:
      break
  }
  
  ElMessage.success(`已加载 ${type.toUpperCase()} 预设`)
}

// 加载真实配置样例
const loadRealConfig = () => {
  configForm.name = 'VLESS1'
  configForm.server = 'vless1.example.com'
  configForm.port = 443
  configForm.uuid = '92c46c09-2b67-4a7e-8fdb-2fe56cbe8d1e'
  configForm.security = 'tls'
  configForm.sni = 'vless1.example.com'
  configForm.flow = ''
  configForm.transport = 'ws'
  configForm.wsPath = '/vless'
  configForm.wsHost = 'vless1.example.com'
  configForm.httpPath = ''
  configForm.httpHost = ''
  configForm.grpcServiceName = ''
  configForm.allowInsecure = false
  configForm.remark = 'VLESS1'
  
  ElMessage.success('已加载真实配置')
}
</script>

<style scoped>
.test-page-container {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 20px;
}

.title {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.test-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-size: 16px;
  font-weight: 600;
}

.form-container {
  padding: 0 10px;
}

.actions {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.qrcode-container {
  display: flex;
  justify-content: center;
  padding: 10px;
}

.presets {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
}

@media (max-width: 768px) {
  .presets {
    flex-direction: column;
  }
  
  .presets .el-button {
    width: 100%;
  }
}
</style> 