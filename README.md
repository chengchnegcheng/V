# V - 多协议代理面板

V 是一个功能强大的多协议代理面板，基于 x-ui 项目重新开发，提供更优雅的用户界面和更完善的功能体验。

## 主要功能

- **多协议支持**: 支持 vmess、vless、trojan、shadowsocks、dokodemo-door、socks 和 http 协议
- **用户管理**: 完善的多用户支持，权限分级控制
- **系统监控**: 实时监控 CPU、内存、网络和磁盘使用情况
- **流量统计与限制**: 精确的流量统计和自定义流量限制
- **SSL 证书管理**: 自动申请、更新和管理 SSL 证书
- **日志系统**: 详细的系统和操作日志记录
- **美观界面**: 现代化、响应式的管理界面

## 安装要求

- 操作系统: Linux (Ubuntu 16.04+, CentOS 7+) / macOS / Windows
- Go 语言环境: 1.16 或更高版本
- 数据库: SQLite (内置)

## 快速安装

### 使用预编译二进制文件

```bash
# 下载最新版本
wget https://github.com/yourusername/v/releases/latest/download/v-linux-amd64.tar.gz

# 解压文件
tar -zxvf v-linux-amd64.tar.gz

# 进入目录
cd v

# 运行
./v
```

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/yourusername/v.git

# 进入项目目录
cd v

# 编译
go build -o v

# 运行
./v
```

## 使用说明

1. 安装并启动 V 后，访问 `http://your_server_ip:8080` 打开管理界面
2. 默认管理员账号: `admin`，密码: `admin`，初次登录请立即修改密码
3. 在管理界面中，可以进行以下操作:
   - 添加和管理用户
   - 配置代理协议
   - 监控系统状态
   - 查看流量统计
   - 管理 SSL 证书
   - 配置系统设置

## 配置说明

配置文件位于 `config/settings.json`，主要配置项包括:

- 监听地址和端口
- 数据库设置
- 日志级别
- 通知设置 (Email)
- 证书设置

## 常见问题

1. **如何更改默认端口?**
   修改配置文件中的 `listen` 字段，或使用命令行参数 `--listen :新端口号`

2. **如何备份数据?**
   数据存储在 `data/v.db` 文件中，备份此文件即可

3. **如何设置自动启动?**
   请参考 [安装文档](docs/installation.md) 中的系统服务配置部分

## 开发计划

- [ ] 国际化支持
- [ ] 两因素认证
- [ ] 更多优化和性能提升
- [ ] 移动端应用

## 贡献代码

欢迎提交 Pull Request 或 Issue 来帮助改进项目。在提交 PR 前，请确保你的代码:

1. 遵循 Go 代码规范
2. 通过所有测试
3. 包含必要的文档

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件 