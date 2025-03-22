# V - 多协议代理面板

基于 x-ui 项目重新开发的多协议代理面板，支持多种协议和用户管理功能。

## 主要特性

- 多协议支持: vmess、vless、trojan、shadowsocks、dokodemo-door、socks、http
- 用户管理: 多用户支持与权限控制
- 系统监控: 实时监控系统资源使用情况
- 流量控制: 精确统计和限制流量
- SSL 证书: 自动申请与管理
- 现代界面: 响应式设计，操作便捷

## 安装指南

### 系统要求

- 操作系统: Linux (Ubuntu 16.04+, CentOS 7+) / macOS / Windows
- Go 环境: 1.16+
- 数据库: SQLite (内置)

### Windows安装

1. 下载最新版本的`v.exe`
2. 直接运行`v.exe`文件

### Linux安装

使用预编译二进制文件:

```bash
# 下载最新版本
wget https://github.com/chengchnegcheng/V/releases/latest/download/v-linux-amd64.tar.gz

# 解压文件
tar -zxvf v-linux-amd64.tar.gz

# 进入目录并运行
cd V && ./v
```

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/chengchnegcheng/V.git

# 进入项目目录
cd V

# 编译
go build -o v

# 运行
./v
```

### 前端开发与构建

如需开发或重新构建前端界面:

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

## 快速使用

1. 启动程序后，访问 `http://[服务器IP]:8080`
2. 默认账号: `admin`, 密码: `admin`
3. 首次登录后请立即修改默认密码

## 常见问题

- **修改默认端口**: 使用`--listen :新端口号`启动或修改配置文件
- **数据备份**: 备份`data/v.db`文件或使用界面中的备份功能
- **自动启动**: 配置系统服务，详见故障排除指南
- **SSL证书更新**: 系统自动处理或在界面手动更新

## 配置文件

配置文件位于`config/settings.json`，包含站点设置、安全设置、代理设置等配置项。

## 更多文档

- [故障排除指南](troubleshooting_guide.md)
- [SSL证书管理](ssl_implementation_plan.md)
- [完整开发文档](development_summary.md)

## 许可证

MIT许可证 - 详见[LICENSE](LICENSE)文件 