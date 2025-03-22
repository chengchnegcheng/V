# V - 多协议代理管理面板

<div align="center">
  
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8.svg)
![Vue Version](https://img.shields.io/badge/Vue-3.x-4FC08D.svg)
![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)

</div>

一个现代化的多协议代理管理面板，基于 Go 和 Vue 3 开发，提供强大的用户管理、流量控制和实时监控功能。

## 📋 核心功能

- **多协议支持**：集成 VMess、VLESS、Trojan、Shadowsocks 等多种协议
- **用户管理系统**：多用户支持与精细化权限控制
- **实时监控**：CPU、内存、网络等系统资源使用情况可视化展示
- **流量控制**：精确统计、限制和可视化用户流量
- **证书管理**：自动申请与管理 SSL 证书，确保连接安全
- **数据备份**：支持数据库备份和恢复，保障数据安全
- **现代界面**：基于 Element Plus 的响应式设计，操作便捷直观

## 🚀 快速安装

### 系统要求

| 组件 | 最低要求 | 推荐配置 |
|------|---------|---------|
| 操作系统 | Linux / macOS / Windows | Ubuntu 20.04+ |
| CPU | 1核 | 2核+ |
| 内存 | 1GB | 2GB+ |
| 存储 | 10GB | 20GB+ |
| Go 环境 | 1.16+ | 1.18+ |
| Node.js* | 16+ | 18+ |

*仅前端开发需要

### Windows 安装

1. 从 [Releases 页面](https://github.com/chengchnegcheng/V/releases/latest) 下载最新版本的 `v.exe`
2. 双击运行 `v.exe` 文件
3. 通过浏览器访问 `http://localhost:8080`

### Linux/macOS 安装

```bash
# 下载最新版本
wget https://github.com/chengchnegcheng/V/releases/latest/download/v-linux-amd64.tar.gz

# 解压文件
tar -zxvf v-linux-amd64.tar.gz

# 进入目录并运行
cd V && ./v
```

### Docker 安装

```bash
# 拉取镜像
docker pull chengcheng/v-panel:latest

# 运行容器
docker run -d --name v-panel \
  -p 8080:8080 \
  -v $PWD/data:/app/data \
  --restart unless-stopped \
  chengcheng/v-panel:latest
```

## 📝 使用指南

1. 安装并启动程序后，访问以下地址：
   - Web 管理面板：`http://[服务器IP]:8080`
   - 开发模式前端：`http://localhost:3000`

2. 使用默认账号登录：
   - 用户名：`admin`
   - 密码：`admin`

3. **重要**：首次登录后请立即修改默认密码以确保安全

4. 通过面板可以：
   - 创建和管理代理协议
   - 添加和管理用户账号
   - 监控系统资源和流量使用情况
   - 管理 SSL 证书
   - 配置备份与恢复

## 🖥️ 界面预览

<div align="center">
  <img src="docs/screenshots/dashboard.png" alt="仪表盘" width="45%">
  <img src="docs/screenshots/traffic.png" alt="流量监控" width="45%">
</div>

## ⚙️ 高级配置

配置文件位于 `config/settings.json`，主要配置项包括：

| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| server.port | HTTP 服务端口 | 8080 |
| server.address | 监听地址 | 0.0.0.0 |
| database.type | 数据库类型 | sqlite |
| database.path | 数据库文件路径 | ./data/v.db |
| log.level | 日志级别 | info |
| ssl.auto | 自动申请证书 | true |
| admin.username | 管理员用户名 | admin |

## 🛠️ 开发指南

### 后端开发

```bash
# 克隆仓库
git clone https://github.com/chengchnegcheng/V.git

# 进入项目目录
cd V

# 编译后端
go build -o v

# 运行
./v
```

### 前端开发

```bash
# 进入前端目录
cd web

# 安装依赖
npm install

# 开发模式运行
npm run dev

# 构建生产版本
npm run build
```

## 🔧 技术栈

### 后端
- **语言**：Go
- **数据库**：SQLite
- **API**：RESTful

### 前端
- **框架**：Vue 3
- **状态管理**：Pinia
- **UI库**：Element Plus
- **图表**：ECharts
- **API请求**：Axios

## ❓ 常见问题解答

<details>
<summary><b>如何更改默认端口？</b></summary>
<p>启动时使用 <code>--port 新端口号</code> 参数，或修改配置文件中的 server.port 值。</p>
</details>

<details>
<summary><b>如何配置自动启动？</b></summary>

**Linux (systemd)**:
```bash
# 创建服务文件
cat > /etc/systemd/system/v-panel.service << EOF
[Unit]
Description=V Panel Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/to/V
ExecStart=/path/to/V/v
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
systemctl daemon-reload
systemctl enable v-panel
systemctl start v-panel
```

**Windows**:
- 使用任务计划程序创建开机自启动任务
</details>

<details>
<summary><b>如何备份数据？</b></summary>
<p>
1. 通过 Web 界面：导航至"系统管理" > "备份"页面，点击"创建备份"按钮<br>
2. 手动备份：复制 <code>data/v.db</code> 文件到安全位置
</p>
</details>

<details>
<summary><b>申请 SSL 证书失败怎么办？</b></summary>
<p>
1. 确保域名正确解析到服务器IP<br>
2. 检查80/443端口是否开放<br>
3. 查看日志文件分析错误原因<br>
4. 尝试手动上传已有证书
</p>
</details>

## 📊 开发状态

| 功能 | 状态 | 备注 |
|------|------|------|
| 核心功能 | ✅ 已完成 | 所有基础功能可用 |
| 多协议支持 | ✅ 已完成 | 已支持主流协议 |
| 状态管理 | ✅ 已完成 | 已从Vuex迁移到Pinia |

## 🤝 贡献指南

欢迎提交问题报告和功能请求！如果您想贡献代码：


## 📚 相关文档

- [完整API文档](docs/api/README.md)
- [故障排除指南](docs/troubleshooting.md)
- [SSL证书管理](docs/ssl_guide.md)
- [安全最佳实践](docs/security.md)

## 📜 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件 