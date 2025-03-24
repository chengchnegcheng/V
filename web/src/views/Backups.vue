<template>
  <div class="backups-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>备份管理</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleCreateBackup">创建备份</el-button>
            <el-button type="success" @click="handleUploadBackup">上传备份</el-button>
            <el-button type="info" @click="handleConfig">配置</el-button>
          </div>
        </div>
      </template>

      <!-- 备份统计 -->
      <el-row :gutter="20" class="stats-row">
        <el-col :span="6">
          <el-card shadow="hover" class="stats-card">
            <div class="stats-content">
              <div class="stats-title">总备份数</div>
              <div class="stats-value">{{ stats.totalBackups }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover" class="stats-card">
            <div class="stats-content">
              <div class="stats-title">总存储空间</div>
              <div class="stats-value">{{ formatSize(stats.totalSize) }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover" class="stats-card">
            <div class="stats-content">
              <div class="stats-title">上次备份时间</div>
              <div class="stats-value">{{ stats.lastBackupTime || '无' }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover" class="stats-card">
            <div class="stats-content">
              <div class="stats-title">下次备份时间</div>
              <div class="stats-value">{{ stats.nextBackupTime || '无' }}</div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 备份列表 -->
      <el-card class="backup-list-card">
        <template #header>
          <div class="card-header">
            <span>备份列表</span>
            <div class="header-actions">
              <el-input
                v-model="searchQuery"
                placeholder="搜索备份"
                style="width: 200px"
                clearable
              />
              <el-button type="primary" @click="refreshBackups">刷新</el-button>
            </div>
          </div>
        </template>
        <el-table
          :data="filteredBackups"
          border
          style="width: 100%"
          v-loading="loading"
        >
          <el-table-column prop="name" label="名称" width="200" />
          <el-table-column prop="size" label="大小" width="120">
            <template #default="scope">
              {{ formatSize(scope.row.size) }}
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="300" fixed="right">
            <template #default="scope">
              <el-button
                type="primary"
                size="small"
                @click="handleDownload(scope.row)"
              >
                下载
              </el-button>
              <el-button
                type="success"
                size="small"
                @click="handleRestore(scope.row)"
              >
                恢复
              </el-button>
              <el-button
                type="warning"
                size="small"
                @click="handleVerify(scope.row)"
              >
                验证
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
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </el-card>

    <!-- 创建备份对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建备份"
      width="40%"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        label-width="100px"
      >
        <el-form-item label="备份名称">
          <el-input v-model="createForm.name" placeholder="请输入备份名称" />
        </el-form-item>
        <el-form-item label="备份内容">
          <el-checkbox-group v-model="createForm.content">
            <el-checkbox label="config">配置文件</el-checkbox>
            <el-checkbox label="database">数据库</el-checkbox>
            <el-checkbox label="certificates">SSL证书</el-checkbox>
            <el-checkbox label="logs">日志文件</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="加密备份">
          <el-switch v-model="createForm.encrypt" />
        </el-form-item>
        <el-form-item
          v-if="createForm.encrypt"
          label="加密密码"
          prop="password"
        >
          <el-input
            v-model="createForm.password"
            type="password"
            placeholder="请输入加密密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmCreate">创建</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 上传备份对话框 -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传备份"
      width="40%"
    >
      <el-upload
        class="upload-demo"
        drag
        action="/api/backups/upload"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        :before-upload="beforeUpload"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            支持 .zip, .tar, .gz 格式的备份文件
          </div>
        </template>
      </el-upload>
    </el-dialog>

    <!-- 恢复备份对话框 -->
    <el-dialog
      v-model="restoreDialogVisible"
      title="恢复备份"
      width="40%"
    >
      <el-form
        ref="restoreFormRef"
        :model="restoreForm"
        label-width="100px"
      >
        <el-form-item label="备份文件">
          <el-upload
            class="upload-demo"
            drag
            action="/api/backups/restore"
            :on-success="handleRestoreSuccess"
            :on-error="handleRestoreError"
            :before-upload="beforeRestoreUpload"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              将备份文件拖到此处，或<em>点击上传</em>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item label="恢复选项">
          <el-checkbox-group v-model="restoreForm.options">
            <el-checkbox label="config">配置文件</el-checkbox>
            <el-checkbox label="database">数据库</el-checkbox>
            <el-checkbox label="certificates">SSL证书</el-checkbox>
            <el-checkbox label="logs">日志文件</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="restoreDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmRestore">恢复</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 配置对话框 -->
    <el-dialog
      v-model="configDialogVisible"
      title="备份配置"
      width="50%"
    >
      <el-tabs v-model="activeConfigTab">
        <!-- 存储配置 -->
        <el-tab-pane label="存储配置" name="storage">
          <el-form
            ref="storageFormRef"
            :model="storageConfig"
            label-width="120px"
          >
            <el-form-item label="存储类型">
              <el-select v-model="storageConfig.type">
                <el-option label="本地存储" value="local" />
                <el-option label="FTP" value="ftp" />
                <el-option label="S3" value="s3" />
                <el-option label="SFTP" value="sftp" />
              </el-select>
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type !== 'local'"
              label="存储路径"
            >
              <el-input v-model="storageConfig.path" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'ftp'"
              label="FTP服务器"
            >
              <el-input v-model="storageConfig.host" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'ftp'"
              label="FTP端口"
            >
              <el-input-number v-model="storageConfig.port" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'ftp'"
              label="FTP用户名"
            >
              <el-input v-model="storageConfig.username" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'ftp'"
              label="FTP密码"
            >
              <el-input
                v-model="storageConfig.password"
                type="password"
                show-password
              />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 's3'"
              label="Access Key"
            >
              <el-input v-model="storageConfig.accessKey" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 's3'"
              label="Secret Key"
            >
              <el-input
                v-model="storageConfig.secretKey"
                type="password"
                show-password
              />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 's3'"
              label="Bucket"
            >
              <el-input v-model="storageConfig.bucket" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'sftp'"
              label="SFTP主机"
            >
              <el-input v-model="storageConfig.host" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'sftp'"
              label="SFTP端口"
            >
              <el-input-number v-model="storageConfig.port" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'sftp'"
              label="SFTP用户名"
            >
              <el-input v-model="storageConfig.username" />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'sftp'"
              label="SFTP密码"
            >
              <el-input
                v-model="storageConfig.password"
                type="password"
                show-password
              />
            </el-form-item>
            <el-form-item
              v-if="storageConfig.type === 'sftp'"
              label="私钥文件"
            >
              <el-upload
                class="upload-demo"
                action="/api/backups/upload-key"
                :on-success="handleKeyUploadSuccess"
                :on-error="handleKeyUploadError"
              >
                <el-button type="primary">上传私钥</el-button>
              </el-upload>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 计划配置 -->
        <el-tab-pane label="计划配置" name="schedule">
          <el-form
            ref="scheduleFormRef"
            :model="scheduleConfig"
            label-width="120px"
          >
            <el-form-item label="启用自动备份">
              <el-switch v-model="scheduleConfig.enabled" />
            </el-form-item>
            <el-form-item
              v-if="scheduleConfig.enabled"
              label="备份频率"
            >
              <el-select v-model="scheduleConfig.frequency">
                <el-option label="每天" value="daily" />
                <el-option label="每周" value="weekly" />
                <el-option label="每月" value="monthly" />
              </el-select>
            </el-form-item>
            <el-form-item
              v-if="scheduleConfig.enabled"
              label="备份时间"
            >
              <el-time-picker
                v-model="scheduleConfig.time"
                format="HH:mm"
                placeholder="选择时间"
              />
            </el-form-item>
            <el-form-item
              v-if="scheduleConfig.enabled"
              label="保留备份数"
            >
              <el-input-number
                v-model="scheduleConfig.retention"
                :min="1"
                :max="100"
              />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="configDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveConfig">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { backupsApi } from '@/api'

// 备份数据
const backups = ref([])
const loading = ref(false)
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 统计信息
const stats = ref({
  totalBackups: 0,
  totalSize: 0,
  lastBackupTime: null,
  nextBackupTime: null
})

// 创建备份
const createDialogVisible = ref(false)
const createForm = ref({
  name: '',
  content: ['config', 'database'],
  encrypt: false,
  password: ''
})

// 上传备份
const uploadDialogVisible = ref(false)

// 恢复备份
const restoreDialogVisible = ref(false)
const restoreForm = ref({
  options: ['config', 'database']
})

// 配置
const configDialogVisible = ref(false)
const activeConfigTab = ref('storage')
const storageConfig = ref({
  type: 'local',
  path: '',
  host: '',
  port: 21,
  username: '',
  password: '',
  accessKey: '',
  secretKey: '',
  bucket: ''
})
const scheduleConfig = ref({
  enabled: false,
  frequency: 'daily',
  time: new Date(2000, 1, 1, 0, 0),
  retention: 7
})

// 过滤后的备份列表
const filteredBackups = computed(() => {
  if (!searchQuery.value) return backups.value
  const query = searchQuery.value.toLowerCase()
  return backups.value.filter(backup =>
    backup.name.toLowerCase().includes(query)
  )
})

// 获取备份列表
const getBackups = async () => {
  loading.value = true
  try {
    // 使用模拟数据，因为后端API可能还没完全实现
    // const data = await backupsApi.list()
    // backups.value = data.backups
    // total.value = data.total
    
    // 模拟数据
    setTimeout(() => {
      const mockBackups = []
      for (let i = 1; i <= 10; i++) {
        mockBackups.push({
          id: i,
          name: `备份_${i}_${new Date().toISOString().split('T')[0]}`,
          size: Math.floor(Math.random() * 1024 * 1024 * 100), // 随机大小，最大100MB
          createdAt: new Date(Date.now() - i * 24 * 60 * 60 * 1000).toLocaleString(), // 每天一个备份
          status: ['completed', 'failed', 'in_progress', 'pending'][Math.floor(Math.random() * 4)] // 随机状态
        })
      }
      backups.value = mockBackups
      total.value = mockBackups.length
      loading.value = false
    }, 500)
  } catch (error) {
    console.error('获取备份列表失败:', error)
    ElMessage.error('获取备份列表失败')
    loading.value = false
  }
}

// 获取统计信息
const getStats = async () => {
  try {
    // const response = await backupsApi.getStats()
    // stats.value = response
    
    // 模拟数据
    setTimeout(() => {
      stats.value = {
        totalBackups: 10,
        totalSize: 1024 * 1024 * 256, // 256 MB
        lastBackupTime: new Date(Date.now() - 24 * 60 * 60 * 1000).toLocaleString(), // 昨天
        nextBackupTime: new Date(Date.now() + 24 * 60 * 60 * 1000).toLocaleString() // 明天
      }
    }, 500)
  } catch (error) {
    console.error('获取统计信息失败:', error)
    ElMessage.error('获取统计信息失败')
  }
}

// 获取存储配置
const getStorageConfig = async () => {
  try {
    // const config = await backupsApi.getStorageConfig()
    // storageConfig.value = config
    
    // 模拟数据
    setTimeout(() => {
      storageConfig.value = {
        type: 'local',
        path: '/backups',
        host: '',
        port: 21,
        username: '',
        password: '',
        accessKey: '',
        secretKey: '',
        bucket: ''
      }
    }, 300)
  } catch (error) {
    console.error('获取存储配置失败:', error)
    ElMessage.error('获取存储配置失败')
  }
}

// 获取计划配置
const getScheduleConfig = async () => {
  try {
    // const config = await backupsApi.getScheduleConfig()
    // scheduleConfig.value = config
    
    // 模拟数据
    setTimeout(() => {
      scheduleConfig.value = {
        enabled: true,
        frequency: 'daily',
        time: new Date(2000, 1, 1, 2, 0), // 凌晨2点
        retention: 7
      }
    }, 300)
  } catch (error) {
    console.error('获取计划配置失败:', error)
    ElMessage.error('获取计划配置失败')
  }
}

// 创建备份
const handleCreateBackup = () => {
  createDialogVisible.value = true
}

const confirmCreate = async () => {
  try {
    // await backupsApi.create(createForm.value)
    
    // 模拟创建成功
    setTimeout(() => {
      ElMessage.success('备份创建成功')
      createDialogVisible.value = false
      getBackups()
      getStats()
    }, 1000)
  } catch (error) {
    console.error('备份创建失败:', error)
    ElMessage.error('备份创建失败')
  }
}

// 上传备份
const handleUploadBackup = () => {
  uploadDialogVisible.value = true
}

const handleUploadSuccess = () => {
  ElMessage.success('备份上传成功')
  uploadDialogVisible.value = false
  getBackups()
  getStats()
}

const handleUploadError = () => {
  ElMessage.error('备份上传失败')
}

const beforeUpload = (file) => {
  const isValidType = ['application/zip', 'application/x-tar', 'application/gzip'].includes(file.type)
  if (!isValidType) {
    ElMessage.error('只能上传 ZIP/TAR/GZ 格式的文件')
    return false
  }
  return true
}

// 恢复备份
const handleRestore = (backup) => {
  restoreDialogVisible.value = true
}

const handleRestoreSuccess = () => {
  ElMessage.success('备份恢复成功')
  restoreDialogVisible.value = false
}

const handleRestoreError = () => {
  ElMessage.error('备份恢复失败')
}

const beforeRestoreUpload = (file) => {
  const isValidType = ['application/zip', 'application/x-tar', 'application/gzip'].includes(file.type)
  if (!isValidType) {
    ElMessage.error('只能上传 ZIP/TAR/GZ 格式的文件')
    return false
  }
  return true
}

// 下载备份
const handleDownload = async (backup) => {
  try {
    // await backupsApi.download(backup.id)
    
    // 模拟下载
    ElMessage.success('备份开始下载')
    // 模拟下载逻辑
    setTimeout(() => {
      console.log('模拟下载备份:', backup.name)
    }, 1000)
  } catch (error) {
    console.error('备份下载失败:', error)
    ElMessage.error('备份下载失败')
  }
}

// 验证备份
const handleVerify = async (backup) => {
  try {
    // const result = await backupsApi.verify(backup.id)
    
    // 模拟验证
    const valid = Math.random() > 0.2 // 80%概率验证成功
    setTimeout(() => {
      if (valid) {
        ElMessage.success('备份验证成功')
      } else {
        ElMessage.error('备份验证失败：文件可能已损坏')
      }
    }, 1000)
  } catch (error) {
    console.error('备份验证失败:', error)
    ElMessage.error('备份验证失败')
  }
}

// 删除备份
const handleDelete = async (backup) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除此备份吗？此操作不可恢复。',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // await backupsApi.delete(backup.id)
    
    // 模拟删除
    setTimeout(() => {
      ElMessage.success('备份删除成功')
      getBackups()
      getStats()
    }, 500)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('备份删除失败:', error)
      ElMessage.error('备份删除失败')
    }
  }
}

// 配置
const handleConfig = () => {
  configDialogVisible.value = true
  getStorageConfig()
  getScheduleConfig()
}

const saveConfig = async () => {
  try {
    // await backupsApi.updateStorageConfig(storageConfig.value)
    // await backupsApi.updateScheduleConfig(scheduleConfig.value)
    
    // 模拟保存配置
    setTimeout(() => {
      ElMessage.success('配置保存成功')
      configDialogVisible.value = false
    }, 1000)
  } catch (error) {
    console.error('配置保存失败:', error)
    ElMessage.error('配置保存失败')
  }
}

// 私钥上传
const handleKeyUploadSuccess = () => {
  ElMessage.success('私钥上传成功')
}

const handleKeyUploadError = () => {
  ElMessage.error('私钥上传失败')
}

// 刷新备份列表
const refreshBackups = () => {
  getBackups()
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  getBackups()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  getBackups()
}

// 工具函数
const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getStatusType = (status) => {
  const types = {
    'completed': 'success',
    'failed': 'danger',
    'in_progress': 'warning',
    'pending': 'info'
  }
  return types[status] || 'info'
}

onMounted(() => {
  getBackups()
  getStats()
})
</script>

<style scoped>
.backups-container {
  padding: 20px;
}

.stats-row {
  margin-bottom: 20px;
}

.stats-card {
  height: 100px;
}

.stats-content {
  text-align: center;
}

.stats-title {
  font-size: 14px;
  color: #909399;
  margin-bottom: 10px;
}

.stats-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.backup-list-card {
  margin-bottom: 20px;
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

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.upload-demo {
  text-align: center;
}

.el-upload__tip {
  margin-top: 10px;
  color: #909399;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

