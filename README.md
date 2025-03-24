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

V是一个用Go语言编写的高性能代理服务器，支持多种代理协议，包括Shadowsocks、VMess、Trojan等。它提供了完整的用户管理、流量统计、证书管理等功能。

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

- 系统管理
  - 完整的日志系统
  - 系统状态监控
  - 配置管理
  - 通知系统

## 快速开始

### 系统要求
- Go 1.16或更高版本
- SQLite3
- 支持的操作系统：
  - Linux
  - Windows
  - macOS

### 安装方式

#### 1. 直接下载运行
1. 从 [Releases](https://github.com/chengchnegcheng/V/releases) 页面下载最新版本
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
   git clone https://github.com/chengchnegcheng/V.git
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

#### 3. 使用Docker
1. 拉取镜像：
   ```bash
   docker pull chengchnegcheng/v:latest
   ```

2. 运行容器：
   ```bash
   docker run -d \
     --name v \
     -p 8080:8080 \
     -v $PWD/data:/app/data \
     -v $PWD/config:/app/config \
     --restart unless-stopped \
     chengchnegcheng/v:latest
   ```

### 首次运行配置
1. 程序首次运行时会自动创建必要的目录和配置文件：
   - `config/` - 配置文件目录
   - `data/` - 数据文件目录
   - `logs/` - 日志文件目录

2. 访问Web管理界面：
   - 打开浏览器访问 `http://localhost:8080`
   - 默认管理员账号：`admin`
   - 默认密码：`admin`
   - 首次登录后请立即修改默认密码

3. 配置说明：
   - 配置文件位于 `config/config.yaml`
   - 数据库文件位于 `data/v.db`
   - 日志文件位于 `logs/`

### 常见问题
1. 端口被占用
   - 检查8080端口是否被其他程序占用
   - 可以在配置文件中修改端口号

2. 权限问题
   - Linux/macOS系统确保有执行权限
   - 确保数据目录有写入权限

3. 数据库问题
   - 确保SQLite3已安装
   - 检查数据目录权限

## 开发指南

### 开发环境搭建

#### 后端开发
1. 克隆仓库：
   ```bash
   git clone https://github.com/chengchnegcheng/V.git
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
   # Windows
   go build -o v.exe

   # Linux/macOS
   go build -o v
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
   - 配置文件：`config/config.yaml`
   - 开发模式日志：`logs/dev.log`
   - 开发模式端口：8080

2. 前端配置：
   - 开发服务器端口：3000
   - API代理配置：`web/vite.config.ts`
   - 环境变量：`web/.env.development`

### 开发工具推荐
1. 后端开发：
   - IDE: GoLand 或 VS Code + Go插件
   - API测试: Postman 或 Insomnia
   - 数据库工具: DB Browser for SQLite

2. 前端开发：
   - IDE: VS Code
   - 浏览器插件: Vue.js devtools
   - 代码格式化: Prettier

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
   - 后端API: `http://localhost:8080`
   - 前端页面: `http://localhost:3000`

3. 开发调试：
   - 后端日志实时查看：`tail -f logs/dev.log`
   - 前端热更新：修改代码后自动刷新
   - API调试：使用Postman或浏览器开发者工具

### 测试
1. 后端测试：
   ```bash
   # 运行所有测试
   go test ./...

   # 运行特定包的测试
   go test ./database/...

   # 运行带覆盖率的测试
   go test -cover ./...
   ```

2. 前端测试：
   ```bash
   cd web
   npm run test
   ```

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
   - 配置文件：`config/config.yaml`
   - 数据文件：`data/v.db`

## 文档

### 目录结构
```
v/
├── cmd/            # 命令行入口
├── config/         # 配置文件和配置管理
├── database/       # 数据库相关代码
├── logger/         # 日志系统
├── model/          # 数据模型
├── notification/   # 通知系统
├── proxy/          # 代理协议实现
├── server/         # HTTP服务器和API
├── settings/       # 设置管理
├── ssl/            # SSL证书管理
├── stats/          # 流量统计
└── utils/          # 工具函数
```

### API文档

#### 用户管理API
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users` - 获取用户列表
- `GET /api/v1/users/:id` - 获取用户详情
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

#### 流量统计API
- `GET /api/v1/stats/traffic/:user_id` - 获取用户流量统计
- `GET /api/v1/stats/daily/:user_id` - 获取用户每日流量统计
- `GET /api/v1/stats/protocol/:protocol_id` - 获取协议流量统计

#### 证书管理API
- `POST /api/v1/certificates` - 创建证书
- `GET /api/v1/certificates` - 获取证书列表
- `GET /api/v1/certificates/:id` - 获取证书详情
- `DELETE /api/v1/certificates/:id` - 删除证书

### 配置说明
配置文件位于`config/config.yaml`，包含以下主要配置项：

- 服务器设置
  - 监听地址和端口
  - TLS配置
  - 协议设置

- 数据库设置
  - 数据库类型
  - 连接参数

- 用户设置
  - 默认流量限制
  - 用户权限

- 通知设置
  - 邮件通知
  - 其他通知方式

## 贡献

欢迎提交Issue和Pull Request。在提交代码前，请确保：

1. 代码符合Go代码规范
2. 添加了必要的测试
3. 更新了相关文档

## 许可证

MIT License 