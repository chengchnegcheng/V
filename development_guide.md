# V项目开发指南

## 项目概述
V是一个基于x-ui项目重新开发的多协议、多用户代理面板，提供系统状态监控、流量统计、流量限制和到期时间等功能。支持的协议包括vmess、vless、trojan、shadowsocks、dokodemo-door、socks和http。

## 开发环境设置

### 系统要求
- 操作系统：Linux（推荐Ubuntu 16+、CentOS 7+）或macOS
- Go语言：1.15及以上版本
- 数据库：SQLite
- 前端：Vue.js 2.x

### 环境配置步骤
1. 安装Go语言环境
   ```bash
   # 下载并安装Go
   wget https://golang.org/dl/go1.17.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.17.linux-amd64.tar.gz
   
   # 配置环境变量
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   echo 'export GOPATH=$HOME/go' >> ~/.bashrc
   source ~/.bashrc
   ```

2. 克隆项目代码
   ```bash
   git clone https://github.com/yourusername/V.git
   cd V
   ```

3. 安装依赖
   ```bash
   go mod tidy
   ```

4. 编译运行
   ```bash
   go build -o bin/v main.go
   ./bin/v
   ```

## 项目结构

```
V/
├── bin/                # 可执行文件
├── config/             # 配置文件
│   └── config.json     # 主配置文件
├── database/           # 数据库文件
│   └── data.db         # SQLite数据库
├── logger/             # 日志文件
├── media/              # 多媒体资源
├── model/              # 数据模型
│   ├── db.go           # 数据库接口定义
│   └── sqlite.go       # SQLite实现
├── traffic/            # 流量相关
│   └── traffic.go      # 流量统计模块
├── util/               # 工具函数
├── web/                # 前端资源
│   ├── src/            # 前端源代码
│   └── dist/           # 构建后的静态文件
└── main.go             # 程序入口
```

## 常见开发任务

### 数据库操作
V项目使用GORM操作SQLite数据库，主要数据模型在`model`目录下。

示例：添加新的数据模型
```go
// model/model.go
type NewFeature struct {
    gorm.Model
    Name        string
    Description string
    Enabled     bool
}

// 在数据库初始化时添加
func initDatabase() {
    db.AutoMigrate(&User{}, &ProtocolStats{}, &NewFeature{})
}
```

### 添加新API接口
1. 在相应模块中添加处理函数
2. 在`main.go`或相应的路由文件中注册新路由

示例：
```go
// 处理函数
func handleNewFeature(w http.ResponseWriter, r *http.Request) {
    // 处理逻辑
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// 注册路由
http.HandleFunc("/api/new-feature", handleNewFeature)
```

### 前端开发
前端使用Vue.js框架，源代码在`web/src`目录下。

1. 进入前端目录
   ```bash
   cd web
   ```

2. 安装依赖
   ```bash
   npm install
   ```

3. 开发模式运行
   ```bash
   npm run serve
   ```

4. 构建生产版本
   ```bash
   npm run build
   ```

## 调试技巧

### 后端调试
使用Go的内置日志包或第三方日志库记录调试信息：

```go
import "log"

func someFunction() {
    log.Println("Debug: entering someFunction")
    // 函数逻辑
    log.Println("Debug: exiting someFunction")
}
```

### 数据库调试
启用GORM的日志模式查看SQL语句：

```go
db = db.Debug() // 启用调试模式
```

### 前端调试
使用Vue开发者工具和浏览器控制台进行调试。

## 代码风格规范

### Go代码规范
- 使用`gofmt`或`goimports`格式化代码
- 遵循[Effective Go](https://golang.org/doc/effective_go)的建议
- 使用有意义的变量名和函数名
- 为公共函数和类型添加注释

### 提交规范
- 使用有意义的提交信息
- 每个提交专注于一个功能或修复
- 提交前运行测试确保代码正常工作

## 已知问题和解决方案

### 方法重复声明问题
在`model/sqlite.go`文件中存在方法重复声明问题，遵循`fixed_sqlite_notes.txt`中的指南进行修复。

### GORM风格API误用
项目中一些地方错误地使用了GORM风格的API，但实际上使用的是原生SQL。请确保正确使用数据库API。

### 时间字段处理
涉及时间的字段处理需要特别注意，确保使用指针类型以正确处理空值。

## 打包和部署

### 构建可执行文件
```bash
go build -o bin/v main.go
```

### 创建系统服务
在Linux系统中，可以创建systemd服务：

```bash
sudo cat > /etc/systemd/system/v.service << EOF
[Unit]
Description=V Panel Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/path/to/V
ExecStart=/path/to/V/bin/v
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable v
sudo systemctl start v
```

## 贡献指南
1. Fork项目仓库
2. 创建功能分支: `git checkout -b feature/your-feature`
3. 提交更改: `git commit -am 'Add some feature'`
4. 推送到分支: `git push origin feature/your-feature`
5. 提交Pull Request

## 联系与支持
如有问题或需要支持，请通过以下方式联系：
- 提交GitHub Issue
- 发送邮件至：support@example.com 