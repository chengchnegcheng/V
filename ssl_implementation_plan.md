# SSL证书管理功能实现计划

## 1. 概述

SSL证书管理是V项目的重要组成部分，用于自动申请、更新和管理HTTPS所需的SSL证书。此功能将使用Let's Encrypt作为证书颁发机构，通过ACME协议自动完成证书的申请和更新过程。

## 2. 功能需求

- 自动申请新的SSL证书
- 自动更新即将过期的证书
- 证书状态监控和告警
- 支持多域名证书管理
- 提供手动添加自定义证书的功能
- 前端界面展示和管理

## 3. 技术选型

- 使用Go语言的[acme/autocert](https://golang.org/x/crypto/acme/autocert)包或[lego](https://github.com/go-acme/lego)库
- 使用文件系统存储证书文件
- 数据库中存储证书元数据

## 4. 数据模型

已有`Certificate`模型（在`model/model.go`中），但可能需要调整或扩展：

```go
// Certificate SSL证书信息
type Certificate struct {
    Base
    Domain        string    `json:"domain" db:"domain"`
    CertFile      string    `json:"cert_file" db:"cert_file"`
    KeyFile       string    `json:"key_file" db:"key_file"`
    Status        string    `json:"status" db:"status"`
    LastCheckedAt time.Time `json:"last_checked_at" db:"last_checked_at"`
    LastRenewedAt time.Time `json:"last_renewed_at" db:"last_renewed_at"`
    ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
    AutoRenew     bool      `json:"auto_renew" db:"auto_renew"`
    Type          string    `json:"type" db:"type"` // "letsencrypt" 或 "custom"
}
```

## 5. 数据库接口扩展

在`model/db.go`中添加以下接口方法：

```go
// 证书相关
CreateCertificate(cert *Certificate) error
GetCertificate(id int64) (*Certificate, error)
GetCertificateByDomain(domain string) (*Certificate, error)
UpdateCertificate(cert *Certificate) error
DeleteCertificate(id int64) error
ListCertificates(page, pageSize int) ([]*Certificate, error)
GetExpiringCertificates(within time.Duration) ([]*Certificate, error)
```

## 6. 详细实现步骤

### 6.1 创建SSL证书管理模块

1. 创建`ssl/manager.go`文件:

```go
package ssl

import (
    "crypto/tls"
    "fmt"
    "log/slog"
    "sync"
    "time"
    
    "github.com/go-acme/lego/v4/certcrypto"
    "github.com/go-acme/lego/v4/certificate"
    "github.com/go-acme/lego/v4/challenge/http01"
    "github.com/go-acme/lego/v4/lego"
    "github.com/go-acme/lego/v4/registration"
    
    "v/model"
    "v/notification"
)

// Manager SSL证书管理器
type Manager struct {
    logger      *slog.Logger
    db          model.DB
    certStorage string        // 证书存储路径
    checkInterval time.Duration  // 检查证书有效期的间隔
    renewBefore time.Duration    // 提前多久更新证书
    notifier    notification.Notifier
    
    mu          sync.Mutex
    stop        chan struct{}
    wg          sync.WaitGroup
}

// New 创建SSL证书管理器
func New(logger *slog.Logger, db model.DB, notifier notification.Notifier, certStorage string) *Manager {
    return &Manager{
        logger:      logger,
        db:          db,
        certStorage: certStorage,
        checkInterval: 24 * time.Hour,  // 每日检查一次
        renewBefore:   30 * 24 * time.Hour, // 提前30天更新
        notifier:    notifier,
        stop:        make(chan struct{}),
    }
}

// Start 启动SSL证书管理服务
func (m *Manager) Start() {
    m.wg.Add(1)
    go m.run()
}

// Stop 停止SSL证书管理服务
func (m *Manager) Stop() {
    close(m.stop)
    m.wg.Wait()
}

// run 运行SSL证书检查和更新服务
func (m *Manager) run() {
    defer m.wg.Done()
    
    ticker := time.NewTicker(m.checkInterval)
    defer ticker.Stop()
    
    // 立即检查一次
    m.checkCertificates()
    
    for {
        select {
        case <-ticker.C:
            m.checkCertificates()
        case <-m.stop:
            return
        }
    }
}

// checkCertificates 检查所有证书的有效期
func (m *Manager) checkCertificates() {
    certs, err := m.db.GetExpiringCertificates(m.renewBefore)
    if err != nil {
        m.logger.Error("Failed to get expiring certificates", "error", err)
        return
    }
    
    for _, cert := range certs {
        if cert.AutoRenew && cert.Type == "letsencrypt" {
            if err := m.renewCertificate(cert); err != nil {
                m.logger.Error("Failed to renew certificate", "domain", cert.Domain, "error", err)
                m.notifier.SendAlert("证书更新失败", fmt.Sprintf("域名 %s 的证书更新失败: %v", cert.Domain, err))
            }
        } else {
            // 发送提醒但不自动更新
            m.notifier.SendAlert("证书即将过期", fmt.Sprintf("域名 %s 的证书将在 %s 过期", cert.Domain, cert.ExpiresAt.Format("2006-01-02")))
        }
    }
}

// RequestCertificate 请求新证书
func (m *Manager) RequestCertificate(domain string, email string) error {
    // 实现证书申请逻辑
}

// renewCertificate 更新证书
func (m *Manager) renewCertificate(cert *model.Certificate) error {
    // 实现证书更新逻辑
}

// LoadCertificate 加载SSL证书
func (m *Manager) LoadCertificate(domain string) (*tls.Certificate, error) {
    // 实现证书加载逻辑
}
```

2. 创建`ssl/client.go`文件处理ACME客户端:

```go
package ssl

import (
    "crypto"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "fmt"
    
    "github.com/go-acme/lego/v4/certcrypto"
    "github.com/go-acme/lego/v4/lego"
    "github.com/go-acme/lego/v4/registration"
)

// User 实现acme.User接口
type User struct {
    Email        string
    Registration *registration.Resource
    Key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
    return u.Email
}

func (u *User) GetRegistration() *registration.Resource {
    return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
    return u.Key
}

// createACMEClient 创建ACME客户端
func createACMEClient(email string) (*lego.Client, error) {
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        return nil, err
    }
    
    user := &User{
        Email: email,
        Key:   privateKey,
    }
    
    config := lego.NewConfig(user)
    config.CADirURL = lego.LEDirectoryProduction // 或 lego.LEDirectoryStaging 用于测试
    config.Certificate.KeyType = certcrypto.RSA2048
    
    client, err := lego.NewClient(config)
    if err != nil {
        return nil, err
    }
    
    return client, nil
}
```

### 6.2 实现SSL证书HTTP验证

创建`ssl/http_challenge.go`文件:

```go
package ssl

import (
    "net/http"
    "sync"
    
    "github.com/go-acme/lego/v4/challenge/http01"
)

// HTTPChallengeServer HTTP验证服务器
type HTTPChallengeServer struct {
    server *http.Server
    mu     sync.Mutex
}

// NewHTTPChallengeServer 创建HTTP验证服务器
func NewHTTPChallengeServer() *HTTPChallengeServer {
    return &HTTPChallengeServer{}
}

// Start 启动HTTP验证服务器
func (s *HTTPChallengeServer) Start() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if s.server != nil {
        return nil
    }
    
    provider := http01.NewProviderServer("", "80")
    s.server = &http.Server{
        Addr:    ":80",
        Handler: provider.HTTPHandler(nil),
    }
    
    go func() {
        s.server.ListenAndServe()
    }()
    
    return nil
}

// Stop 停止HTTP验证服务器
func (s *HTTPChallengeServer) Stop() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if s.server == nil {
        return nil
    }
    
    err := s.server.Close()
    s.server = nil
    return err
}
```

### 6.3 实现SSL证书数据库接口

在`model/sqlite.go`中实现证书相关方法:

```go
// CreateCertificate 创建SSL证书记录
func (db *SQLiteDB) CreateCertificate(cert *model.Certificate) error {
    query := `INSERT INTO certificates (domain, cert_file, key_file, status, last_checked_at, last_renewed_at, expires_at, auto_renew, type, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    
    now := time.Now()
    cert.CreatedAt = now
    cert.UpdatedAt = now
    
    result, err := db.db.Exec(
        query,
        cert.Domain,
        cert.CertFile,
        cert.KeyFile,
        cert.Status,
        cert.LastCheckedAt,
        cert.LastRenewedAt,
        cert.ExpiresAt,
        cert.AutoRenew,
        cert.Type,
        cert.CreatedAt,
        cert.UpdatedAt,
    )
    
    if err != nil {
        return err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    
    cert.ID = id
    return nil
}

// GetCertificate 获取SSL证书
func (db *SQLiteDB) GetCertificate(id int64) (*model.Certificate, error) {
    // 实现方法
}

// GetCertificateByDomain 根据域名获取证书
func (db *SQLiteDB) GetCertificateByDomain(domain string) (*model.Certificate, error) {
    // 实现方法
}

// UpdateCertificate 更新SSL证书
func (db *SQLiteDB) UpdateCertificate(cert *model.Certificate) error {
    // 实现方法
}

// DeleteCertificate 删除SSL证书
func (db *SQLiteDB) DeleteCertificate(id int64) error {
    // 实现方法
}

// ListCertificates 列出所有证书
func (db *SQLiteDB) ListCertificates(page, pageSize int) ([]*model.Certificate, error) {
    // 实现方法
}

// GetExpiringCertificates 获取即将过期的证书
func (db *SQLiteDB) GetExpiringCertificates(within time.Duration) ([]*model.Certificate, error) {
    // 实现方法
}
```

### 6.4 SSL证书管理控制器

在`server/handlers/certificate.go`中实现:

```go
package handlers

import (
    "net/http"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    
    "v/model"
    "v/ssl"
)

// CertificateHandler 证书处理器
type CertificateHandler struct {
    db      model.DB
    manager *ssl.Manager
}

// NewCertificateHandler 创建证书处理器
func NewCertificateHandler(db model.DB, manager *ssl.Manager) *CertificateHandler {
    return &CertificateHandler{
        db:      db,
        manager: manager,
    }
}

// RegisterRoutes 注册路由
func (h *CertificateHandler) RegisterRoutes(router *gin.RouterGroup) {
    certs := router.Group("/certificates")
    {
        certs.GET("", h.ListCertificates)
        certs.GET("/:id", h.GetCertificate)
        certs.POST("", h.CreateCertificate)
        certs.PUT("/:id", h.UpdateCertificate)
        certs.DELETE("/:id", h.DeleteCertificate)
        certs.POST("/request", h.RequestCertificate)
        certs.POST("/:id/renew", h.RenewCertificate)
    }
}

// ListCertificates 列出所有证书
func (h *CertificateHandler) ListCertificates(c *gin.Context) {
    // 实现方法
}

// GetCertificate 获取证书详情
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
    // 实现方法
}

// CreateCertificate 创建证书
func (h *CertificateHandler) CreateCertificate(c *gin.Context) {
    // 实现方法
}

// UpdateCertificate 更新证书
func (h *CertificateHandler) UpdateCertificate(c *gin.Context) {
    // 实现方法
}

// DeleteCertificate 删除证书
func (h *CertificateHandler) DeleteCertificate(c *gin.Context) {
    // 实现方法
}

// RequestCertificate 请求新证书
func (h *CertificateHandler) RequestCertificate(c *gin.Context) {
    // 实现方法
}

// RenewCertificate 手动更新证书
func (h *CertificateHandler) RenewCertificate(c *gin.Context) {
    // 实现方法
}
```

### 6.5 前端界面实现

创建`web/src/views/Certificates.vue`:

```vue
<template>
  <div class="certificates-container">
    <div class="header">
      <h1>SSL证书管理</h1>
      <el-button type="primary" @click="showAddDialog">申请新证书</el-button>
    </div>
    
    <!-- 证书列表 -->
    <el-table :data="certificates" v-loading="loading" stripe>
      <el-table-column prop="domain" label="域名" width="200" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag :type="getCertStatusType(scope.row.status)">
            {{ scope.row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="120" />
      <el-table-column prop="expiresAt" label="过期时间" width="180">
        <template #default="scope">
          {{ formatDate(scope.row.expiresAt) }}
        </template>
      </el-table-column>
      <el-table-column prop="autoRenew" label="自动更新" width="100">
        <template #default="scope">
          <el-switch v-model="scope.row.autoRenew" @change="toggleAutoRenew(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button size="small" @click="viewDetails(scope.row)">查看</el-button>
          <el-button size="small" type="success" @click="renewCert(scope.row)">更新</el-button>
          <el-button size="small" type="danger" @click="deleteCert(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 添加证书对话框 -->
    <el-dialog v-model="dialogVisible" title="申请SSL证书" width="500px">
      <el-form :model="certForm" :rules="rules" ref="certFormRef" label-width="100px">
        <el-form-item label="域名" prop="domain">
          <el-input v-model="certForm.domain" placeholder="请输入域名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="certForm.email" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="自动更新" prop="autoRenew">
          <el-switch v-model="certForm.autoRenew" />
        </el-form-item>
        <el-form-item label="证书类型" prop="type">
          <el-radio-group v-model="certForm.type">
            <el-radio label="letsencrypt">Let's Encrypt</el-radio>
            <el-radio label="custom">自定义证书</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <!-- 自定义证书上传 -->
        <template v-if="certForm.type === 'custom'">
          <el-form-item label="证书文件" prop="certFile">
            <el-upload action="#" :auto-upload="false" :limit="1" ref="certFileRef">
              <el-button>选择证书文件</el-button>
            </el-upload>
          </el-form-item>
          <el-form-item label="密钥文件" prop="keyFile">
            <el-upload action="#" :auto-upload="false" :limit="1" ref="keyFileRef">
              <el-button>选择密钥文件</el-button>
            </el-upload>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCertForm">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 证书详情对话框 -->
    <el-dialog v-model="detailsVisible" title="证书详情" width="600px">
      <div v-if="selectedCert">
        <div class="cert-detail-item">
          <span class="label">域名:</span>
          <span>{{ selectedCert.domain }}</span>
        </div>
        <div class="cert-detail-item">
          <span class="label">状态:</span>
          <el-tag :type="getCertStatusType(selectedCert.status)">
            {{ selectedCert.status }}
          </el-tag>
        </div>
        <div class="cert-detail-item">
          <span class="label">类型:</span>
          <span>{{ selectedCert.type }}</span>
        </div>
        <div class="cert-detail-item">
          <span class="label">证书文件:</span>
          <span>{{ selectedCert.certFile }}</span>
        </div>
        <div class="cert-detail-item">
          <span class="label">密钥文件:</span>
          <span>{{ selectedCert.keyFile }}</span>
        </div>
        <div class="cert-detail-item">
          <span class="label">过期时间:</span>
          <span>{{ formatDate(selectedCert.expiresAt) }}</span>
        </div>
        <div class="cert-detail-item">
          <span class="label">上次检查时间:</span>
          <span>{{ formatDate(selectedCert.lastCheckedAt) }}</span>
        </div>
        <div class="cert-detail-item">
          <span class="label">上次更新时间:</span>
          <span>{{ formatDate(selectedCert.lastRenewedAt) }}</span>
        </div>
        <div class="cert-detail-item">
          <span class="label">自动更新:</span>
          <el-switch v-model="selectedCert.autoRenew" disabled />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
// 实现JavaScript逻辑
</script>

<style scoped>
/* 实现样式 */
</style>
```

### 6.6 服务配置与整合

修改`main.go`，集成SSL证书管理服务:

```go
// 初始化SSL证书管理器
certStorage := filepath.Join(config.DataDir, "certs")
sslManager := ssl.New(logger, db, notifier, certStorage)

// 启动SSL证书管理服务
sslManager.Start()
defer sslManager.Stop()

// 注册SSL证书管理API
certHandler := handlers.NewCertificateHandler(db, sslManager)
certHandler.RegisterRoutes(apiGroup)
```

## 7. 测试计划

1. 单元测试
   - 测试证书申请逻辑
   - 测试证书更新逻辑
   - 测试证书状态检查

2. 集成测试
   - 测试HTTP验证服务器
   - 测试与Let's Encrypt服务的交互
   - 测试自定义证书上传和加载

3. 端到端测试
   - 测试Web界面证书管理功能
   - 测试自动更新证书流程

## 8. 部署注意事项

1. 确保应用有权限在证书存储目录读写文件
2. 对于HTTP验证，确保:
   - 端口80可访问
   - 域名已正确解析到服务器IP
3. 在生产环境中使用Let's Encrypt正式环境，在开发环境中使用测试环境
4. 为避免触发Let's Encrypt的速率限制，添加适当的冷却期和重试机制

## 9. 时间估计

- 基础框架搭建: 1天
- 证书申请和更新功能: 2天
- 数据库接口实现: 1天
- 前端界面开发: 2天
- 测试与调试: 2天
- 集成与部署: 1天

总计: 约9个工作日

## 10. 优先级和里程碑

1. 优先级1: 基础SSL证书管理功能 (自动申请和更新)
2. 优先级2: 证书状态监控和告警
3. 优先级3: 自定义证书管理
4. 优先级4: 前端界面完善和用户体验优化

## 11. 风险评估

1. Let's Encrypt API变更风险
   - 缓解: 使用稳定的客户端库，保持更新
2. 证书验证失败风险
   - 缓解: 实现完善的错误处理和重试机制
3. 证书存储安全风险
   - 缓解: 实现适当的权限控制，限制证书文件的访问 