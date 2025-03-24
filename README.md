# V - 高性能代理服务器

V是一个用Go语言编写的高性能代理服务器，支持多种代理协议，包括Shadowsocks、VMess、Trojan等。它提供了完整的用户管理、流量统计、证书管理等功能。

## 功能特点

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

## 系统要求

- Go 1.16或更高版本
- SQLite3
- 支持的操作系统：
  - Linux
  - Windows
  - macOS

## 安装

1. 克隆仓库：
```bash
git clone https://github.com/yourusername/v.git
cd v
```

2. 安装依赖：
```bash
go mod download
```

3. 编译：
```bash
go build -o v
```

4. 运行：
```bash
./v
```

## 配置

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

## 目录结构

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

## API文档

### 用户管理API

- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users` - 获取用户列表
- `GET /api/v1/users/:id` - 获取用户详情
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

### 流量统计API

- `GET /api/v1/stats/traffic/:user_id` - 获取用户流量统计
- `GET /api/v1/stats/daily/:user_id` - 获取用户每日流量统计
- `GET /api/v1/stats/protocol/:protocol_id` - 获取协议流量统计

### 证书管理API

- `POST /api/v1/certificates` - 创建证书
- `GET /api/v1/certificates` - 获取证书列表
- `GET /api/v1/certificates/:id` - 获取证书详情
- `DELETE /api/v1/certificates/:id` - 删除证书

## 开发

### 添加新协议

1. 在`proxy/protocols`目录下创建新的协议包
2. 实现`Protocol`接口
3. 在`proxy/proxy.go`中注册新协议

### 添加新功能

1. 在相应的包中实现功能
2. 添加必要的API端点
3. 更新配置和文档

## 贡献

欢迎提交Issue和Pull Request。在提交代码前，请确保：

1. 代码符合Go代码规范
2. 添加了必要的测试
3. 更新了相关文档

## 许可证

MIT License 