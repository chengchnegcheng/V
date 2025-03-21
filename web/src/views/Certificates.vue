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
            <el-radio label="dns">DNS 验证</el-radio>
            <el-radio label="http">HTTP 验证</el-radio>
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
    { required: true, message: '请选择证书文件', trigger: 'change' }
  ],
  keyFile: [
    { required: true, message: '请选择私钥文件', trigger: 'change' }
  ]
}

// 验证结果
const validateDialogVisible = ref(false)
const validateResult = ref(null)

// 获取证书列表
const getCertificates = async () => {
  loading.value = true
  try {
    certificates.value = await certificatesApi.getCertificates()
  } catch (error) {
    ElMessage.error('获取证书列表失败')
  } finally {
    loading.value = false
  }
}

// 申请证书
const handleApply = () => {
  applyDialogVisible.value = true
}

// 确认申请
const confirmApply = async () => {
  try {
    await certificatesApi.applyCertificate(applyForm.value)
    ElMessage.success('证书申请成功')
    applyDialogVisible.value = false
    getCertificates()
  } catch (error) {
    ElMessage.error('证书申请失败')
  }
}

// 上传证书
const handleUpload = () => {
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

// 确认上传
const confirmUpload = async () => {
  try {
    await certificatesApi.uploadCertificate(uploadForm.value)
    ElMessage.success('证书上传成功')
    uploadDialogVisible.value = false
    getCertificates()
  } catch (error) {
    ElMessage.error('证书上传失败')
  }
}

// 续期证书
const handleRenew = async (cert) => {
  try {
    await certificatesApi.renewCertificate(cert.id)
    ElMessage.success('证书续期成功')
    getCertificates()
  } catch (error) {
    ElMessage.error('证书续期失败')
  }
}

// 验证证书
const handleValidate = async (cert) => {
  try {
    validateResult.value = await certificatesApi.validateCertificate(cert.id)
    validateDialogVisible.value = true
  } catch (error) {
    ElMessage.error('证书验证失败')
  }
}

// 备份证书
const handleBackup = async (cert) => {
  try {
    await certificatesApi.backupCertificate(cert.id)
    ElMessage.success('证书备份成功')
  } catch (error) {
    ElMessage.error('证书备份失败')
  }
}

// 删除证书
const handleDelete = async (cert) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除证书 ${cert.domain} 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await certificatesApi.deleteCertificate(cert.id)
    ElMessage.success('证书删除成功')
    getCertificates()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('证书删除失败')
    }
  }
}

// 修改自动续期状态
const handleAutoRenewChange = async (cert) => {
  try {
    await certificatesApi.updateAutoRenew(cert.id, cert.autoRenew)
    ElMessage.success('自动续期设置已更新')
  } catch (error) {
    ElMessage.error('更新自动续期设置失败')
    cert.autoRenew = !cert.autoRenew // 恢复状态
  }
}

// DNS 记录管理
const addDnsRecord = () => {
  applyForm.value.dnsRecords.push({
    name: '',
    type: '',
    value: ''
  })
}

const removeDnsRecord = (index) => {
  applyForm.value.dnsRecords.splice(index, 1)
}

// 工具函数
const getProviderType = (provider) => {
  const types = {
    'letsencrypt': 'success',
    'zerossl': 'warning',
    'custom': 'info'
  }
  return types[provider] || 'info'
}

const getExpireStatusType = (expireDate) => {
  const now = new Date()
  const expire = new Date(expireDate)
  const days = Math.floor((expire - now) / (1000 * 60 * 60 * 24))
  
  if (days <= 0) return 'danger'
  if (days <= 30) return 'warning'
  return 'success'
}

// 刷新
const handleRefresh = () => {
  getCertificates()
}

onMounted(() => {
  getCertificates()
})
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