# V - 多协议代理面板

<div align="center">
  
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8.svg)
![Vue Version](https://img.shields.io/badge/Vue-3.x-4FC08D.svg)
![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)

</div>

一个基于 x-ui 项目重新开发的现代化多协议代理管理面板，提供强大的多用户支持、流量控制和实时监控功能。

## 📋 主要特性

- **多协议支持**：vmess、vless、trojan、shadowsocks、dokodemo-door、socks、http
- **用户管理**：多用户支持与精细化权限控制
- **系统监控**：实时监控 CPU、内存、网络等系统资源使用情况
- **流量控制**：精确统计、限制和可视化用户流量
- **SSL 证书**：自动申请与管理，确保连接安全
- **备份恢复**：数据库备份和恢复功能，确保数据安全
- **现代界面**：响应式设计，简洁直观的操作体验

## 🖥️ 界面预览

<div align="center">
  <img src="docs/screenshots/dashboard.png" alt="仪表盘" width="45%">
  <img src="docs/screenshots/traffic.png" alt="流量监控" width="45%">
</div>

## 🚀 快速开始

### 系统要求

| 组件 | 最低要求 | 推荐配置 |
|------|---------|---------|
| 操作系统 | Linux (Ubuntu 16.04+, CentOS 7+) / macOS / Windows | Ubuntu 20.04+ |
| CPU | 1核 | 2核+ |
| 内存 | 1GB | 2GB+ |
| 存储 | 10GB | 20GB+ |
| Go环境 | 1.16+ | 1.18+ |
| 数据库 | SQLite (内置) | SQLite (内置) |
| Node.js* | 16+ | 18+ |
| npm* | 7+ | 9+ |

*仅前端开发需要

### Windows 安装

1. 从[releases页面](https://github.com/chengchnegcheng/V/releases/latest)下载最新版本的`v.exe`
2. 双击运行`v.exe`文件
3. 通过浏览器访问 `http://localhost:8080`

### Linux 安装

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
docker run -d --name v-panel -p 8080:8080 -v $PWD/data:/app/data chengcheng/v-panel:latest
```

## 🛠️ 开发指南

### 从源码编译

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

> **注意**：npm 命令必须在 web 目录下运行，在项目根目录运行将会失败。

## 👨‍💻 技术栈

### 后端
- **语言**：Go
- **数据库**：SQLite
- **API**：RESTful

### 前端
- **框架**：Vue 3
- **状态管理**：Pinia (已从Vuex迁移)
- **路由**：Vue Router 4
- **UI库**：Element Plus
- **图表**：ECharts
- **HTTP请求**：Axios

## 📝 使用指南

1. 启动程序后，访问：
   - 后端API服务器：`http://[服务器IP]:8080`
   - 开发模式下前端：`http://localhost:3000`
2. 使用默认账号登录：
   - 用户名：`admin`
   - 密码：`admin`
3. 首次登录后请立即修改默认密码

## ⚙️ 配置说明

配置文件位于 `config/settings.json`，包含以下主要部分：

- **server**：服务器配置（端口、地址等）
- **database**：数据库配置
- **log**：日志配置
- **ssl**：SSL证书配置
- **admin**：管理员账号设置

## 📊 开发状态

| 功能 | 状态 | 备注 |
|------|------|------|
| 核心功能 | ✅ 已完成 | 所有基础功能均可使用 |
| 前端迁移 | ✅ 已完成 | 已从Vuex迁移到Pinia |
| 开发环境 | ✅ 正常 | 前端构建和开发环境正常工作 |
| 测试覆盖 | ⚠️ 部分完成 | 需增加更多单元测试 |
| 文档完善 | ⚠️ 进行中 | API文档待完善 |

## ❓ 常见问题

<details>
<summary><b>如何修改默认端口？</b></summary>
<p>使用 <code>--listen :新端口号</code> 启动或修改配置文件中的 port 字段。</p>
</details>

<details>
<summary><b>如何备份数据？</b></summary>
<p>备份 <code>data/v.db</code> 文件或使用界面中的备份功能。系统支持自动备份和手动备份。</p>
</details>

<details>
<summary><b>如何设置自动启动？</b></summary>
<p>配置系统服务，详见<a href="troubleshooting_guide.md">故障排除指南</a>。</p>
</details>

<details>
<summary><b>SSL证书更新失败怎么办？</b></summary>
<p>系统会自动处理证书更新，如果失败可在界面手动更新或检查域名DNS解析配置。</p>
</details>

## 📚 相关文档

- [故障排除指南](troubleshooting_guide.md)
- [SSL证书管理](ssl_implementation_plan.md)
- [完整开发文档](development_summary.md)

## 🤝 贡献指南

欢迎提交问题和功能请求！如果您想贡献代码：

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启一个 Pull Request

## 📜 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件 