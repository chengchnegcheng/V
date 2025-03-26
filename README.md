# V - 高性能代理服务器

<div align="center">
  <img src="docs/images/logo.png" alt="V Logo" width="200">
  <p>
    <a href="#功能特点">功能特点</a> •
    <a href="#快速开始">快速开始</a> •
    <a href="#开发指南">开发指南</a> •
    <a href="#文档">文档</a>
  </p>
</div>

V是一个用Go语言编写的高性能代理服务器，基于Xray-core，支持多种代理协议，包括Shadowsocks、VMess、Trojan等。它提供了完整的用户管理、流量统计、证书管理等功能，以及直观的Web管理界面。

## 功能特点

### 核心功能
- 多协议支持
  - Shadowsocks
  - VMess
  - Trojan
  - 更多协议支持计划中

- 用户管理
  - 用户认证和授权
  - 流量限制和统计
  - 用户状态监控
  - 多级用户权限

- 流量管理
  - 实时流量统计
  - 每日流量统计
  - 流量限制和警告
  - 协议级别的流量统计

- 证书管理
  - 自动SSL证书申请和更新
  - 多域名证书支持
  - 证书验证和状态监控

- Xray-core管理
  - 版本切换
  - 远程更新
  - 运行状态监控

- 系统管理
  - 完整的日志系统
  - 系统状态监控
  - 配置管理
  - 通知系统

## 快速开始

### 系统要求
- Go 1.16或更高版本
- Node.js 16+和npm（用于前端开发）
- 支持的操作系统：
  - Linux
  - Windows
  - macOS

### 安装方式

#### 1. 直接下载运行
1. 从Releases页面下载最新版本
   - Windows: `v-windows-amd64.exe`
   - Linux: `v-linux-amd64`
   - macOS: `v-darwin-amd64`

2. 运行程序：
   ```bash
   # Windows
   v-windows-amd64.exe

   # Linux/macOS
   chmod +x v-linux-amd64  # 或 v-darwin-amd64
   ./v-linux-amd64        # 或 ./v-darwin-amd64
   ```

#### 2. 从源码编译
1. 克隆仓库：
   ```bash
   git clone https://github.com/your-username/V.git
   cd V
   ```

2. 安装依赖：
   ```bash
   go mod download
   ```

3. 编译：
   ```bash
   # Windows
   go build -o v.exe

   # Linux/macOS
   go build -o v
   ```

4. 运行：
   ```bash
   # Windows
   v.exe

   # Linux/macOS
   ./v
   ```

### 首次运行配置
1. 程序首次运行时会自动创建必要的目录和配置文件：
   - `config/` - 配置文件目录
   - `data/` - 数据文件目录
   - `logs/` - 日志文件目录
   - `xray/` - Xray-core文件目录

2. 访问Web管理界面：
   - 打开浏览器访问 `http://localhost:9000`
   - 默认管理员账号：`admin`
   - 默认密码：`admin123`
   - 首次登录后请立即修改默认密码

3. 配置说明：
   - 数据库文件位于 `data/v.db`
   - 日志文件位于 `logs/`
   - Xray文件位于 `xray/bin/`

### 常见问题
1. 端口被占用
   - 检查9000端口是否被其他程序占用
   - 可以修改代码中的服务器监听端口

2. 权限问题
   - Linux/macOS系统确保有执行权限
   - 确保数据目录有写入权限

3. Xray版本切换问题
   - 如果下载失败，可以手动下载Xray二进制文件并放入`xray/bin/{version}/`目录
   - 国内网络环境可能导致下载速度慢

## 开发指南

### 开发环境搭建

#### 后端开发
1. 克隆仓库：
   ```bash
   git clone https://github.com/your-username/V.git
   cd V
   ```

2. 安装Go依赖：
   ```bash
   go mod download
   ```

3. 开发模式运行：
   ```bash
   go run main.go
   ```

4. 编译：
   ```bash
   go build -o v.exe  # Windows
   go build -o v      # Linux/macOS
   ```

#### 前端开发
1. 进入前端目录：
   ```bash
   cd web
   ```

2. 安装Node.js依赖：
   ```bash
   npm install
   ```

3. 开发模式运行：
   ```bash
   npm run dev
   ```

4. 构建生产版本：
   ```bash
   npm run build
   ```

### 开发环境配置
1. 后端配置：
   - 开发模式日志：`logs/dev.log`
   - 开发模式端口：9000

2. 前端配置：
   - 开发服务器端口：5173
   - API代理配置：`web/vite.config.js`

### 开发工具推荐
1. 后端开发：
   - IDE: GoLand 或 VS Code + Go插件
   - API测试: Postman 或 Insomnia

2. 前端开发：
   - IDE: VS Code
   - Vue.js相关插件

### 开发流程
1. 启动开发环境：
   ```bash
   # 终端1：启动后端
   go run main.go

   # 终端2：启动前端开发服务器
   cd web
   npm run dev
   ```

2. 访问开发环境：
   - 后端API: `http://localhost:9000`
   - 前端页面: `http://localhost:5173`

3. 开发调试：
   - 后端日志实时查看
   - 前端热更新：修改代码后自动刷新
   - API调试：使用浏览器开发者工具

### 部署
1. 构建生产版本：
   ```bash
   # 构建后端
   go build -o v

   # 构建前端
   cd web
   npm run build
   ```

2. 部署文件：
   - 后端：`v` 可执行文件
   - 前端：`web/dist` 目录下的静态文件

## 文档

### 目录结构
```
v/
├── api/            # API实现
├── common/         # 通用代码
├── config/         # 配置文件和配置管理
├── db/             # 数据库相关代码
├── errors/         # 错误定义
├── logger/         # 日志系统
├── middleware/     # HTTP中间件
├── model/          # 数据模型
├── monitor/        # 系统监控
├── notification/   # 通知系统
├── proxy/          # 代理协议实现
├── router/         # 路由定义
├── server/         # HTTP服务器
├── settings/       # 设置管理
├── ssl/            # SSL证书管理
├── tools/          # 工具脚本
├── utils/          # 工具函数
├── web/            # 前端代码
│   ├── dist/       # 构建输出
│   ├── public/     # 静态资源
│   └── src/        # 前端源码
│       ├── api/    # API调用
│       ├── assets/ # 资源文件
│       ├── components/ # 组件
│       ├── router/ # 路由
│       ├── stores/ # 状态管理
│       ├── utils/  # 工具函数
│       └── views/  # 页面
└── xray/           # Xray相关代码
```

### API文档

#### 系统API
- `GET /api/system/info` - 获取系统信息
- `GET /api/system/status` - 获取系统状态

#### Xray管理API
- `GET /api/xray/versions` - 获取支持的Xray版本
- `POST /api/xray/version` - 切换Xray版本
- `POST /api/xray/start` - 启动Xray
- `POST /api/xray/stop` - 停止Xray
- `POST /api/xray/restart` - 重启Xray

#### 用户管理API
- `POST /api/auth/login` - 用户登录
- `GET /api/auth/user` - 获取当前用户信息
- `POST /api/auth/logout` - 用户登出

## 特别鸣谢

- [Xray-core](https://github.com/XTLS/Xray-core) - 核心代理引擎
- [Vue.js](https://vuejs.org/) - 前端框架
- [Element Plus](https://element-plus.org/) - UI组件库
- [ECharts](https://echarts.apache.org/) - 图表库

## 贡献

欢迎提交Issue和Pull Request。

## 许可证

MIT License 