<template>
  <div class="protocol-settings-container">
    <div class="page-header">
      <div class="title">系统设置</div>
    </div>
    
    <div class="settings-tabs">
      <div class="tab-header">
        <div class="tab-item">服务器配置</div>
        <div class="tab-item">数据库配置</div>
        <div class="tab-item">日志配置</div>
        <div class="tab-item">Xray内核配置</div>
        <div class="tab-item">管理员配置</div>
        <div class="tab-item">安全设置</div>
        <div class="tab-item active">协议管理</div>
      </div>
      
      <div class="tab-content">
        <div class="settings-section">
          <div class="section-header">支持的协议</div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 Trojan 协议</div>
            <div class="protocol-switch">
              <el-switch v-model="protocols.trojan" />
            </div>
            <div class="protocol-desc">Trojan 协议: 基于 TLS 的轻量级协议，伪装成 HTTPS 流量。</div>
            <div class="protocol-status">
              <span :class="['status-tag', protocols.trojan ? 'enabled' : 'disabled']">
                {{ protocols.trojan ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 VMess 协议</div>
            <div class="protocol-switch">
              <el-switch v-model="protocols.vmess" />
            </div>
            <div class="protocol-desc">VMess 协议: V2Ray 的核心传输协议，支持多种传输层。</div>
            <div class="protocol-status">
              <span :class="['status-tag', protocols.vmess ? 'enabled' : 'disabled']">
                {{ protocols.vmess ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 VLESS 协议</div>
            <div class="protocol-switch">
              <el-switch v-model="protocols.vless" />
            </div>
            <div class="protocol-desc">VLESS 协议: 轻量化的 VMess 协议，去除不必要的加密。</div>
            <div class="protocol-status">
              <span :class="['status-tag', protocols.vless ? 'enabled' : 'disabled']">
                {{ protocols.vless ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 Shadowsocks 协议</div>
            <div class="protocol-switch">
              <el-switch v-model="protocols.shadowsocks" />
            </div>
            <div class="protocol-desc">Shadowsocks 协议: 经典的加密代理协议。</div>
            <div class="protocol-status">
              <span :class="['status-tag', protocols.shadowsocks ? 'enabled' : 'disabled']">
                {{ protocols.shadowsocks ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 SOCKS 协议</div>
            <div class="protocol-switch">
              <el-switch v-model="protocols.socks" />
            </div>
            <div class="protocol-desc">SOCKS 协议: 标准代理协议，支持 TCP/UDP。</div>
            <div class="protocol-status">
              <span :class="['status-tag', protocols.socks ? 'enabled' : 'disabled']">
                {{ protocols.socks ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 HTTP 协议</div>
            <div class="protocol-switch">
              <el-switch v-model="protocols.http" />
            </div>
            <div class="protocol-desc">HTTP 协议: 基础代理协议，明文传输。</div>
            <div class="protocol-status">
              <span :class="['status-tag', protocols.http ? 'enabled' : 'disabled']">
                {{ protocols.http ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
        </div>
        
        <div class="settings-divider"></div>
        
        <div class="settings-section">
          <div class="section-header">传输层设置</div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 TCP 传输</div>
            <div class="protocol-switch">
              <el-switch v-model="transports.tcp" />
            </div>
            <div class="protocol-desc">TCP 传输: 最基础的传输方式。</div>
            <div class="protocol-status">
              <span :class="['status-tag', transports.tcp ? 'enabled' : 'disabled']">
                {{ transports.tcp ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 WebSocket 传输</div>
            <div class="protocol-switch">
              <el-switch v-model="transports.ws" />
            </div>
            <div class="protocol-desc">WebSocket 传输: 基于HTTP协议的持久化连接，兼容性好。</div>
            <div class="protocol-status">
              <span :class="['status-tag', transports.ws ? 'enabled' : 'disabled']">
                {{ transports.ws ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 HTTP/2 传输</div>
            <div class="protocol-switch">
              <el-switch v-model="transports.http2" />
            </div>
            <div class="protocol-desc">HTTP/2 传输: 新一代HTTP协议，多路复用，需启用TLS。</div>
            <div class="protocol-status">
              <span :class="['status-tag', transports.http2 ? 'enabled' : 'disabled']">
                {{ transports.http2 ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 gRPC 传输</div>
            <div class="protocol-switch">
              <el-switch v-model="transports.grpc" />
            </div>
            <div class="protocol-desc">gRPC 传输: 基于HTTP/2的高性能RPC框架，抗干扰力强。</div>
            <div class="protocol-status">
              <span :class="['status-tag', transports.grpc ? 'enabled' : 'disabled']">
                {{ transports.grpc ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
          
          <div class="protocol-row">
            <div class="protocol-label">启用 QUIC 传输</div>
            <div class="protocol-switch">
              <el-switch v-model="transports.quic" />
            </div>
            <div class="protocol-desc">QUIC 传输: 基于UDP的传输层协议，低延迟。</div>
            <div class="protocol-status">
              <span :class="['status-tag', transports.quic ? 'enabled' : 'disabled']">
                {{ transports.quic ? '已启用' : '已禁用' }}
              </span>
            </div>
          </div>
        </div>
        
        <div class="settings-footer">
          <el-button type="primary" @click="saveSettings">保存协议配置</el-button>
          <el-button type="success" @click="saveAndRestart">保存并重启Xray</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// 协议设置
const protocols = reactive({
  trojan: true,
  vmess: true,
  vless: true,
  shadowsocks: true,
  socks: false,
  http: false
})

// 传输层设置
const transports = reactive({
  tcp: true,
  ws: true,
  http2: true,
  grpc: true,
  quic: false
})

// 保存设置
const saveSettings = async () => {
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    ElMessage.success('协议配置已保存')
  } catch (error) {
    console.error('保存协议设置失败:', error)
    ElMessage.error('保存协议设置失败')
  }
}

// 保存并重启
const saveAndRestart = async () => {
  try {
    ElMessageBox.confirm(
      '确定要保存配置并重启Xray吗？重启过程中，所有连接将会断开。',
      '重启确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
      .then(async () => {
        // 模拟保存和重启过程
        await new Promise(resolve => setTimeout(resolve, 1000))
        ElMessage.success('协议配置已保存，Xray已重启')
      })
      .catch(() => {
        ElMessage.info('已取消重启')
      })
  } catch (error) {
    console.error('重启失败:', error)
    ElMessage.error('重启失败')
  }
}
</script>

<style scoped>
.protocol-settings-container {
  padding: 20px;
  background-color: #f0f2f5;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 20px;
}

.title {
  font-size: 24px;
  font-weight: 500;
  color: #333;
}

.settings-tabs {
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.tab-header {
  display: flex;
  border-bottom: 1px solid #e8e8e8;
  background-color: #fafafa;
}

.tab-item {
  padding: 12px 16px;
  font-size: 14px;
  cursor: pointer;
  color: #666;
  transition: all 0.3s;
}

.tab-item:hover {
  color: #1890ff;
}

.tab-item.active {
  color: #1890ff;
  border-bottom: 2px solid #1890ff;
  background-color: #fff;
}

.tab-content {
  padding: 20px;
}

.settings-section {
  margin-bottom: 20px;
}

.section-header {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 20px;
  color: #333;
}

.protocol-row {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.protocol-label {
  width: 200px;
  color: #333;
}

.protocol-switch {
  width: 50px;
}

.protocol-desc {
  flex: 1;
  color: #666;
  padding: 0 20px;
}

.protocol-status {
  width: 80px;
  text-align: center;
}

.status-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 2px;
  font-size: 12px;
  color: #fff;
}

.status-tag.enabled {
  background-color: #67c23a;
}

.status-tag.disabled {
  background-color: #f56c6c;
}

.settings-divider {
  height: 1px;
  background-color: #e8e8e8;
  margin: 30px 0;
}

.settings-footer {
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #e8e8e8;
  display: flex;
  justify-content: center;
  gap: 20px;
}
</style> 