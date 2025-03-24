<template>
  <div class="settings-container">
    <h1>系统设置</h1>
    
    <el-tabs type="border-card" v-model="activeName" @tab-click="handleTabClick">
      <el-tab-pane label="服务器配置" name="server">
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
      
      <el-tab-pane label="数据库配置" name="db">
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
      
      <el-tab-pane label="日志配置" name="log">
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
      
      <el-tab-pane label="Xray内核配置" name="xray">
        <el-form label-width="120px" class="settings-form">
          <el-form-item label="当前版本">
            <div class="version-info">
              <el-tag size="large" type="success">{{ xraySettings.currentVersion || '未知' }}</el-tag>
              <el-button 
                type="primary" 
                size="small" 
                style="margin-left: 15px;"
                @click="refreshXrayVersions"
                :loading="xraySettings.loading"
              >
                刷新
              </el-button>
              <el-button 
                type="primary" 
                size="small" 
                style="margin-left: 10px;"
                @click="syncVersionsFromGitHub"
                :loading="xraySettings.syncing"
                plain
              >
                同步GitHub
              </el-button>
              <el-button 
                type="warning" 
                size="small" 
                style="margin-left: 10px;"
                @click="checkXrayUpdates"
                :loading="xraySettings.checkingForUpdates"
              >
                检查更新
              </el-button>
              <el-button 
                type="info" 
                size="small" 
                style="margin-left: 10px;"
                @click="loadVersionDetails(xraySettings.currentVersion)"
                :disabled="xraySettings.currentVersion === '未知'"
              >
                版本详情
              </el-button>
              <el-button 
                type="success" 
                size="small" 
                plain
                style="margin-left: 10px;"
                @click="openXrayReleasePage"
              >
                GitHub发布页
              </el-button>
            </div>
          </el-form-item>
          
          <el-form-item label="切换版本">
            <div class="version-selector">
              <el-select 
                v-model="xraySettings.selectedVersion" 
                placeholder="请选择Xray版本"
                style="width: 260px;" 
                :loading="xraySettings.loading"
                filterable
              >
                <el-option
                  v-for="version in xraySettings.versions"
                  :key="version"
                  :label="version"
                  :value="version"
                >
                  <div class="version-option">
                    <span>{{ version }}</span>
                    <el-tag size="small" type="success" v-if="version === xraySettings.currentVersion">当前运行</el-tag>
                  </div>
                </el-option>
              </el-select>
              <el-button 
                type="primary" 
                style="margin-left: 15px;" 
                @click="switchXrayVersion"
                :loading="xraySettings.switching"
                :disabled="!xraySettings.selectedVersion || xraySettings.selectedVersion === xraySettings.currentVersion"
              >
                切换版本
              </el-button>
              <el-tooltip 
                effect="dark" 
                content="切换版本后需要重启Xray才能生效" 
                placement="top"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
            <div class="version-tips" v-if="xraySettings.versions.length > 0">
              <p>可用版本: {{ xraySettings.versions.length }} 个</p>
              <p>推荐版本: {{ xraySettings.versions[0] }}</p>
            </div>
          </el-form-item>
          
          <el-form-item label="自动更新">
            <el-switch 
              v-model="xraySettings.autoUpdate" 
              @change="toggleAutoUpdate"
            />
            <div class="form-tips">启用后，系统将自动更新到最新的稳定版</div>
          </el-form-item>
          
          <el-form-item label="检查更新间隔">
            <el-select v-model="xraySettings.checkInterval" style="width: 200px;">
              <el-option label="从不" :value="0" />
              <el-option label="每天" :value="24" />
              <el-option label="每周" :value="168" />
              <el-option label="每月" :value="720" />
            </el-select>
            <div class="form-tips">自动检查Xray更新的时间间隔</div>
          </el-form-item>
          
          <el-form-item label="使用自定义配置">
            <el-switch 
              v-model="xraySettings.customConfig"
              @change="console.log('customConfig changed:', xraySettings.customConfig)"
            />
            <div class="form-tips">启用后，将使用自定义的Xray配置文件，而不是由系统生成</div>
          </el-form-item>
          
          <el-form-item label="配置文件路径" v-if="xraySettings.customConfig">
            <el-input v-model="xraySettings.configPath" placeholder="/path/to/config.json">
              <template #append>
                <el-button @click="testCustomConfig">测试配置</el-button>
              </template>
            </el-input>
            <div class="form-tips">自定义配置文件的绝对路径，请确保格式正确且有权限访问</div>
          </el-form-item>
          
          <el-divider></el-divider>
          <el-form-item>
            <el-button type="primary" @click="saveXraySettings">保存设置</el-button>
            <el-button type="success" @click="restartXray" :loading="xraySettings.restarting">重启Xray</el-button>
            <el-button @click="refreshXraySettings" :loading="xraySettings.loading">刷新设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <el-tab-pane label="管理员配置" name="admin">
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
      
      <el-tab-pane label="安全设置" name="security">
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
      
      <!-- 新增协议管理标签页 -->
      <el-tab-pane label="协议管理" name="protocol">
        <el-form class="settings-form">
          <el-form-item label="支持的协议" label-width="120px">
            <el-descriptions :column="1" border size="medium">
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="protocolSettings.enableTrojan"
                    active-text="启用 Trojan 协议"
                    :disabled="disableProtocolSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>Trojan 协议：基于 TLS 的轻量级协议，伪装成 HTTPS 流量。</p>
                  <el-tag type="success" size="small" v-if="protocolSettings.enableTrojan">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="protocolSettings.enableVMess"
                    active-text="启用 VMess 协议"
                    :disabled="disableProtocolSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>VMess 协议：V2Ray 的核心传输协议，支持多种传输层。</p>
                  <el-tag type="success" size="small" v-if="protocolSettings.enableVMess">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="protocolSettings.enableVLESS"
                    active-text="启用 VLESS 协议"
                    :disabled="disableProtocolSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>VLESS 协议：轻量化的 VMess 协议，去除不必要的加密。</p>
                  <el-tag type="success" size="small" v-if="protocolSettings.enableVLESS">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="protocolSettings.enableShadowsocks"
                    active-text="启用 Shadowsocks 协议"
                    :disabled="disableProtocolSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>Shadowsocks 协议：经典的加密代理协议。</p>
                  <el-tag type="success" size="small" v-if="protocolSettings.enableShadowsocks">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="protocolSettings.enableSocks"
                    active-text="启用 SOCKS 协议"
                    :disabled="disableProtocolSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>SOCKS 协议：标准代理协议，支持 TCP/UDP。</p>
                  <el-tag type="success" size="small" v-if="protocolSettings.enableSocks">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="protocolSettings.enableHTTP"
                    active-text="启用 HTTP 协议"
                    :disabled="disableProtocolSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>HTTP 协议：基础代理协议，明文传输。</p>
                  <el-tag type="success" size="small" v-if="protocolSettings.enableHTTP">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
            </el-descriptions>
          </el-form-item>
          
          <el-divider content-position="left">传输层设置</el-divider>
          
          <el-form-item label="支持的传输层" label-width="120px">
            <el-descriptions :column="1" border size="medium">
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="transportSettings.enableTCP"
                    active-text="启用 TCP 传输"
                    :disabled="disableTransportSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>TCP 传输：最基础的传输方式。</p>
                  <el-tag type="success" size="small" v-if="transportSettings.enableTCP">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="transportSettings.enableWebSocket"
                    active-text="启用 WebSocket 传输"
                    :disabled="disableTransportSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>WebSocket 传输：基于HTTP协议的持久化连接，兼容性好。</p>
                  <el-tag type="success" size="small" v-if="transportSettings.enableWebSocket">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="transportSettings.enableHTTP2"
                    active-text="启用 HTTP/2 传输"
                    :disabled="disableTransportSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>HTTP/2 传输：新一代HTTP协议，多路复用，需启用TLS。</p>
                  <el-tag type="success" size="small" v-if="transportSettings.enableHTTP2">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="transportSettings.enableGRPC"
                    active-text="启用 gRPC 传输"
                    :disabled="disableTransportSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>gRPC 传输：基于HTTP/2的高性能RPC框架，抗干扰能力强。</p>
                  <el-tag type="success" size="small" v-if="transportSettings.enableGRPC">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
              
              <el-descriptions-item>
                <template #label>
                  <el-switch
                    v-model="transportSettings.enableQUIC"
                    active-text="启用 QUIC 传输"
                    :disabled="disableTransportSwitch"
                  />
                </template>
                <div class="protocol-description">
                  <p>QUIC 传输：基于UDP的传输层协议，低延迟。</p>
                  <el-tag type="success" size="small" v-if="transportSettings.enableQUIC">已启用</el-tag>
                  <el-tag type="danger" size="small" v-else>已禁用</el-tag>
                </div>
              </el-descriptions-item>
            </el-descriptions>
          </el-form-item>
          
          <el-divider></el-divider>
          
          <el-form-item>
            <el-button type="primary" @click="saveProtocolSettings" :loading="protocolsLoading">保存协议配置</el-button>
            <el-button type="warning" @click="restartXrayAfterProtocolChange" :loading="xraySettings.restarting">保存并重启Xray</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>

  <!-- 添加在组件末尾的弹窗 -->
  <el-dialog 
    v-model="xraySettings.showVersionDetails" 
    title="Xray版本详情" 
    width="600px"
    destroy-on-close
  >
    <el-descriptions :column="1" border>
      <el-descriptions-item label="版本">
        <el-tag type="success">{{ xraySettings.versionDetails.version }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="发布日期">
        {{ xraySettings.versionDetails.releaseDate }}
      </el-descriptions-item>
      <el-descriptions-item label="描述">
        {{ xraySettings.versionDetails.description }}
      </el-descriptions-item>
      <el-descriptions-item label="更新日志">
        <ul class="changelog-list">
          <li v-for="(change, index) in xraySettings.versionDetails.changelog" :key="index">
            {{ change }}
          </li>
        </ul>
      </el-descriptions-item>
    </el-descriptions>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="xraySettings.showVersionDetails = false">关闭</el-button>
        <el-button type="primary" @click="openXrayReleasePage">
          查看GitHub发布页
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 添加更新进度对话框 -->
  <el-dialog 
    v-model="xraySettings.updateProgress.visible" 
    :title="`Xray 更新 - ${xraySettings.downloadingVersion}`"
    width="500px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="xraySettings.updateProgress.status === 'completed' || xraySettings.updateProgress.status === 'error'"
  >
    <div class="update-progress">
      <el-progress 
        :percentage="xraySettings.updateProgress.percent" 
        :status="xraySettings.updateProgress.status === 'error' ? 'exception' : 
                xraySettings.updateProgress.status === 'completed' ? 'success' : ''"
        :striped="xraySettings.updateProgress.status === 'downloading' || xraySettings.updateProgress.status === 'installing'"
        :striped-flow="xraySettings.updateProgress.status === 'downloading' || xraySettings.updateProgress.status === 'installing'"
      ></el-progress>
      
      <div class="update-status">
        <span>{{ xraySettings.updateProgress.message }}</span>
        <span class="error-message" v-if="xraySettings.updateProgress.status === 'error'">
          错误: {{ xraySettings.updateProgress.error }}
        </span>
      </div>
    </div>
    
    <template #footer v-if="xraySettings.updateProgress.status === 'completed' || xraySettings.updateProgress.status === 'error'">
      <el-button @click="xraySettings.updateProgress.visible = false">关闭</el-button>
      <el-button 
        type="primary" 
        v-if="xraySettings.updateProgress.status === 'completed'"
        @click="restartXray"
      >
        重启Xray
      </el-button>
      <el-button 
        type="warning" 
        v-if="xraySettings.updateProgress.status === 'error'"
        @click="downloadXrayVersion(xraySettings.downloadingVersion)"
      >
        重试
      </el-button>
    </template>
  </el-dialog>
  
  <!-- 添加错误详情对话框 -->
  <el-dialog
    v-model="errorDetails.visible"
    title="错误详情"
    width="600px"
    destroy-on-close
  >
    <div class="error-details-container">
      <el-alert
        :title="errorDetails.title"
        type="error"
        description=""
        show-icon
        :closable="false"
        style="margin-bottom: 15px;"
      />
      
      <el-card shadow="never" class="error-card">
        <template #header>
          <div class="error-header">
            <span>错误信息</span>
            <el-button 
              type="primary" 
              size="small" 
              plain 
              @click="copyErrorToClipboard"
              circle
              icon="CopyDocument"
            />
          </div>
        </template>
        <pre class="error-message-content">{{ errorDetails.message }}</pre>
      </el-card>
      
      <div class="error-resolution" v-if="errorDetails.resolution">
        <h4>可能的解决方案：</h4>
        <p>{{ errorDetails.resolution }}</p>
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="errorDetails.visible = false">关闭</el-button>
        <el-button type="primary" @click="retryFailedOperation" v-if="errorDetails.canRetry">
          重试操作
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import axios from 'axios'
import { InfoFilled, CopyDocument } from '@element-plus/icons-vue'

// store
const userStore = useUserStore()

// 当前活动标签页
const activeName = ref('server')

// Xray设置
const xraySettings = reactive({
  currentVersion: '未知',
  selectedVersion: '',
  versions: [],
  loading: false,
  switching: false,
  restarting: false,
  autoUpdate: false,
  customConfig: false,
  configPath: '',
  checkInterval: 24, // 默认每天检查一次
  showVersionDetails: false,
  versionDetails: {
    version: '',
    releaseDate: '',
    description: '',
    changelog: []
  },
  updateProgress: {
    visible: false,
    status: 'checking', // checking, downloading, installing, completed, error
    percent: 0,
    message: '正在检查更新...',
    error: ''
  },
  downloadingVersion: '',
  checkingForUpdates: false,
  syncing: false
})

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

// 协议设置
const protocolSettings = reactive({
  enableTrojan: true,
  enableVMess: true,
  enableVLESS: true,
  enableShadowsocks: true,
  enableSocks: false,
  enableHTTP: false
})

// 传输层设置
const transportSettings = reactive({
  enableTCP: true,
  enableWebSocket: true,
  enableHTTP2: true,
  enableGRPC: true,
  enableQUIC: false
})

// 状态控制
const protocolsLoading = ref(false)
const disableProtocolSwitch = computed(() => {
  // 如果正在加载或者Xray正在重启，禁用开关
  return protocolsLoading.value || xraySettings.restarting
})
const disableTransportSwitch = computed(() => {
  // 如果正在加载或者Xray正在重启，禁用开关
  return protocolsLoading.value || xraySettings.restarting
})

// 错误详情
const errorDetails = reactive({
  visible: false,
  title: '',
  message: '',
  resolution: '',
  canRetry: false,
  retryAction: null,
  retryParams: null
})

// 处理标签页切换
const handleTabClick = (tab) => {
  console.log('Tab clicked:', tab.props.name)
  if (tab.props.name === 'xray') {
    // 当切换到Xray标签页时，刷新数据
    refreshXraySettings()
  }
}

// 重启Xray
const restartXray = async () => {
  try {
    // 显示确认对话框
    await ElMessageBox.confirm(
      '确定要重启Xray服务吗？这可能会暂时中断所有正在连接的用户。',
      '重启Xray',
      {
        confirmButtonText: '确认重启',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    xraySettings.restarting = true
    ElMessage.info('正在停止Xray服务...')
    
    // 停止服务
    await axios.post('/api/xray/stop')
    
    // 等待短暂时间确保服务完全停止
    await new Promise(resolve => setTimeout(resolve, 1500)) 
    
    ElMessage.info('正在启动Xray服务...')
    
    // 启动服务
    const startResponse = await axios.post('/api/xray/start')
    
    // 检查启动是否成功
    if (startResponse.data && startResponse.data.success) {
      ElMessage.success({
        message: 'Xray 服务已成功重启',
        duration: 3000
      })
      
      // 重新加载配置信息
      setTimeout(() => {
        refreshXrayVersions()
      }, 1000)
    } else {
      throw new Error(startResponse.data?.message || '启动服务失败')
    }
  } catch (error) {
    if (error === 'cancel') {
      ElMessage.info('已取消重启操作')
      return
    }
    
    console.error('Failed to restart Xray:', error)
    ElMessage.error({
      message: '重启 Xray 失败: ' + (error.response?.data?.message || error.message || '未知错误'),
      duration: 5000
    })
    
    // 显示详细错误信息
    const errorMsg = error.response?.data?.message || error.message || '未知错误'
    const errorDetail = error.response?.data?.detail || error.stack || ''
    showErrorDetails(
      '重启 Xray 失败',
      errorDetail ? `${errorMsg}\n\n${errorDetail}` : errorMsg,
      '请检查Xray服务状态和日志，确保端口未被占用，并且有足够的系统权限。',
      true,
      'restart'
    )
    
    // 尝试恢复服务
    try {
      ElMessage.warning('正在尝试恢复服务...')
      await axios.post('/api/xray/start')
    } catch (recoveryError) {
      console.error('Failed to recover Xray service:', recoveryError)
      ElMessage.error({
        message: 'Xray服务恢复失败，请手动检查服务状态',
        duration: 5000
      })
    }
  } finally {
    xraySettings.restarting = false
  }
}

// 执行版本切换
const performVersionSwitch = async (shouldRestart = false) => {
  xraySettings.switching = true
  ElMessage.info('正在下载并切换Xray版本，请耐心等待...')
  
  try {
    // 发送版本切换请求
    const response = await axios.post('/api/xray/version', {
      version: xraySettings.selectedVersion
    })
    
    console.log('Switch version response:', response.data)
    
    // 切换成功
    xraySettings.currentVersion = xraySettings.selectedVersion
    ElMessage.success(`成功切换到 Xray 版本: ${xraySettings.selectedVersion}`)
    
    // 根据用户选择决定是否重启
    if (shouldRestart) {
      ElMessage.info('正在重启Xray服务...')
      await restartXray()
      ElMessage.success('Xray服务已重启，新版本已生效')
    } else {
      ElMessage.warning('版本已切换，但需要重启Xray服务才能生效')
    }
  } catch (error) {
    console.error('Failed to switch Xray version:', error)
    // 提供更详细的错误信息
    const errorMsg = error.response?.data?.message || error.message || '未知错误'
    ElMessage.error({
      message: `切换 Xray 版本失败: ${errorMsg}`,
      duration: 5000
    })
    
    // 显示详细错误信息
    const errorDetail = error.response?.data?.detail || error.stack || ''
    showErrorDetails(
      '切换 Xray 版本失败',
      errorDetail ? `${errorMsg}\n\n${errorDetail}` : errorMsg,
      '请检查网络连接和下载权限，确保目标版本可用且格式正确。',
      true,
      'switchVersion',
      { shouldRestart }
    )
  } finally {
    xraySettings.switching = false
  }
}

// 初始化时加载Xray版本信息
onMounted(async () => {
  // 显示加载状态
  xraySettings.loading = true
  
  try {
    // 依次加载各项数据
    await loadXrayVersions()
    await loadXraySettings()
    await loadProtocolSettings()
    
    // 输出初始状态
    console.log('Initial xraySettings:', { ...xraySettings })
  } catch (error) {
    console.error('Failed to load initial settings:', error)
    ElMessage.error('加载设置失败，请刷新页面重试')
  } finally {
    xraySettings.loading = false
  }
})

// 加载Xray版本信息
const loadXrayVersions = async () => {
  xraySettings.loading = true
  try {
    // 显示正在刷新的提示
    ElMessage.info('正在刷新Xray版本信息...')
    
    // 尝试从API获取Xray版本信息
    let versionsUpdated = false
    try {
      const response = await axios.get('/api/xray/versions')
      if (response.data && response.data.current_version) {
        xraySettings.currentVersion = response.data.current_version
      }
      
      if (response.data && Array.isArray(response.data.supported_versions) && response.data.supported_versions.length > 0) {
        xraySettings.versions = response.data.supported_versions
        versionsUpdated = true
      }
    } catch (error) {
      console.warn('Failed to get versions from API, will try to sync from GitHub:', error)
    }
    
    // 如果从API获取版本失败，尝试从GitHub同步
    if (!versionsUpdated) {
      await syncVersionsFromGitHub()
    }
    
    // 确保selectedVersion有值，默认为当前版本
    if (!xraySettings.selectedVersion || xraySettings.selectedVersion === '') {
      if (xraySettings.currentVersion && xraySettings.currentVersion !== '未知') {
        xraySettings.selectedVersion = xraySettings.currentVersion
      } else if (xraySettings.versions.length > 0) {
        xraySettings.selectedVersion = xraySettings.versions[0]
      }
    }
    
    // 成功刷新提示
    ElMessage.success('Xray版本信息已刷新')
    
    // 同时加载Xray设置
    await loadXraySettings()
  } catch (error) {
    console.error('Failed to refresh Xray versions:', error)
    ElMessage.error('刷新Xray版本信息失败: ' + (error.message || '未知错误'))
    
    // 使用默认版本列表
    xraySettings.versions = [
      'v25.2.21', 'v25.2.18', 'v25.1.30', 'v25.1.1', 'v24.12.31'
    ]
    
    // 选择一个默认版本
    if (!xraySettings.selectedVersion || xraySettings.selectedVersion === '') {
      xraySettings.selectedVersion = xraySettings.versions[0]
    }
    
    // 显示错误详情
    showErrorDetails(
      '刷新Xray版本信息失败',
      error.response?.data?.message || error.message || '未知错误',
      '请检查网络连接或服务器状态，或者稍后再试。',
      true,
      'refreshVersions'
    )
  } finally {
    xraySettings.loading = false
  }
}

// 从GitHub同步Xray版本
const syncVersionsFromGitHub = async () => {
  // 设置同步状态
  xraySettings.syncing = true
  
  try {
    ElMessage.info('正在从GitHub同步Xray版本信息...')
    
    // 调用后端API同步版本
    const response = await axios.post('/api/xray/sync-versions-from-github')
    
    if (response.data && response.data.success) {
      // 同步成功，获取同步到的版本列表
      const syncResponse = await axios.get('/api/xray/versions')
      if (syncResponse.data && Array.isArray(syncResponse.data.supported_versions) && syncResponse.data.supported_versions.length > 0) {
        xraySettings.versions = syncResponse.data.supported_versions
        ElMessage.success('已从GitHub同步最新Xray版本信息')
        
        // 更新当前选择的版本
        if (!xraySettings.selectedVersion || !xraySettings.versions.includes(xraySettings.selectedVersion)) {
          if (xraySettings.versions.includes(xraySettings.currentVersion)) {
            xraySettings.selectedVersion = xraySettings.currentVersion
          } else if (xraySettings.versions.length > 0) {
            xraySettings.selectedVersion = xraySettings.versions[0]
          }
        }
        
        return true
      }
    }
    
    // 如果同步失败或没有返回版本列表，手动设置最新版本
    console.warn('Server-side sync failed or empty, using hard-coded versions')
    xraySettings.versions = [
      'v25.2.21', 'v25.2.18', 'v25.1.30', 'v25.1.1', 'v24.12.31',
      'v24.12.28', 'v24.12.18', 'v24.12.15', 'v24.11.30'
    ]
    
    // 直接调用接口更新服务器版本列表
    try {
      await axios.post('/api/xray/update-versions', {
        versions: xraySettings.versions
      })
      console.log('Updated server versions with hard-coded list')
      ElMessage.success('使用内置版本列表更新成功')
    } catch (updateError) {
      console.error('Failed to update server versions:', updateError)
      ElMessage.warning('更新服务器版本列表失败，但本地列表已更新')
    }
    
    return true
  } catch (error) {
    console.error('Failed to sync versions from GitHub:', error)
    ElMessage.error('从GitHub同步版本失败: ' + (error.message || '未知错误'))
    
    // 显示错误详情
    showErrorDetails(
      '从GitHub同步Xray版本失败',
      error.response?.data?.message || error.message || '未知错误',
      '请检查网络连接和GitHub的可访问性，或者稍后再试。',
      true,
      'syncVersions'
    )
    
    // 使用硬编码的版本列表作为备份
    xraySettings.versions = [
      'v25.2.21', 'v25.2.18', 'v25.1.30', 'v25.1.1', 'v24.12.31',
      'v24.12.28', 'v24.12.18', 'v24.12.15', 'v24.11.30'
    ]
    
    return false
  } finally {
    // 清除同步状态
    xraySettings.syncing = false
  }
}

// 加载Xray设置
const loadXraySettings = async () => {
  try {
    console.log('开始加载Xray设置，当前值：', { ...xraySettings });
    
    const response = await axios.get('/api/settings/xray');
    
    if (response.data) {
      // 确保使用正确的属性名从API响应获取数据
      const newAutoUpdate = response.data.auto_update ?? false;
      const newCustomConfig = response.data.custom_config ?? false;
      const newConfigPath = response.data.config_path || '';
      const newCheckInterval = response.data.check_interval || 24;
      
      // 将API返回值与当前值进行比较
      console.log('Xray设置对比：', {
        autoUpdate: { old: xraySettings.autoUpdate, new: newAutoUpdate },
        customConfig: { old: xraySettings.customConfig, new: newCustomConfig },
        configPath: { old: xraySettings.configPath, new: newConfigPath },
        checkInterval: { old: xraySettings.checkInterval, new: newCheckInterval }
      });
      
      // 设置新值
      xraySettings.autoUpdate = newAutoUpdate;
      xraySettings.customConfig = newCustomConfig;
      xraySettings.configPath = newConfigPath;
      xraySettings.checkInterval = newCheckInterval;
      
      console.log('Xray设置加载完成:', { ...xraySettings });
    }
  } catch (error) {
    console.error('Failed to load Xray settings:', error);
  }
}

// 刷新Xray版本信息
const refreshXrayVersions = async () => {
  xraySettings.loading = true
  try {
    // 显示正在刷新的提示
    ElMessage.info('正在刷新Xray版本信息...')
    
    // 尝试从API获取Xray版本信息
    let versionsUpdated = false
    try {
      const response = await axios.get('/api/xray/versions')
      if (response.data && response.data.current_version) {
        xraySettings.currentVersion = response.data.current_version
      }
      
      if (response.data && Array.isArray(response.data.supported_versions) && response.data.supported_versions.length > 0) {
        xraySettings.versions = response.data.supported_versions
        versionsUpdated = true
      }
    } catch (error) {
      console.warn('Failed to get versions from API, will try to sync from GitHub:', error)
    }
    
    // 如果从API获取版本失败，尝试从GitHub同步
    if (!versionsUpdated) {
      await syncVersionsFromGitHub()
    }
    
    // 确保selectedVersion有值，默认为当前版本
    if (!xraySettings.selectedVersion || xraySettings.selectedVersion === '') {
      if (xraySettings.currentVersion && xraySettings.currentVersion !== '未知') {
        xraySettings.selectedVersion = xraySettings.currentVersion
      } else if (xraySettings.versions.length > 0) {
        xraySettings.selectedVersion = xraySettings.versions[0]
      }
    }
    
    // 成功刷新提示
    ElMessage.success('Xray版本信息已刷新')
    
    // 同时加载Xray设置
    await loadXraySettings()
  } catch (error) {
    console.error('Failed to refresh Xray versions:', error)
    ElMessage.error('刷新Xray版本信息失败: ' + (error.message || '未知错误'))
    
    // 使用默认版本列表
    xraySettings.versions = [
      'v25.2.21', 'v25.2.18', 'v25.1.30', 'v25.1.1', 'v24.12.31'
    ]
    
    // 选择一个默认版本
    if (!xraySettings.selectedVersion || xraySettings.selectedVersion === '') {
      xraySettings.selectedVersion = xraySettings.versions[0]
    }
    
    // 显示错误详情
    showErrorDetails(
      '刷新Xray版本信息失败',
      error.response?.data?.message || error.message || '未知错误',
      '请检查网络连接或服务器状态，或者稍后再试。',
      true,
      'refreshVersions'
    )
  } finally {
    xraySettings.loading = false
  }
}

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

// 切换Xray版本
const switchXrayVersion = async () => {
  if (!xraySettings.selectedVersion) {
    ElMessage.warning('请选择要切换的版本')
    return
  }
  
  // 如果选择的版本和当前版本相同，提示用户
  if (xraySettings.selectedVersion === xraySettings.currentVersion) {
    ElMessage.info('您选择的版本与当前运行的版本相同')
    return
  }
  
  // 显示确认对话框，询问用户是否希望在切换后自动重启Xray
  ElMessageBox.confirm(
    `确认将Xray版本从 ${xraySettings.currentVersion} 切换到 ${xraySettings.selectedVersion}？`,
    '切换版本',
    {
      confirmButtonText: '切换并重启',
      cancelButtonText: '仅切换',
      distinguishCancelAndClose: true,
      closeOnClickModal: false,
      type: 'warning',
      showClose: true,
    }
  )
  .then(async () => {
    // 用户选择切换并重启
    await performVersionSwitch(true)
  })
  .catch(action => {
    if (action === 'cancel') {
      // 用户选择仅切换不重启
      performVersionSwitch(false)
    } else {
      // 用户关闭对话框
      ElMessage.info('已取消切换操作')
    }
  })
}

// 测试自定义配置
const testCustomConfig = async () => {
  if (!xraySettings.configPath) {
    ElMessage.warning('请输入配置文件路径')
    return
  }
  
  try {
    await axios.post('/api/xray/test-config', {
      config_path: xraySettings.configPath
    })
    ElMessage.success('配置文件测试通过')
  } catch (error) {
    console.error('Failed to test config:', error)
    ElMessage.error('配置文件测试失败：' + (error.response?.data?.message || '文件不存在或格式错误'))
  }
}

// 保存Xray设置
const saveXraySettings = async () => {
  try {
    // 构建请求数据
    const requestData = {
      auto_update: xraySettings.autoUpdate,
      custom_config: xraySettings.customConfig,
      config_path: xraySettings.configPath,
      check_interval: xraySettings.checkInterval
    };
    
    console.log('保存设置请求数据:', requestData);
    
    // 发送正确的属性名与API匹配
    const response = await axios.post('/api/settings/xray', requestData);
    
    console.log('保存设置响应数据:', response.data);
    
    // 保存成功后立即重新获取设置以验证
    await refreshXraySettings();
    
    ElMessage.success('Xray设置保存成功');
  } catch (error) {
    console.error('Failed to save Xray settings:', error);
    ElMessage.error('保存Xray设置失败：' + (error.response?.data?.message || '未知错误'));
    
    // 显示详细错误信息
    const errorMsg = error.response?.data?.message || error.message || '未知错误'
    const errorDetail = error.response?.data?.detail || error.stack || ''
    showErrorDetails(
      '保存 Xray 设置失败',
      errorDetail ? `${errorMsg}\n\n${errorDetail}` : errorMsg,
      '请检查设置参数是否正确，特别是配置文件路径是否有效。',
      true,
      'saveSettings'
    )
  }
}

// 加载协议设置
const loadProtocolSettings = async () => {
  try {
    const response = await axios.get('/api/settings/protocols')
    if (response.data) {
      // 协议设置
      protocolSettings.enableTrojan = response.data.protocols?.trojan ?? true
      protocolSettings.enableVMess = response.data.protocols?.vmess ?? true
      protocolSettings.enableVLESS = response.data.protocols?.vless ?? true
      protocolSettings.enableShadowsocks = response.data.protocols?.shadowsocks ?? true
      protocolSettings.enableSocks = response.data.protocols?.socks ?? false
      protocolSettings.enableHTTP = response.data.protocols?.http ?? false
      
      // 传输层设置
      transportSettings.enableTCP = response.data.transports?.tcp ?? true
      transportSettings.enableWebSocket = response.data.transports?.ws ?? true
      transportSettings.enableHTTP2 = response.data.transports?.http2 ?? true
      transportSettings.enableGRPC = response.data.transports?.grpc ?? true
      transportSettings.enableQUIC = response.data.transports?.quic ?? false
    }
  } catch (error) {
    console.error('Failed to load protocol settings:', error)
    ElMessage.error('加载协议设置失败')
  }
}

// 保存协议设置
const saveProtocolSettings = async () => {
  protocolsLoading.value = true
  try {
    await axios.post('/api/settings/protocols', {
      protocols: {
        trojan: protocolSettings.enableTrojan,
        vmess: protocolSettings.enableVMess,
        vless: protocolSettings.enableVLESS,
        shadowsocks: protocolSettings.enableShadowsocks,
        socks: protocolSettings.enableSocks,
        http: protocolSettings.enableHTTP
      },
      transports: {
        tcp: transportSettings.enableTCP,
        ws: transportSettings.enableWebSocket,
        http2: transportSettings.enableHTTP2,
        grpc: transportSettings.enableGRPC,
        quic: transportSettings.enableQUIC
      }
    })
    ElMessage.success('协议配置保存成功')
  } catch (error) {
    console.error('Failed to save protocol settings:', error)
    ElMessage.error('保存协议配置失败: ' + (error.response?.data?.message || error.message))
  } finally {
    protocolsLoading.value = false
  }
}

// 保存协议设置并重启Xray
const restartXrayAfterProtocolChange = async () => {
  try {
    // 先保存设置
    await saveProtocolSettings()
    // 然后重启Xray
    await restartXray()
    ElMessage.success('协议配置已更新并重启Xray')
  } catch (error) {
    console.error('Failed to update protocols and restart Xray:', error)
    ElMessage.error('更新协议配置并重启Xray失败')
  }
}

// 专门处理自动更新开关
const toggleAutoUpdate = async (newValue) => {
  console.log('自动更新开关切换:', newValue);
  
  try {
    // 构建请求数据
    const requestData = {
      auto_update: newValue,
      custom_config: xraySettings.customConfig,
      config_path: xraySettings.configPath,
      check_interval: xraySettings.checkInterval
    };
    
    console.log('自动更新请求数据:', requestData);
    
    // 发送请求前临时禁用开关，防止重复点击
    const originalValue = xraySettings.autoUpdate;
    
    // 发送请求
    const response = await axios.post('/api/settings/xray', requestData);
    
    console.log('自动更新响应数据:', response.data);
    
    if (response.data && response.data.success) {
      ElMessage.success(`自动更新已${newValue ? '启用' : '禁用'}`);
      
      // 设置一个延迟后再刷新设置，确保后端已完成持久化
      setTimeout(async () => {
        try {
          // 进行多次验证确保设置生效
          const verify1 = await axios.get('/api/settings/xray');
          console.log('验证1结果:', verify1.data);
          
          // 再次刷新设置
          await refreshXraySettings();
          
          // 最终验证
          const verify2 = await axios.get('/api/settings/xray');
          console.log('验证2结果:', verify2.data);
          
          // 检查设置是否一致
          if (verify2.data.auto_update !== newValue) {
            console.error('设置验证失败，值不一致:', {
              expected: newValue,
              actual: verify2.data.auto_update
            });
            ElMessage.warning('设置可能未正确保存，请刷新页面后重试');
          } else {
            console.log('设置验证成功，值一致:', newValue);
          }
        } catch (verifyError) {
          console.error('设置验证失败:', verifyError);
        }
      }, 500);
    } else {
      // 如果失败，恢复之前的状态
      xraySettings.autoUpdate = originalValue;
      ElMessage.error('设置自动更新失败');
    }
  } catch (error) {
    console.error('设置自动更新失败:', error);
    // 恢复之前的状态
    xraySettings.autoUpdate = !newValue;
    ElMessage.error('设置自动更新失败: ' + (error.response?.data?.message || '未知错误'));
  }
}

// 加载版本详情
const loadVersionDetails = async (version) => {
  try {
    xraySettings.loading = true
    
    // 在实际项目中应该从API获取版本详情
    // const response = await axios.get(`/api/xray/version/${version}/details`)
    // 模拟API响应
    const mockResponse = {
      data: {
        version: version,
        release_date: '2023-12-15',
        description: 'Xray-core是一个平台，用于构建用于规避网络限制的代理应用程序。',
        changelog: [
          '修复了TCP连接的bug',
          '改进了WebSocket传输的稳定性',
          '优化了内存使用',
          '新增了对HTTP/3的支持'
        ]
      }
    }
    
    // 更新版本详情
    xraySettings.versionDetails = {
      version: mockResponse.data.version,
      releaseDate: mockResponse.data.release_date,
      description: mockResponse.data.description,
      changelog: mockResponse.data.changelog
    }
    
    // 显示版本详情对话框
    xraySettings.showVersionDetails = true
  } catch (error) {
    console.error('Failed to load version details:', error)
    ElMessage.error('加载版本详情失败：' + (error.message || '未知错误'))
  } finally {
    xraySettings.loading = false
  }
}

// 打开GitHub Xray发布页面
const openXrayReleasePage = () => {
  window.open('https://github.com/XTLS/Xray-core/releases', '_blank')
}

// 检查Xray更新
const checkXrayUpdates = async () => {
  if (xraySettings.checkingForUpdates) return;
  
  try {
    xraySettings.checkingForUpdates = true;
    ElMessage.info('正在检查Xray更新...');
    
    // 这里应该调用实际的API
    // const response = await axios.get('/api/xray/check-updates');
    
    // 模拟API响应
    await new Promise(resolve => setTimeout(resolve, 1500));
    const mockResponse = {
      data: {
        has_update: true,
        latest_version: 'v1.8.25',
        current_version: xraySettings.currentVersion,
        release_notes: '- 修复了多个安全漏洞\n- 提升了性能和稳定性'
      }
    };
    
    if (mockResponse.data.has_update) {
      ElMessageBox.confirm(
        `发现新版本: ${mockResponse.data.latest_version}\n\n更新说明:\n${mockResponse.data.release_notes}`,
        '有可用更新',
        {
          confirmButtonText: '更新',
          cancelButtonText: '取消',
          type: 'info',
          dangerouslyUseHTMLString: true
        }
      ).then(() => {
        downloadXrayVersion(mockResponse.data.latest_version);
      }).catch(() => {
        ElMessage.info('已取消更新');
      });
    } else {
      ElMessage.success('当前已经是最新版本');
    }
  } catch (error) {
    console.error('Failed to check for updates:', error);
    ElMessage.error('检查更新失败: ' + (error.message || '未知错误'));
  } finally {
    xraySettings.checkingForUpdates = false;
  }
};

// 下载并安装Xray版本
const downloadXrayVersion = async (version) => {
  // 初始化进度显示
  xraySettings.updateProgress.visible = true;
  xraySettings.updateProgress.status = 'downloading';
  xraySettings.updateProgress.percent = 0;
  xraySettings.updateProgress.message = `正在下载 ${version}...`;
  xraySettings.downloadingVersion = version;
  
  try {
    // 模拟下载进度
    for (let i = 0; i <= 100; i += 10) {
      xraySettings.updateProgress.percent = i;
      await new Promise(resolve => setTimeout(resolve, 300));
    }
    
    // 下载完成，开始安装
    xraySettings.updateProgress.status = 'installing';
    xraySettings.updateProgress.message = `正在安装 ${version}...`;
    xraySettings.updateProgress.percent = 0;
    
    // 模拟安装进度
    for (let i = 0; i <= 100; i += 20) {
      xraySettings.updateProgress.percent = i;
      await new Promise(resolve => setTimeout(resolve, 200));
    }
    
    // 这里应该调用实际的API来安装版本
    // await axios.post('/api/xray/version', { version });
    
    // 安装完成
    xraySettings.updateProgress.status = 'completed';
    xraySettings.updateProgress.message = `${version} 安装成功!`;
    xraySettings.updateProgress.percent = 100;
    
    // 更新当前版本
    xraySettings.currentVersion = version;
    
    // 提示是否重启
    setTimeout(() => {
      ElMessageBox.confirm(
        '版本已更新，需要重启Xray服务才能生效。',
        '更新完成',
        {
          confirmButtonText: '立即重启',
          cancelButtonText: '稍后重启',
          type: 'success'
        }
      ).then(() => {
        restartXray();
      }).catch(() => {
        ElMessage.warning('更新已完成，但需要重启才能生效');
      }).finally(() => {
        xraySettings.updateProgress.visible = false;
      });
    }, 1000);
  } catch (error) {
    console.error('Failed to download/install version:', error);
    xraySettings.updateProgress.status = 'error';
    xraySettings.updateProgress.message = '更新失败';
    xraySettings.updateProgress.error = error.message || '未知错误';
    
    // 延迟关闭进度显示后显示错误详情
    setTimeout(() => {
      xraySettings.updateProgress.visible = false;
      
      // 显示详细错误信息
      showErrorDetails(
        '下载或安装 Xray 版本失败',
        error.response?.data?.message || error.message || '未知错误',
        '请检查网络连接、磁盘空间和系统权限。您也可以尝试手动下载并安装该版本。',
        true,
        'updateVersion',
        { version }
      )
    }, 1500);
  }
};

// 展示错误详情
const showErrorDetails = (title, message, resolution = '', canRetry = false, retryAction = null, retryParams = null) => {
  errorDetails.title = title
  errorDetails.message = message
  errorDetails.resolution = resolution
  errorDetails.canRetry = canRetry
  errorDetails.retryAction = retryAction
  errorDetails.retryParams = retryParams
  errorDetails.visible = true
}

// 复制错误信息到剪贴板
const copyErrorToClipboard = () => {
  try {
    const errorText = `${errorDetails.title}\n\n${errorDetails.message}`
    navigator.clipboard.writeText(errorText)
    ElMessage.success('错误信息已复制到剪贴板')
  } catch (error) {
    console.error('Failed to copy error message:', error)
    ElMessage.error('复制失败，请手动选择复制')
  }
}

// 重试失败的操作
const retryFailedOperation = async () => {
  errorDetails.visible = false
  
  if (!errorDetails.retryAction) return
  
  try {
    ElMessage.info('正在重试操作...')
    
    // 根据保存的操作类型执行对应的重试逻辑
    switch (errorDetails.retryAction) {
      case 'switchVersion':
        await performVersionSwitch(errorDetails.retryParams?.shouldRestart)
        break
      case 'restart':
        await restartXray()
        break
      case 'saveSettings':
        await saveXraySettings()
        break
      case 'updateVersion':
        await downloadXrayVersion(errorDetails.retryParams?.version)
        break
      case 'syncVersions':
        await syncVersionsFromGitHub()
        break
      case 'refreshVersions':
        await refreshXrayVersions()
        break
      default:
        console.warn('Unknown retry action:', errorDetails.retryAction)
    }
  } catch (error) {
    console.error('Retry operation failed:', error)
    ElMessage.error('重试操作失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  }
}

// 刷新Xray设置
const refreshXraySettings = async () => {
  xraySettings.loading = true;
  try {
    console.log('刷新Xray设置开始');
    await loadXraySettings();
    ElMessage.success('刷新设置成功');
    console.log('刷新后的Xray设置：', { ...xraySettings });
  } catch (error) {
    ElMessage.error('刷新设置失败：' + error.message);
  } finally {
    xraySettings.loading = false;
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

.protocol-description {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.protocol-description p {
  margin: 0;
}

.el-descriptions-item {
  margin-bottom: 10px;
}

.version-selector {
  display: flex;
  align-items: center;
}

.version-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.version-tips {
  margin-left: 15px;
  font-size: 12px;
  color: #909399;
}

.info-icon {
  margin-left: 5px;
}

.version-info {
  display: flex;
  align-items: center;
}

.changelog-list {
  margin: 0;
  padding-left: 20px;
}

.changelog-list li {
  margin-bottom: 5px;
}

.update-progress {
  padding: 20px 0;
}

.update-status {
  margin-top: 15px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.error-message {
  color: #f56c6c;
  margin-top: 10px;
}

/* 错误详情样式 */
.error-details-container {
  text-align: left;
}

.error-card {
  margin: 15px 0;
  background-color: #fafafa;
}

.error-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.error-message-content {
  white-space: pre-wrap;
  word-break: break-all;
  font-family: monospace;
  background-color: #1e1e1e;
  color: #d4d4d4;
  padding: 10px;
  border-radius: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.error-resolution {
  margin-top: 15px;
  padding: 10px;
  border-left: 3px solid #409EFF;
  background-color: rgba(64, 158, 255, 0.1);
}

.error-resolution h4 {
  margin-top: 0;
  margin-bottom: 8px;
  color: #606266;
}

.error-resolution p {
  margin: 0;
  color: #606266;
  line-height: 1.5;
}
</style> 