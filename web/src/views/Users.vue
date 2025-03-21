<template>
  <div class="users-container">
    <div class="header">
      <h1>用户管理</h1>
      <el-button type="primary" @click="showAddUserDialog">添加用户</el-button>
    </div>

    <!-- 用户列表 -->
    <el-table :data="users" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" width="150" />
      <el-table-column prop="email" label="邮箱" width="200" />
      <el-table-column label="流量使用" width="200">
        <template #default="scope">
          <el-progress 
            :percentage="calculateTrafficPercentage(scope.row)" 
            :status="getTrafficStatus(scope.row)"
          />
          <div>{{ formatTraffic(scope.row.trafficUsed) }} / {{ formatTraffic(scope.row.trafficLimit) }}</div>
        </template>
      </el-table-column>
      <el-table-column label="有效期" width="180">
        <template #default="scope">
          <span v-if="scope.row.expiryDate">{{ formatDate(scope.row.expiryDate) }}</span>
          <span v-else>永久有效</span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag :type="getStatusType(scope.row.status)">
            {{ getStatusText(scope.row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="220">
        <template #default="scope">
          <el-button size="small" @click="showEditUserDialog(scope.row)">编辑</el-button>
          <el-button size="small" type="success" @click="resetTraffic(scope.row)">重置流量</el-button>
          <el-button 
            size="small" 
            :type="scope.row.status === 'active' ? 'danger' : 'success'"
            @click="toggleUserStatus(scope.row)"
          >
            {{ scope.row.status === 'active' ? '禁用' : '启用' }}
          </el-button>
          <el-button size="small" type="danger" @click="confirmDeleteUser(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加/编辑用户对话框 -->
    <el-dialog 
      :title="isEdit ? '编辑用户' : '添加用户'" 
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="userForm" label-width="100px" :rules="rules" ref="userFormRef">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="userForm.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="流量限制" prop="trafficLimit">
          <el-input-number v-model="userForm.trafficLimit" :min="1" :step="1" />
          <el-select v-model="trafficUnit" style="margin-left: 10px">
            <el-option label="MB" value="MB" />
            <el-option label="GB" value="GB" />
            <el-option label="TB" value="TB" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效期" prop="expiryDate">
          <el-date-picker
            v-model="userForm.expiryDate"
            type="date"
            placeholder="选择日期"
            :disabled-date="disabledDate"
          />
        </el-form-item>
        <el-form-item label="管理员" prop="isAdmin">
          <el-switch v-model="userForm.isAdmin" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitUserForm">确认</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 确认删除对话框 -->
    <el-dialog
      title="确认删除"
      v-model="deleteDialogVisible"
      width="400px"
    >
      <p>确定要删除用户 "{{ userToDelete?.username }}" 吗？此操作不可恢复。</p>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="deleteDialogVisible = false">取消</el-button>
          <el-button type="danger" @click="deleteUser">确认删除</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'Users',
  setup() {
    // 状态
    const loading = ref(false)
    const users = ref([])
    const dialogVisible = ref(false)
    const deleteDialogVisible = ref(false)
    const isEdit = ref(false)
    const userToDelete = ref(null)
    const userFormRef = ref(null)
    const trafficUnit = ref('GB')

    // 表单数据
    const userForm = reactive({
      id: null,
      username: '',
      email: '',
      password: '',
      trafficLimit: 10,
      expiryDate: null,
      isAdmin: false
    })

    // 表单验证规则
    const rules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' },
        { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
      ],
      email: [
        { required: true, message: '请输入邮箱地址', trigger: 'blur' },
        { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度至少为 6 个字符', trigger: 'blur' }
      ]
    }

    // 获取用户列表
    const fetchUsers = async () => {
      loading.value = true
      try {
        // TODO: 替换为实际的 API 调用
        // const response = await userService.getUsers()
        // users.value = response.data
        
        // 模拟数据
        users.value = [
          {
            id: 1,
            username: 'admin',
            email: 'admin@example.com',
            trafficUsed: 2 * 1024 * 1024 * 1024, // 2GB
            trafficLimit: 10 * 1024 * 1024 * 1024, // 10GB
            expiryDate: new Date('2023-12-31'),
            status: 'active',
            isAdmin: true
          },
          {
            id: 2,
            username: 'user1',
            email: 'user1@example.com',
            trafficUsed: 8 * 1024 * 1024 * 1024, // 8GB
            trafficLimit: 10 * 1024 * 1024 * 1024, // 10GB
            expiryDate: new Date('2023-11-15'),
            status: 'active',
            isAdmin: false
          },
          {
            id: 3,
            username: 'user2',
            email: 'user2@example.com',
            trafficUsed: 12 * 1024 * 1024 * 1024, // 12GB
            trafficLimit: 10 * 1024 * 1024 * 1024, // 10GB
            expiryDate: null,
            status: 'disabled',
            isAdmin: false
          }
        ]
      } catch (error) {
        ElMessage.error('获取用户列表失败')
        console.error(error)
      } finally {
        loading.value = false
      }
    }

    // 添加用户对话框
    const showAddUserDialog = () => {
      isEdit.value = false
      resetUserForm()
      dialogVisible.value = true
    }

    // 编辑用户对话框
    const showEditUserDialog = (user) => {
      isEdit.value = true
      Object.assign(userForm, {
        id: user.id,
        username: user.username,
        email: user.email,
        trafficLimit: convertTrafficToUnit(user.trafficLimit),
        expiryDate: user.expiryDate,
        isAdmin: user.isAdmin
      })
      dialogVisible.value = true
    }

    // 重置表单
    const resetUserForm = () => {
      Object.assign(userForm, {
        id: null,
        username: '',
        email: '',
        password: '',
        trafficLimit: 10,
        expiryDate: null,
        isAdmin: false
      })
      trafficUnit.value = 'GB'
      if (userFormRef.value) {
        userFormRef.value.resetFields()
      }
    }

    // 提交表单
    const submitUserForm = async () => {
      if (!userFormRef.value) return
      
      await userFormRef.value.validate(async (valid) => {
        if (valid) {
          const userData = { ...userForm }
          // 转换流量单位
          userData.trafficLimit = convertUnitToTraffic(userData.trafficLimit, trafficUnit.value)
          
          try {
            if (isEdit.value) {
              // TODO: 调用编辑用户 API
              // await userService.updateUser(userData.id, userData)
              // 模拟更新
              const index = users.value.findIndex(u => u.id === userData.id)
              if (index !== -1) {
                users.value[index] = { ...users.value[index], ...userData }
              }
              ElMessage.success('用户更新成功')
            } else {
              // TODO: 调用添加用户 API
              // const response = await userService.createUser(userData)
              // 模拟添加
              const newUser = {
                id: users.value.length + 1,
                ...userData,
                trafficUsed: 0,
                status: 'active'
              }
              users.value.push(newUser)
              ElMessage.success('用户添加成功')
            }
            dialogVisible.value = false
          } catch (error) {
            ElMessage.error(isEdit.value ? '更新用户失败' : '添加用户失败')
            console.error(error)
          }
        }
      })
    }

    // 确认删除用户
    const confirmDeleteUser = (user) => {
      userToDelete.value = user
      deleteDialogVisible.value = true
    }

    // 删除用户
    const deleteUser = async () => {
      if (!userToDelete.value) return
      
      try {
        // TODO: 调用删除用户 API
        // await userService.deleteUser(userToDelete.value.id)
        // 模拟删除
        users.value = users.value.filter(u => u.id !== userToDelete.value.id)
        ElMessage.success('用户删除成功')
        deleteDialogVisible.value = false
      } catch (error) {
        ElMessage.error('删除用户失败')
        console.error(error)
      }
    }

    // 重置用户流量
    const resetTraffic = async (user) => {
      try {
        // TODO: 调用重置流量 API
        // await userService.resetTraffic(user.id)
        // 模拟重置
        const index = users.value.findIndex(u => u.id === user.id)
        if (index !== -1) {
          users.value[index].trafficUsed = 0
        }
        ElMessage.success('流量重置成功')
      } catch (error) {
        ElMessage.error('流量重置失败')
        console.error(error)
      }
    }

    // 切换用户状态
    const toggleUserStatus = async (user) => {
      const newStatus = user.status === 'active' ? 'disabled' : 'active'
      try {
        // TODO: 调用更新用户状态 API
        // await userService.updateUserStatus(user.id, newStatus)
        // 模拟更新
        const index = users.value.findIndex(u => u.id === user.id)
        if (index !== -1) {
          users.value[index].status = newStatus
        }
        ElMessage.success(`用户${newStatus === 'active' ? '启用' : '禁用'}成功`)
      } catch (error) {
        ElMessage.error(`用户${newStatus === 'active' ? '启用' : '禁用'}失败`)
        console.error(error)
      }
    }

    // 格式化流量显示
    const formatTraffic = (bytes) => {
      if (bytes < 1024) {
        return bytes + ' B'
      } else if (bytes < 1024 * 1024) {
        return (bytes / 1024).toFixed(2) + ' KB'
      } else if (bytes < 1024 * 1024 * 1024) {
        return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
      } else if (bytes < 1024 * 1024 * 1024 * 1024) {
        return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
      } else {
        return (bytes / (1024 * 1024 * 1024 * 1024)).toFixed(2) + ' TB'
      }
    }

    // 计算流量使用百分比
    const calculateTrafficPercentage = (user) => {
      if (!user.trafficLimit) return 0
      return Math.min(100, Math.round((user.trafficUsed / user.trafficLimit) * 100))
    }

    // 获取流量状态
    const getTrafficStatus = (user) => {
      const percentage = calculateTrafficPercentage(user)
      if (percentage >= 90) return 'exception'
      if (percentage >= 70) return 'warning'
      return 'success'
    }

    // 格式化日期
    const formatDate = (date) => {
      if (!date) return ''
      const d = new Date(date)
      return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')}`
    }

    // 获取状态样式
    const getStatusType = (status) => {
      return status === 'active' ? 'success' : 'danger'
    }

    // 获取状态文本
    const getStatusText = (status) => {
      return status === 'active' ? '启用' : '禁用'
    }

    // 禁用日期
    const disabledDate = (time) => {
      return time.getTime() < Date.now() - 8.64e7 // 禁用今天之前的日期
    }

    // 流量单位转换
    const convertUnitToTraffic = (value, unit) => {
      const unitMap = {
        'MB': 1024 * 1024,
        'GB': 1024 * 1024 * 1024,
        'TB': 1024 * 1024 * 1024 * 1024
      }
      return value * unitMap[unit]
    }

    const convertTrafficToUnit = (bytes) => {
      const unitValue = trafficUnit.value
      const unitMap = {
        'MB': 1024 * 1024,
        'GB': 1024 * 1024 * 1024,
        'TB': 1024 * 1024 * 1024 * 1024
      }
      return Math.round(bytes / unitMap[unitValue])
    }

    // 生命周期
    onMounted(() => {
      fetchUsers()
    })

    return {
      users,
      loading,
      dialogVisible,
      deleteDialogVisible,
      isEdit,
      userToDelete,
      userForm,
      userFormRef,
      trafficUnit,
      rules,
      showAddUserDialog,
      showEditUserDialog,
      submitUserForm,
      confirmDeleteUser,
      deleteUser,
      resetTraffic,
      toggleUserStatus,
      formatTraffic,
      calculateTrafficPercentage,
      getTrafficStatus,
      formatDate,
      getStatusType,
      getStatusText,
      disabledDate
    }
  }
}
</script>

<style scoped>
.users-container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style> 