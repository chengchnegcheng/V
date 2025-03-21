# V项目测试计划

## 1. 概述

本测试计划概述了V项目的测试策略、测试范围、测试环境和测试流程。V项目是一个多协议、多用户的代理面板系统，提供系统状态监控、流量统计、限制流量和到期时间等功能。

## 2. 测试目标

- 确保V项目符合功能需求规格说明
- 确保系统稳定可靠，能够长时间运行
- 验证系统的安全性，防止未授权访问
- 确保良好的用户体验和界面响应性能
- 测试系统在高负载下的性能表现

## 3. 测试范围

### 3.1 功能测试

#### 用户管理
- 用户注册、登录和登出
- 用户权限控制
- 用户流量限制设置
- 用户到期时间设置

#### 协议支持
- vmess协议配置和连接
- vless协议配置和连接
- trojan协议配置和连接
- shadowsocks协议配置和连接
- dokodemo-door协议配置和连接
- socks协议配置和连接
- http协议配置和连接

#### 系统监控
- CPU使用率监控
- 内存使用率监控
- 磁盘使用率监控
- 网络流量监控

#### 流量统计
- 用户级别流量统计
- 协议级别流量统计
- 流量历史记录查询
- 流量图表显示

#### SSL证书管理
- 自动申请证书
- 自动更新证书
- 手动导入证书
- 证书状态显示

### 3.2 性能测试

- 多用户并发连接测试
- 长时间运行稳定性测试
- 高流量传输测试
- 大量用户数据下的数据库性能测试

### 3.3 安全测试

- 身份验证与授权测试
- 数据传输加密测试
- 密码策略测试
- SQL注入防护测试
- XSS防护测试

### 3.4 兼容性测试

- 不同浏览器兼容性测试
- 不同操作系统下的服务端测试
- 不同客户端软件的兼容性测试

## 4. 测试环境

### 4.1 硬件环境

- **服务器**: 
  - CPU: 至少2核
  - 内存: 至少2GB
  - 磁盘: 至少20GB SSD
  - 网络: 千兆网卡

- **客户端测试机**:
  - 各种操作系统(Windows, macOS, Linux)
  - 不同配置的测试机器

### 4.2 软件环境

- **服务器操作系统**: Ubuntu 20.04 LTS / CentOS 8
- **Web服务器**: Nginx
- **数据库**: SQLite
- **客户端浏览器**: 
  - Chrome (最新版)
  - Firefox (最新版)
  - Safari (最新版)
  - Edge (最新版)
- **客户端软件**:
  - v2ray
  - Shadowsocks
  - Trojan

### 4.3 网络环境

- 局域网测试环境
- 公网测试环境
- 模拟不同网络条件(高延迟、丢包等)

## 5. 测试方法

### 5.1 单元测试

使用Go语言的标准测试框架，对关键模块进行单元测试:

```go
func TestUserLogin(t *testing.T) {
    // 测试用户登录逻辑
}

func TestTrafficCalculation(t *testing.T) {
    // 测试流量计算逻辑
}
```

### 5.2 集成测试

测试不同模块之间的交互:

- 数据库接口与业务逻辑的集成
- 前端与后端API的集成
- 流量统计与代理服务的集成

### 5.3 端到端测试

使用Selenium或Cypress等工具进行端到端测试，模拟真实用户操作:

```javascript
describe('User management', () => {
  it('should allow admin to create new user', () => {
    // 模拟创建用户的过程
  });
  
  it('should display correct traffic statistics', () => {
    // 模拟查看流量统计的过程
  });
});
```

### 5.4 性能测试

使用JMeter或Locust等工具进行性能测试:

- 测试API响应时间
- 测试系统在高并发下的表现
- 测试数据库在大量记录下的查询性能

### 5.5 安全测试

- 使用OWASP ZAP进行安全扫描
- 进行渗透测试，检测可能的安全漏洞
- 检查SSL/TLS配置的安全性

## 6. 测试用例

### 6.1 用户管理测试用例

| 用例ID | 用例描述 | 预期结果 | 优先级 |
|--------|---------|----------|--------|
| UM-001 | 管理员创建新用户 | 成功创建用户并能登录系统 | 高 |
| UM-002 | 用户登录验证 | 正确凭证可登录，错误凭证被拒绝 | 高 |
| UM-003 | 用户密码修改 | 密码成功修改且能用新密码登录 | 高 |
| UM-004 | 用户流量限制 | 达到流量限制后无法继续使用服务 | 高 |
| UM-005 | 用户到期管理 | 到期后账户自动停用 | 高 |

### 6.2 协议管理测试用例

| 用例ID | 用例描述 | 预期结果 | 优先级 |
|--------|---------|----------|--------|
| PM-001 | 添加VMess协议 | 成功添加协议并能正常连接 | 高 |
| PM-002 | 添加VLESS协议 | 成功添加协议并能正常连接 | 高 |
| PM-003 | 添加Trojan协议 | 成功添加协议并能正常连接 | 高 |
| PM-004 | 添加Shadowsocks协议 | 成功添加协议并能正常连接 | 高 |
| PM-005 | 修改协议配置 | 配置更改后生效 | 中 |
| PM-006 | 删除协议 | 协议被删除且不能再连接 | 中 |

### 6.3 流量统计测试用例

| 用例ID | 用例描述 | 预期结果 | 优先级 |
|--------|---------|----------|--------|
| TS-001 | 单用户流量统计 | 准确记录用户上传下载流量 | 高 |
| TS-002 | 多用户流量汇总 | 系统总流量等于所有用户流量之和 | 中 |
| TS-003 | 流量图表展示 | 图表正确展示流量趋势 | 中 |
| TS-004 | 流量重置功能 | 重置后流量计数归零 | 中 |
| TS-005 | 历史流量查询 | 能查询不同时间段的历史流量 | 低 |

### 6.4 系统监控测试用例

| 用例ID | 用例描述 | 预期结果 | 优先级 |
|--------|---------|----------|--------|
| SM-001 | CPU监控 | 准确显示CPU使用率 | 高 |
| SM-002 | 内存监控 | 准确显示内存使用情况 | 高 |
| SM-003 | 磁盘监控 | 准确显示磁盘使用情况 | 高 |
| SM-004 | 网络监控 | 准确显示网络吞吐量 | 高 |
| SM-005 | 警报功能 | 资源使用超过阈值时发出警报 | 中 |

### 6.5 SSL证书管理测试用例

| 用例ID | 用例描述 | 预期结果 | 优先级 |
|--------|---------|----------|--------|
| SSL-001 | 自动申请证书 | 成功从Let's Encrypt申请证书 | 高 |
| SSL-002 | 证书自动更新 | 证书到期前自动更新 | 高 |
| SSL-003 | 手动导入证书 | 成功导入并使用自定义证书 | 中 |
| SSL-004 | 查看证书状态 | 正确显示证书有效期和状态 | 中 |
| SSL-005 | 证书过期提醒 | 证书即将过期时发出提醒 | 中 |

## 7. 测试进度计划

| 阶段 | 开始日期 | 结束日期 | 测试内容 | 完成标准 |
|------|---------|----------|---------|----------|
| 准备阶段 | 第1天 | 第3天 | 搭建测试环境，准备测试数据 | 测试环境可用 |
| 单元测试 | 第4天 | 第7天 | 各模块单元测试 | 单元测试通过率>90% |
| 集成测试 | 第8天 | 第12天 | 模块间集成测试 | 主要功能正常工作 |
| 系统测试 | 第13天 | 第18天 | 全系统功能测试 | 所有关键功能正常 |
| 性能测试 | 第19天 | 第21天 | 负载测试，稳定性测试 | 满足性能指标要求 |
| 安全测试 | 第22天 | 第24天 | 安全漏洞扫描和渗透测试 | 无严重安全漏洞 |
| 回归测试 | 第25天 | 第27天 | 修复后的功能回归测试 | 修复的问题不再复现 |
| 验收测试 | 第28天 | 第30天 | 用户验收测试 | 产品满足用户需求 |

## 8. 缺陷管理

1. 缺陷严重程度分类:
   - 严重: 导致系统崩溃或无法使用核心功能
   - 高: 影响主要功能但有替代方案
   - 中: 影响次要功能
   - 低: 界面或文档问题

2. 缺陷报告格式:
   - 缺陷ID
   - 发现日期
   - 发现人
   - 影响模块
   - 严重程度
   - 复现步骤
   - 预期结果
   - 实际结果
   - 附件(截图、日志等)

3. 缺陷生命周期:
   - 新建 -> 分配 -> 修复中 -> 待验证 -> 关闭/重新打开

## 9. 测试交付物

1. 测试计划文档
2. 测试用例集
3. 测试执行报告
4. 缺陷报告
5. 性能测试报告
6. 安全测试报告
7. 最终测试总结报告

## 10. 风险与应对策略

| 风险 | 可能性 | 影响 | 应对策略 |
|------|-------|------|----------|
| 测试环境不稳定 | 中 | 高 | 准备备用测试环境，做好环境快照 |
| 测试时间不足 | 高 | 高 | 优先测试核心功能，安排合理的测试计划 |
| 测试数据不足 | 中 | 中 | 提前准备足够的测试数据，必要时编写数据生成脚本 |
| 需求变更导致测试返工 | 高 | 高 | 与开发团队保持沟通，灵活调整测试计划 |
| 发现严重缺陷延迟发布 | 中 | 高 | 提前进行关键功能的冒烟测试，尽早发现问题 |

## 11. 测试资源需求

1. 人员资源:
   - 测试工程师: 2名
   - 开发支持: 1名
   - 运维支持: 1名

2. 设备资源:
   - 测试服务器: 2台
   - 各类客户端设备: 5台以上
   - 各类网络环境: 至少包含局域网和公网环境

3. 工具资源:
   - 缺陷管理工具: Jira
   - 自动化测试工具: Selenium/Cypress
   - 性能测试工具: JMeter/Locust
   - 安全测试工具: OWASP ZAP

## 12. 附录

### 12.1 测试模板

- 测试用例模板
- 缺陷报告模板
- 测试执行报告模板

### 12.2 环境配置指南

- 测试服务器配置指南
- 客户端环境配置指南
- 测试工具配置指南 