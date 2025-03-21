# V项目

V是一个基于x-ui项目重新开发的多协议、多用户代理面板，提供丰富的系统管理和流量控制功能。

## 功能特点

- **多协议支持**: 支持vmess、vless、trojan、shadowsocks、dokodemo-door、socks和http协议
- **多用户管理**: 支持多用户系统，每个用户可单独设置权限和流量限制
- **系统状态监控**: 实时监控CPU、内存、磁盘和网络流量状态
- **流量统计**: 精确统计每个用户的流量使用情况，支持按协议分类
- **流量控制**: 可设置用户流量限制和到期时间
- **SSL证书管理**: 支持自动申请和更新SSL证书
- **日志系统**: 完整的系统运行日志和错误日志

## 快速开始

### 系统要求

- 操作系统: Linux (Ubuntu 16+, CentOS 7+)或macOS
- 依赖: Go 1.15+

### 安装方法

#### 方法1: 使用自动安装脚本

```bash
bash <(curl -Ls https://github.com/yourusername/V/raw/master/install.sh)
```

#### 方法2: 手动安装

1. 下载最新版本:

```bash
# 创建工作目录
mkdir -p /usr/local/v
cd /usr/local/v

# 下载最新版本
wget https://github.com/yourusername/V/releases/latest/download/v-linux-64.tar.gz
tar -zxvf v-linux-64.tar.gz
```

2. 启动服务:

```bash
# 给予执行权限
chmod +x v

# 启动服务
./v
```

3. 配置系统服务(可选):

```bash
cat > /etc/systemd/system/v.service << EOF
[Unit]
Description=V Panel Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/usr/local/v
ExecStart=/usr/local/v/v
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# 重新加载systemd
systemctl daemon-reload

# 设置开机启动
systemctl enable v

# 启动服务
systemctl start v
```

### 访问管理面板

安装完成后，可通过以下地址访问管理面板:

```
http://你的服务器IP:54321
```

默认用户名和密码:
- 用户名: `admin`
- 密码: `admin`

**首次登录后请立即修改默认密码!**

## 配置指南

### 基本配置

服务启动后，首先进行以下基本配置:

1. 修改管理员密码
2. 配置面板访问端口(默认54321)
3. 配置SSL证书(推荐)

### 添加用户

1. 在管理面板中选择"用户管理"
2. 点击"添加用户"
3. 填写用户信息，包括用户名、密码、流量限制和到期时间
4. 点击"保存"完成创建

### 配置协议

1. 在用户详情页面点击"添加协议"
2. 选择要添加的协议类型(vmess/vless/trojan等)
3. 配置协议参数
4. 点击"保存"完成配置

## 常见问题

### 无法连接到管理面板

- 检查服务是否正常运行: `systemctl status v`
- 检查防火墙是否开放面板端口
- 检查服务器网络连接是否正常

### 流量统计不准确

- 检查流量统计服务是否正常运行
- 确保数据库连接正常
- 查看错误日志: `cat /usr/local/v/logger/error.log`

### SSL证书申请失败

- 确保域名解析已正确设置
- 确保80和443端口未被其他服务占用
- 查看SSL申请日志获取详细错误信息

## 开发指南

V项目使用Go语言开发，前端使用Vue.js框架。如果您想参与开发，请参考[开发指南](development_guide.md)。

## 故障排除

遇到问题时，请参考[故障排除指南](troubleshooting_guide.md)获取详细的诊断和解决方法。

## 数据库接口修复

如果您遇到数据库相关问题，请参考[数据库接口修复指南](db_interface_fix.md)。

## 贡献

欢迎提交Pull Request或Issue来帮助改进V项目。在提交代码前，请确保:

1. 代码遵循Go和Vue.js的编码规范
2. 添加必要的测试和文档
3. 确保所有测试通过

## 许可证

V项目基于[MIT许可证](LICENSE)开源。

## 鸣谢

- [x-ui项目](https://github.com/vaxilu/x-ui)：提供了V项目的基础
- 所有贡献者和用户：感谢您的支持和反馈 