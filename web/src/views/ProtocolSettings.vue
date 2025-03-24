<template>
  <div class="protocol-settings-container">
    <div class="page-header">
      <div class="title">协议管理</div>
    </div>
    
    <el-card class="settings-card">
      <el-tabs type="card">
        <el-tab-pane label="服务器配置">服务器配置内容</el-tab-pane>
        <el-tab-pane label="数据库配置">数据库配置内容</el-tab-pane>
        <el-tab-pane label="日志配置">日志配置内容</el-tab-pane>
        <el-tab-pane label="Xray内核配置">Xray配置内容</el-tab-pane>
        <el-tab-pane label="管理员配置">管理员配置内容</el-tab-pane>
        <el-tab-pane label="安全设置">安全设置内容</el-tab-pane>
        <el-tab-pane label="协议管理">
          <div class="settings-section">
            <div class="section-title">支持的协议</div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 Trojan 协议</div>
              <div class="protocol-switch">
                <el-switch v-model="enableTrojan" />
              </div>
              <div class="protocol-desc">
                Trojan 协议: 基于 TLS 的轻量级协议，伪装成 HTTPS 流量。
                <el-tag v-if="enableTrojan" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 VMess 协议</div>
              <div class="protocol-switch">
                <el-switch v-model="enableVMess" />
              </div>
              <div class="protocol-desc">
                VMess 协议: V2Ray 的核心传输协议，支持多种传输层。
                <el-tag v-if="enableVMess" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 VLESS 协议</div>
              <div class="protocol-switch">
                <el-switch v-model="enableVLESS" />
              </div>
              <div class="protocol-desc">
                VLESS 协议: 轻量化的 VMess 协议，去除不必要的加密。
                <el-tag v-if="enableVLESS" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 Shadowsocks 协议</div>
              <div class="protocol-switch">
                <el-switch v-model="enableShadowsocks" />
              </div>
              <div class="protocol-desc">
                Shadowsocks 协议: 经典的加密代理协议。
                <el-tag v-if="enableShadowsocks" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 SOCKS 协议</div>
              <div class="protocol-switch">
                <el-switch v-model="enableSOCKS" />
              </div>
              <div class="protocol-desc">
                SOCKS 协议: 标准代理协议，支持 TCP/UDP。
                <el-tag v-if="enableSOCKS" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 HTTP 协议</div>
              <div class="protocol-switch">
                <el-switch v-model="enableHTTP" />
              </div>
              <div class="protocol-desc">
                HTTP 协议: 基础代理协议，明文传输。
                <el-tag v-if="enableHTTP" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
          </div>
          
          <div class="settings-section">
            <div class="section-title">传输层设置</div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 TCP 传输</div>
              <div class="protocol-switch">
                <el-switch v-model="enableTCP" />
              </div>
              <div class="protocol-desc">
                TCP 传输: 最基础的传输方式。
                <el-tag v-if="enableTCP" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 WebSocket 传输</div>
              <div class="protocol-switch">
                <el-switch v-model="enableWebSocket" />
              </div>
              <div class="protocol-desc">
                WebSocket 传输: 基于HTTP协议的持久化连接，兼容性好。
                <el-tag v-if="enableWebSocket" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 HTTP/2 传输</div>
              <div class="protocol-switch">
                <el-switch v-model="enableHTTP2" />
              </div>
              <div class="protocol-desc">
                HTTP/2 传输: 新一代HTTP协议，多路复用，需启用TLS。
                <el-tag v-if="enableHTTP2" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 gRPC 传输</div>
              <div class="protocol-switch">
                <el-switch v-model="enableGRPC" />
              </div>
              <div class="protocol-desc">
                gRPC 传输: 基于HTTP/2的高性能RPC框架，抗干扰力强。
                <el-tag v-if="enableGRPC" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
            
            <div class="protocol-item">
              <div class="protocol-label">启用 QUIC 传输</div>
              <div class="protocol-switch">
                <el-switch v-model="enableQUIC" />
              </div>
              <div class="protocol-desc">
                QUIC 传输: 基于UDP的传输层协议，低延迟。
                <el-tag v-if="enableQUIC" type="success" size="small" style="margin-left: 10px">已启用</el-tag>
                <el-tag v-else type="danger" size="small" style="margin-left: 10px">已禁用</el-tag>
              </div>
            </div>
          </div>
          
          <div class="action-buttons">
            <el-button type="primary" @click="saveSettings">保存协议配置</el-button>
            <el-button type="success" @click="saveAndRestart">保存并重启Xray</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// 协议设置
const enableTrojan = ref(true)
const enableVMess = ref(true)
const enableVLESS = ref(true)
const enableShadowsocks = ref(true)
const enableSOCKS = ref(false)
const enableHTTP = ref(false)

// 传输层设置
const enableTCP = ref(true)
const enableWebSocket = ref(true)
const enableHTTP2 = ref(true)
const enableGRPC = ref(true)
const enableQUIC = ref(false)

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
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title {
  font-size: 20px;
  font-weight: 500;
  color: #333;
}

.settings-card {
  margin-bottom: 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

.settings-section {
  margin-bottom: 30px;
}

.section-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.protocol-item {
  display: flex;
  align-items: flex-start;
  margin-bottom: 15px;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.protocol-label {
  width: 180px;
  font-weight: 500;
  color: #333;
}

.protocol-switch {
  width: 60px;
}

.protocol-desc {
  flex: 1;
  color: #606266;
  font-size: 14px;
  line-height: 1.5;
}

.action-buttons {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: center;
  gap: 20px;
}
</style> 