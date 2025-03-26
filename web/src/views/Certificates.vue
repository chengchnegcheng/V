<template>
  <div class="certificates-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>SSL 证书管理</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleApply">申请证书</el-button>
            <el-button type="success" @click="handleUpload">上传证书</el-button>
            <el-button type="info" @click="handleRefresh">刷新</el-button>
          </div>
        </div>
      </template>

      <!-- 证书列表 -->
      <el-table
        :data="certificates"
        border
        style="width: 100%"
        v-loading="loading"
      >
        <el-table-column prop="domain" label="域名" min-width="150" />
        <el-table-column prop="provider" label="提供商" width="120">
          <template #default="scope">
            <el-tag :type="getProviderType(scope.row.provider)">
              {{ scope.row.provider }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="issueDate" label="签发日期" width="120" />
        <el-table-column prop="expireDate" label="过期日期" width="120">
          <template #default="scope">
            <el-tag :type="getExpireStatusType(scope.row.expireDate)">
              {{ scope.row.expireDate }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="autoRenew" label="自动续期" width="100">
          <template #default="scope">
            <el-switch
              v-model="scope.row.autoRenew"
              @change="handleAutoRenewChange(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="scope">
            <el-button
              type="primary"
              size="small"
              @click="handleRenew(scope.row)"
            >
              续期
            </el-button>
            <el-button
              type="success"
              size="small"
              @click="handleValidate(scope.row)"
            >
              验证
            </el-button>
            <el-button
              type="warning"
              size="small"
              @click="handleBackup(scope.row)"
            >
              备份
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 申请证书对话框 -->
    <el-dialog
      v-model="applyDialogVisible"
      title="申请证书"
      width="50%"
    >
      <el-form
        ref="applyFormRef"
        :model="applyForm"
        :rules="applyRules"
        label-width="100px"
      >
        <el-form-item label="域名" prop="domain">
          <el-input v-model="applyForm.domain" placeholder="请输入域名" />
        </el-form-item>
        <el-form-item label="提供商" prop="provider">
          <el-select v-model="applyForm.provider" placeholder="请选择提供商">
            <el-option label="Let's Encrypt" value="letsencrypt" />
            <el-option label="ZeroSSL" value="zerossl" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="验证方式" prop="validationMethod">
          <el-radio-group v-model="applyForm.validationMethod">
            <el-radio value="dns">DNS 验证</el-radio>
            <el-radio value="http">HTTP 验证</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          v-if="applyForm.validationMethod === 'dns'"
          label="DNS 记录"
        >
          <div v-for="(record, index) in applyForm.dnsRecords" :key="index" class="dns-record">
            <el-input v-model="record.name" placeholder="记录名" />
            <el-input v-model="record.type" placeholder="类型" />
            <el-input v-model="record.value" placeholder="值" />
            <el-button type="danger" @click="removeDnsRecord(index)">删除</el-button>
          </div>
          <el-button type="primary" @click="addDnsRecord">添加记录</el-button>
        </el-form-item>
        <el-form-item
          v-if="applyForm.validationMethod === 'http'"
          label="验证路径"
        >
          <el-input v-model="applyForm.validationPath" placeholder="验证路径" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="applyDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmApply">确认申请</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 上传证书对话框 -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传证书"
      width="50%"
    >
      <el-form
        ref="uploadFormRef"
        :model="uploadForm"
        :rules="uploadRules"
        label-width="100px"
      >
        <el-form-item label="域名" prop="domain">
          <el-input v-model="uploadForm.domain" placeholder="请输入域名" />
        </el-form-item>
        <el-form-item label="证书文件" prop="certFile">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleCertFileChange"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 .pem, .crt 格式的证书文件
              </div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="私钥文件" prop="keyFile">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleKeyFileChange"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 .key 格式的私钥文件
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="uploadDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmUpload">确认上传</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 验证结果对话框 -->
    <el-dialog
      v-model="validateDialogVisible"
      title="证书验证"
      width="50%"
    >
      <div v-if="validateResult" class="validate-result">
        <div class="result-status">
          <el-tag :type="validateResult.success ? 'success' : 'danger'">
            {{ validateResult.success ? '验证成功' : '验证失败' }}
          </el-tag>
        </div>
        <div class="result-details">
          <div v-if="validateResult.message" class="detail-item">
            <span class="label">消息：</span>
            <span>{{ validateResult.message }}</span>
          </div>
          <div v-if="validateResult.details" class="detail-item">
            <span class="label">详情：</span>
            <pre class="details-content">{{ validateResult.details }}</pre>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { certificatesApi } from '@/api'

// 证书列表
const certificates = ref([])
const loading = ref(false)

// 申请证书
const applyDialogVisible = ref(false)
const applyFormRef = ref(null)
const applyForm = ref({
  domain: '',
  provider: 'letsencrypt',
  validationMethod: 'dns',
  dnsRecords: [],
  validationPath: ''
})
const applyRules = {
  domain: [
    { required: true, message: '请输入域名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/, message: '请输入有效的域名', trigger: 'blur' }
  ],
  provider: [
    { required: true, message: '请选择提供商', trigger: 'change' }
  ],
  validationMethod: [
    { required: true, message: '请选择验证方式', trigger: 'change' }
  ]
}

// 上传证书
const uploadDialogVisible = ref(false)
const uploadFormRef = ref(null)
const uploadForm = ref({
  domain: '',
  certFile: null,
  keyFile: null
})
const uploadRules = {
  domain: [
    { required: true, message: '请输入域名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/, message: '请输入有效的域名', trigger: 'blur' }
  ],
  certFile: [
    { required: true, message: '请上传证书文件', trigger: 'change' }
  ],
  keyFile: [
    { required: true, message: '请上传私钥文件', trigger: 'change' }
  ]
}

// 验证结果
const validateDialogVisible = ref(false)
const validateResult = ref(null)

// 生命周期钩子
onMounted(async () => {
  loading.value = true
  try {
    // 模拟获取证书数据
    // 实际环境应替换为 certificatesApi.getCertificates()
    setTimeout(() => {
      certificates.value = [
        {
          id: 1,
          domain: 'example.com',
          provider: 'letsencrypt',
          issueDate: '2023-06-01',
          expireDate: '2023-09-01',
          autoRenew: true
        },
        {
          id: 2,
          domain: 'test.example.com',
          provider: 'zerossl',
          issueDate: '2023-05-15',
          expireDate: '2023-08-15',
          autoRenew: false
        }
      ]
      loading.value = false
    }, 500)
  } catch (error) {
    ElMessage.error('获取证书列表失败')
    loading.value = false
  }
})

// 获取证书列表
const fetchCertificates = async () => {
  loading.value = true
  try {
    // 实际项目中应使用API获取数据
    // const response = await certificatesApi.getCertificates()
    // certificates.value = response
    
    // 模拟数据
    setTimeout(() => {
      certificates.value = [
        {
          id: 1,
          domain: 'example.com',
          provider: 'letsencrypt',
          issueDate: '2023-01-01',
          expireDate: '2023-04-01',
          autoRenew: true
        },
        {
          id: 2,
          domain: 'test.example.com',
          provider: 'zerossl',
          issueDate: '2023-01-15',
          expireDate: '2023-04-15',
          autoRenew: false
        },
        {
          id: 3,
          domain: 'api.example.com',
          provider: 'custom',
          issueDate: '2023-02-01',
          expireDate: '2024-02-01',
          autoRenew: true
        }
      ]
      loading.value = false
    }, 500)
  } catch (error) {
    console.error('Failed to fetch certificates:', error)
    ElMessage.error('获取证书列表失败')
    loading.value = false
  }
}

// 处理申请证书
const handleApply = () => {
  applyForm.value = {
    domain: '',
    provider: 'letsencrypt',
    validationMethod: 'dns',
    dnsRecords: [],
    validationPath: ''
  }
  applyDialogVisible.value = true
}

// 添加DNS记录
const addDnsRecord = () => {
  applyForm.value.dnsRecords.push({
    name: '',
    type: 'TXT',
    value: ''
  })
}

// 移除DNS记录
const removeDnsRecord = (index) => {
  applyForm.value.dnsRecords.splice(index, 1)
}

// 确认申请证书
const confirmApply = async () => {
  if (!applyFormRef.value) return
  
  try {
    await applyFormRef.value.validate()
    
    // 实际项目中应调用API申请证书
    // await certificatesApi.applyCertificate(applyForm.value)
    
    ElMessage.success('证书申请已提交，请等待处理结果')
    applyDialogVisible.value = false
    
    // 重新获取证书列表
    fetchCertificates()
  } catch (error) {
    console.error('Failed to apply certificate:', error)
    ElMessage.error('申请证书失败')
  }
}

// 处理上传证书
const handleUpload = () => {
  uploadForm.value = {
    domain: '',
    certFile: null,
    keyFile: null
  }
  uploadDialogVisible.value = true
}

// 处理证书文件选择
const handleCertFileChange = (file) => {
  uploadForm.value.certFile = file.raw
}

// 处理私钥文件选择
const handleKeyFileChange = (file) => {
  uploadForm.value.keyFile = file.raw
}

// 确认上传证书
const confirmUpload = async () => {
  if (!uploadFormRef.value) return
  
  try {
    await uploadFormRef.value.validate()
    
    // 实际项目中应调用API上传证书
    // await certificatesApi.uploadCertificate(uploadForm.value)
    
    ElMessage.success('证书上传成功')
    uploadDialogVisible.value = false
    
    // 重新获取证书列表
    fetchCertificates()
  } catch (error) {
    console.error('Failed to upload certificate:', error)
    ElMessage.error('上传证书失败')
  }
}

// 处理自动续期设置变更
const handleAutoRenewChange = async (row) => {
  try {
    // 实际项目中应调用API更新自动续期设置
    // await certificatesApi.updateAutoRenew(row.id, row.autoRenew)
    
    ElMessage.success(`${row.autoRenew ? '已开启' : '已关闭'}自动续期`)
  } catch (error) {
    console.error('Failed to update auto renew setting:', error)
    ElMessage.error('更新自动续期设置失败')
    // 恢复原状态
    row.autoRenew = !row.autoRenew
  }
}

// 处理续期证书
const handleRenew = async (row) => {
  try {
    ElMessageBox.confirm(`确定要为域名 ${row.domain} 续期证书吗？`, '续期证书', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      // 实际项目中应调用API续期证书
      // await certificatesApi.renewCertificate(row.id)
      
      ElMessage.success('证书续期已提交，请等待处理结果')
      
      // 重新获取证书列表
      fetchCertificates()
    })
  } catch (error) {
    console.error('Failed to renew certificate:', error)
    ElMessage.error('续期证书失败')
  }
}

// 处理验证证书
const handleValidate = async (row) => {
  try {
    // 实际项目中应调用API验证证书
    // const result = await certificatesApi.validateCertificate(row.id)
    
    // 模拟响应
    const result = {
      success: true,
      message: '证书有效',
      details: `域名: ${row.domain}\n提供商: ${row.provider}\n签发日期: ${row.issueDate}\n过期日期: ${row.expireDate}`
    }
    
    validateResult.value = result
    validateDialogVisible.value = true
  } catch (error) {
    console.error('Failed to validate certificate:', error)
    ElMessage.error('验证证书失败')
  }
}

// 处理备份证书
const handleBackup = async (row) => {
  try {
    // 实际项目中应调用API备份证书
    // await certificatesApi.backupCertificate(row.id)
    ElMessage.success('证书备份已下载')
  } catch (error) {
    console.error('Failed to backup certificate:', error)
    ElMessage.error('备份证书失败')
  }
}

// 处理删除证书
const handleDelete = async (row) => {
  try {
    ElMessageBox.confirm(`确定要删除域名 ${row.domain} 的证书吗？此操作不可恢复！`, '删除证书', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      // 实际项目中应调用API删除证书
      // await certificatesApi.deleteCertificate(row.id)
      
      ElMessage.success('证书已删除')
      
      // 重新获取证书列表
      fetchCertificates()
    })
  } catch (error) {
    console.error('Failed to delete certificate:', error)
    ElMessage.error('删除证书失败')
  }
}

// 处理刷新
const handleRefresh = () => {
  fetchCertificates()
}

// 获取提供商类型
const getProviderType = (provider) => {
  const types = {
    'letsencrypt': 'success',
    'zerossl': 'primary',
    'custom': 'warning'
  }
  return types[provider] || 'info'
}

// 获取过期状态类型
const getExpireStatusType = (expireDate) => {
  const now = new Date()
  const expire = new Date(expireDate)
  const diff = expire - now
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days < 0) {
    return 'danger'
  } else if (days < 30) {
    return 'warning'
  } else {
    return 'success'
  }
}
</script>

<style scoped>
.certificates-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.dns-record {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.validate-result {
  padding: 20px;
}

.result-status {
  margin-bottom: 20px;
}

.result-details {
  background-color: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
}

.detail-item {
  margin-bottom: 10px;
}

.detail-item .label {
  font-weight: bold;
  margin-right: 10px;
}

.details-content {
  margin: 10px 0;
  padding: 10px;
  background-color: #fff;
  border-radius: 4px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style> 