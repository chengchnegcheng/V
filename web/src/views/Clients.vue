<template>
  <div class="clients-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>客户端管理</span>
          <el-button type="primary" @click="showAddDialog">添加客户端</el-button>
        </div>
      </template>
      
      <el-table 
        :data="clients" 
        border 
        style="width: 100%" 
        v-loading="loading"
      >
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="proxy" label="使用协议" width="120">
          <template #default="{ row }">
            <el-tag :type="getProtocolType(row.proxy?.type || '')">
              {{ row.proxy?.name || '未设置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="uuid" label="唯一标识" width="100">
          <template #default="{ row }">
            <el-tag type="info">{{ row.uuid ? row.uuid.substring(0, 8) + '...' : 'N/A' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="upload" label="上传流量" width="120">
          <template #default="{ row }">
            {{ formatTraffic(row.upload) }}
          </template>
        </el-table-column>
        <el-table-column prop="download" label="下载流量" width="120">
          <template #default="{ row }">
            {{ formatTraffic(row.download) }}
          </template>
        </el-table-column>
        <el-table-column prop="totalGB" label="流量限制" width="120">
          <template #default="{ row }">
            {{ row.totalGB ? `${row.totalGB} GB` : '不限制' }}
          </template>
        </el-table-column>
        <el-table-column prop="expiryTime" label="到期时间" width="160">
          <template #default="{ row }">
            {{ row.expiryTime ? formatDate(row.expiryTime) : '永不过期' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'">
              {{ row.enabled ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
              <el-button size="small" type="success" @click="handleGenerateConfig(row)">生成配置</el-button>
              <el-button 
                size="small" 
                :type="row.enabled ? 'warning' : 'success'"
                @click="handleToggleStatus(row)"
              >
                {{ row.enabled ? '停用' : '启用' }}
              </el-button>
              <el-button size="small" type="info" @click="handleResetTraffic(row)">重置流量</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination-container">
        <el-pagination
          :current-page="currentPage"
          :page-sizes="[10, 20, 50, 100]"
          :page-size="pageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
    
    <!-- 添加/编辑客户端对话框 -->
    <el-dialog 
      :title="dialogType === 'add' ? '添加客户端' : '编辑客户端'" 
      v-model="dialogVisible"
      width="60%"
    >
      <el-form 
        ref="clientFormRef"
        :model="clientForm"
        :rules="formRules"
        label-width="120px"
        label-position="left"
      >
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本设置" name="basic">
            <el-form-item label="客户端名称" prop="name">
              <el-input v-model="clientForm.name" placeholder="请输入客户端名称" />
            </el-form-item>
            <el-form-item label="关联协议" prop="proxyId">
              <el-select v-model="clientForm.proxyId" placeholder="请选择关联协议" style="width: 100%">
                <el-option 
                  v-for="proxy in proxyOptions" 
                  :key="proxy.id" 
                  :label="proxy.name" 
                  :value="proxy.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="唯一标识" prop="uuid">
              <el-input v-model="clientForm.uuid" placeholder="自动生成">
                <template #append>
                  <el-button @click="generateUUID">生成</el-button>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="启用状态" prop="enabled">
              <el-switch v-model="clientForm.enabled" />
            </el-form-item>
            <el-form-item label="备注信息" prop="remark">
              <el-input 
                type="textarea" 
                v-model="clientForm.remark" 
                placeholder="请输入备注信息"
                rows="3"
              />
            </el-form-item>
          </el-tab-pane>
          <el-tab-pane label="限制设置" name="limits">
            <el-form-item label="流量限制(GB)" prop="totalGB">
              <el-input-number 
                v-model="clientForm.totalGB" 
                :min="0" 
                :precision="1" 
                style="width: 100%"
                placeholder="0表示不限制"
              />
            </el-form-item>
            <el-form-item label="到期时间" prop="expiryTime">
              <el-date-picker
                v-model="clientForm.expiryTime"
                type="datetime"
                placeholder="选择日期时间（不选则永不过期）"
                style="width: 100%"
                value-format="YYYY-MM-DD HH:mm:ss"
              />
            </el-form-item>
            <el-form-item label="限速(Mbps)" prop="speed">
              <el-input-number 
                v-model="clientForm.speed" 
                :min="0" 
                :precision="0" 
                style="width: 100%"
                placeholder="0表示不限制"
              />
            </el-form-item>
            <el-form-item label="设备数限制" prop="deviceLimit">
              <el-input-number 
                v-model="clientForm.deviceLimit" 
                :min="0" 
                :precision="0" 
                style="width: 100%"
                placeholder="0表示不限制"
              />
            </el-form-item>
          </el-tab-pane>
          <el-tab-pane label="高级设置" name="advanced">
            <el-form-item label="IP限制模式" prop="ipRestrictionMode">
              <el-select v-model="clientForm.ipRestrictionMode" style="width: 100%">
                <el-option label="不限制" value="none" />
                <el-option label="白名单模式" value="whitelist" />
                <el-option label="黑名单模式" value="blacklist" />
              </el-select>
            </el-form-item>
            <el-form-item 
              v-if="clientForm.ipRestrictionMode !== 'none'" 
              :label="clientForm.ipRestrictionMode === 'whitelist' ? 'IP白名单' : 'IP黑名单'" 
              prop="ipList"
            >
              <el-input 
                type="textarea" 
                v-model="clientForm.ipList" 
                placeholder="每行一个IP地址或CIDR"
                rows="4"
              />
            </el-form-item>
            <el-form-item label="启用定时模式" prop="scheduleEnabled">
              <el-switch v-model="clientForm.scheduleEnabled" />
            </el-form-item>
            <el-form-item v-if="clientForm.scheduleEnabled" label="可用时间段" prop="scheduleRules">
              <div 
                v-for="(rule, index) in clientForm.scheduleRules" 
                :key="index" 
                class="schedule-rule"
              >
                <el-row :gutter="10">
                  <el-col :span="8">
                    <el-select v-model="rule.days" multiple placeholder="选择生效日" style="width: 100%">
                      <el-option label="周一" value="1" />
                      <el-option label="周二" value="2" />
                      <el-option label="周三" value="3" />
                      <el-option label="周四" value="4" />
                      <el-option label="周五" value="5" />
                      <el-option label="周六" value="6" />
                      <el-option label="周日" value="0" />
                    </el-select>
                  </el-col>
                  <el-col :span="7">
                    <el-time-picker 
                      v-model="rule.startTime" 
                      format="HH:mm"
                      placeholder="开始时间"
                      style="width: 100%"
                    />
                  </el-col>
                  <el-col :span="7">
                    <el-time-picker 
                      v-model="rule.endTime" 
                      format="HH:mm"
                      placeholder="结束时间"
                      style="width: 100%"
                    />
                  </el-col>
                  <el-col :span="2">
                    <el-button type="danger" @click="removeScheduleRule(index)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </el-col>
                </el-row>
              </div>
              <div class="add-schedule-rule">
                <el-button type="primary" @click="addScheduleRule">添加时间规则</el-button>
              </div>
            </el-form-item>
          </el-tab-pane>
        </el-tabs>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveClient">保存</el-button>
        </span>
      </template>
    </el-dialog>
    
    <!-- 生成配置对话框 -->
    <el-dialog 
      title="生成客户端配置" 
      v-model="configDialogVisible"
      width="500px"
    >
      <el-form v-if="selectedClient" label-width="100px">
        <el-form-item label="客户端">
          <span>{{ selectedClient.name }}</span>
        </el-form-item>
        <el-form-item label="协议">
          <span>{{ selectedClient.proxy?.name || '未设置' }}</span>
        </el-form-item>
        <el-form-item label="配置格式">
          <el-select v-model="configFormat" placeholder="请选择配置格式" style="width: 100%">
            <el-option label="V2rayN" value="v2rayn" />
            <el-option label="Clash" value="clash" />
            <el-option label="Shadowrocket" value="shadowrocket" />
            <el-option label="Quantumult X" value="quantumultx" />
            <el-option label="V2rayNG" value="v2rayng" />
            <el-option label="Surfboard" value="surfboard" />
            <el-option label="ClashMeta" value="clashmeta" />
          </el-select>
        </el-form-item>
        <el-form-item label="包含备注">
          <el-switch v-model="includeRemark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="configDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="generateConfig">生成配置</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { clientsApi } from '@/api'
import { proxies as proxiesApi } from '@/api'

// 数据
const loading = ref(false)
const clients = ref([])
const proxyOptions = ref([])
const dialogVisible = ref(false)
const configDialogVisible = ref(false)
const dialogType = ref('add')
const activeTab = ref('basic')
const clientFormRef = ref(null)
const selectedClient = ref(null)
const configFormat = ref('v2rayn')
const includeRemark = ref(true)

// 分页
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 表单校验规则
const formRules = {
  name: [
    { required: true, message: '请输入客户端名称', trigger: 'blur' }
  ],
  proxyId: [
    { required: true, message: '请选择关联协议', trigger: 'change' }
  ]
}

// 客户端表单
const clientForm = reactive({
  id: null,
  name: '',
  proxyId: '',
  uuid: '',
  enabled: true,
  remark: '',
  totalGB: 0,
  expiryTime: '',
  speed: 0,
  deviceLimit: 0,
  ipRestrictionMode: 'none',
  ipList: '',
  scheduleEnabled: false,
  scheduleRules: []
})

// 生命周期钩子
onMounted(() => {
  fetchClients()
  fetchProxies()
})

// 获取客户端列表
const fetchClients = async () => {
  loading.value = true
  try {
    // 实际项目中应使用API获取数据
    // const response = await clientsApi.list({
    //   page: currentPage.value,
    //   pageSize: pageSize.value
    // })
    // clients.value = response.data
    // total.value = response.total
    
    // 模拟数据
    setTimeout(() => {
      clients.value = [
        {
          id: 1,
          name: '测试客户端1',
          proxy: { id: 1, name: '默认VMess协议', type: 'vmess' },
          uuid: '23d44ff-1412-4e40-ab46-ac76d1ab11a6',
          upload: 1024 * 1024 * 1024 * 2.5, // 2.5 GB
          download: 1024 * 1024 * 1024 * 5.3, // 5.3 GB
          totalGB: 100,
          expiryTime: '2024-12-31 23:59:59',
          enabled: true,
          remark: '测试用户',
          speed: 10,
          deviceLimit: 3
        },
        {
          id: 2,
          name: '测试客户端2',
          proxy: { id: 2, name: '默认Trojan协议', type: 'trojan' },
          uuid: '35e55ff-2513-5e50-cd57-bd87e2cd22b7',
          upload: 1024 * 1024 * 1024 * 0.8, // 0.8 GB
          download: 1024 * 1024 * 1024 * 1.2, // 1.2 GB
          totalGB: 50,
          expiryTime: '2024-10-15 23:59:59',
          enabled: false,
          remark: '测试用户2',
          speed: 5,
          deviceLimit: 1
        }
      ]
      total.value = 2
      loading.value = false
    }, 500)
  } catch (error) {
    console.error('Failed to fetch clients:', error)
    ElMessage.error('获取客户端列表失败')
    loading.value = false
  }
}

// 获取协议列表
const fetchProxies = async () => {
  try {
    // 实际项目中应使用API获取数据
    // const response = await proxiesApi.list()
    // proxyOptions.value = response
    
    // 模拟数据
    proxyOptions.value = [
      { id: 1, name: '默认VMess协议', type: 'vmess' },
      { id: 2, name: '默认Trojan协议', type: 'trojan' },
      { id: 3, name: 'WebSocket协议', type: 'vmess' }
    ]
  } catch (error) {
    console.error('Failed to fetch proxies:', error)
    ElMessage.error('获取协议列表失败')
  }
}

// 显示添加对话框
const showAddDialog = () => {
  dialogType.value = 'add'
  resetForm()
  dialogVisible.value = true
  activeTab.value = 'basic'
}

// 重置表单
const resetForm = () => {
  Object.assign(clientForm, {
    id: null,
    name: '',
    proxyId: '',
    uuid: '',
    enabled: true,
    remark: '',
    totalGB: 0,
    expiryTime: '',
    speed: 0,
    deviceLimit: 0,
    ipRestrictionMode: 'none',
    ipList: '',
    scheduleEnabled: false,
    scheduleRules: []
  })
}

// 生成UUID
const generateUUID = () => {
  clientForm.uuid = crypto.randomUUID()
}

// 处理编辑
const handleEdit = (row) => {
  dialogType.value = 'edit'
  resetForm()
  // 深拷贝以避免直接修改列表数据
  const client = JSON.parse(JSON.stringify(row))
  Object.assign(clientForm, {
    id: client.id,
    name: client.name,
    proxyId: client.proxy?.id || '',
    uuid: client.uuid,
    enabled: client.enabled,
    remark: client.remark,
    totalGB: client.totalGB,
    expiryTime: client.expiryTime,
    speed: client.speed,
    deviceLimit: client.deviceLimit,
    ipRestrictionMode: client.ipRestrictionMode || 'none',
    ipList: client.ipList || '',
    scheduleEnabled: client.scheduleEnabled || false,
    scheduleRules: client.scheduleRules || []
  })
  dialogVisible.value = true
  activeTab.value = 'basic'
}

// 处理生成配置
const handleGenerateConfig = (row) => {
  selectedClient.value = row
  configFormat.value = 'v2rayn'
  includeRemark.value = true
  configDialogVisible.value = true
}

// 生成配置
const generateConfig = async () => {
  if (!selectedClient.value) return
  
  try {
    // 实际项目中应调用API
    // await clientsApi.generateConfig(selectedClient.value.id, configFormat.value)
    
    ElMessage.success('配置生成成功，已自动下载')
    configDialogVisible.value = false
  } catch (error) {
    console.error('Failed to generate config:', error)
    ElMessage.error('配置生成失败')
  }
}

// 处理启用/停用
const handleToggleStatus = async (row) => {
  try {
    // 实际项目中应调用API
    // if (row.enabled) {
    //   await clientsApi.disable(row.id)
    // } else {
    //   await clientsApi.enable(row.id)
    // }
    
    const action = row.enabled ? '停用' : '启用'
    ElMessage.success(`已${action}客户端：${row.name}`)
    row.enabled = !row.enabled
  } catch (error) {
    console.error('Failed to toggle client status:', error)
    ElMessage.error(`操作失败`)
  }
}

// 处理重置流量
const handleResetTraffic = (row) => {
  ElMessageBox.confirm(`确定要重置客户端 ${row.name} 的流量统计吗?`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      // 实际项目中应调用API
      // await clientsApi.resetTraffic(row.id)
      
      row.upload = 0
      row.download = 0
      ElMessage.success('流量重置成功')
    } catch (error) {
      console.error('Failed to reset traffic:', error)
      ElMessage.error('操作失败')
    }
  }).catch(() => {
    ElMessage.info('已取消操作')
  })
}

// 处理删除
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除客户端 ${row.name} 吗?`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      // 实际项目中应调用API
      // await clientsApi.delete(row.id)
      
      clients.value = clients.value.filter(c => c.id !== row.id)
      ElMessage.success('删除成功')
    } catch (error) {
      console.error('Failed to delete client:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {
    ElMessage.info('已取消删除')
  })
}

// 添加时间规则
const addScheduleRule = () => {
  clientForm.scheduleRules.push({
    days: [],
    startTime: '',
    endTime: ''
  })
}

// 删除时间规则
const removeScheduleRule = (index) => {
  clientForm.scheduleRules.splice(index, 1)
}

// 保存客户端
const handleSaveClient = async () => {
  if (!clientFormRef.value) return
  
  try {
    await clientFormRef.value.validate()
    
    if (dialogType.value === 'add') {
      // 添加新客户端
      // 实际项目中应调用API
      // await clientsApi.create(clientForm)
      
      const proxy = proxyOptions.value.find(p => p.id === clientForm.proxyId)
      const newClient = {
        ...JSON.parse(JSON.stringify(clientForm)),
        id: Date.now(),
        proxy: proxy,
        upload: 0,
        download: 0,
        created: new Date().toLocaleString()
      }
      clients.value.push(newClient)
      ElMessage.success('添加成功')
    } else {
      // 更新现有客户端
      // 实际项目中应调用API
      // await clientsApi.update(clientForm.id, clientForm)
      
      const index = clients.value.findIndex(c => c.id === clientForm.id)
      if (index !== -1) {
        const proxy = proxyOptions.value.find(p => p.id === clientForm.proxyId)
        const updatedClient = {
          ...JSON.parse(JSON.stringify(clientForm)),
          proxy: proxy
        }
        clients.value[index] = { ...clients.value[index], ...updatedClient }
        ElMessage.success('更新成功')
      }
    }
    dialogVisible.value = false
  } catch (error) {
    console.error('Form validation failed:', error)
    ElMessage.error('表单验证失败，请检查输入')
  }
}

// 分页大小变化
const handleSizeChange = (size) => {
  pageSize.value = size
  fetchClients()
}

// 分页页码变化
const handleCurrentChange = (page) => {
  currentPage.value = page
  fetchClients()
}

// 流量格式化
const formatTraffic = (bytes) => {
  if (!bytes) return '0 B'
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  
  return `${bytes.toFixed(2)} ${units[i]}`
}

// 日期格式化
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  
  const date = new Date(dateStr)
  return date.toLocaleString()
}

// 获取协议类型对应的样式
const getProtocolType = (type) => {
  const types = {
    'vmess': 'primary',
    'vless': 'success',
    'trojan': 'warning',
    'shadowsocks': 'danger',
    'socks': 'info',
    'http': 'default'
  }
  return types[type] || 'default'
}
</script>

<style scoped>
.clients-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.schedule-rule {
  margin-bottom: 10px;
  padding: 10px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background-color: #f9f9f9;
}

.add-schedule-rule {
  margin-top: 10px;
}
</style> 