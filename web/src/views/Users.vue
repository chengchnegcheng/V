<template>
  <div class="users-container">
    <h1>用户管理</h1>
    
    <div class="actions">
      <el-button type="primary" @click="showAddDialog">添加用户</el-button>
      <el-input 
        v-model="searchQuery" 
        placeholder="搜索用户" 
        clearable 
        style="width: 200px; margin-left: 10px;"
        @input="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>
    
    <el-table :data="filteredUsers" border v-loading="loading" style="width: 100%">
      <el-table-column prop="username" label="用户名" width="150"></el-table-column>
      <el-table-column prop="email" label="邮箱" width="200"></el-table-column>
      <el-table-column prop="role" label="角色" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.role === 'admin' ? 'danger' : 'primary'">
            {{ scope.row.role === 'admin' ? '管理员' : '普通用户' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created" label="创建时间" width="180"></el-table-column>
      <el-table-column prop="lastLogin" label="最后登录" width="180"></el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.status ? 'success' : 'danger'">
            {{ scope.row.status ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button-group>
            <el-button size="small" type="primary" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button 
              size="small" 
              :type="scope.row.status ? 'warning' : 'success'"
              @click="handleToggleStatus(scope.row)"
            >
              {{ scope.row.status ? '禁用' : '启用' }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 分页控件 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        :total="totalUsers"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
    
    <!-- 添加/编辑用户对话框 -->
    <el-dialog 
      :title="dialogType === 'add' ? '添加用户' : '编辑用户'" 
      v-model="dialogVisible"
      width="500px"
    >
      <el-form 
        :model="userForm" 
        :rules="rules" 
        ref="userFormRef" 
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名"></el-input>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱"></el-input>
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="dialogType === 'add'">
          <el-input v-model="userForm.password" type="password" placeholder="请输入密码"></el-input>
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="管理员" value="admin"></el-option>
            <el-option label="普通用户" value="user"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveUser">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

// Store
const userStore = useUserStore()

// 状态
const users = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogType = ref('add')
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const totalUsers = ref(0)
const userFormRef = ref(null)

// 表单
const userForm = reactive({
  id: null,
  username: '',
  email: '',
  password: '',
  role: 'user'
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
    { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 计算属性
const filteredUsers = computed(() => {
  if (!searchQuery.value) {
    return users.value
  }
  const query = searchQuery.value.toLowerCase()
  return users.value.filter(user => 
    user.username.toLowerCase().includes(query) || 
    user.email.toLowerCase().includes(query)
  )
})

// 生命周期钩子
onMounted(() => {
  fetchUsers()
})

// 方法
const fetchUsers = async () => {
  loading.value = true
  try {
    // 实际项目中应从API获取用户列表
    // 此处为模拟数据
    setTimeout(() => {
      users.value = [
        {
          id: 1,
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin',
          created: '2023-01-01 12:00:00',
          lastLogin: '2023-03-15 08:30:22',
          status: true
        },
        {
          id: 2,
          username: 'user1',
          email: 'user1@example.com',
          role: 'user',
          created: '2023-01-02 14:30:00',
          lastLogin: '2023-03-14 16:42:51',
          status: true
        },
        {
          id: 3,
          username: 'moderator',
          email: 'mod@example.com',
          role: 'user',
          created: '2023-01-03 10:15:30',
          lastLogin: '2023-03-10 11:20:15',
          status: true
        }
      ]
      totalUsers.value = users.value.length
      loading.value = false
    }, 500)
  } catch (error) {
    console.error('Failed to fetch users:', error)
    ElMessage.error('获取用户列表失败')
    loading.value = false
  }
}

const showAddDialog = () => {
  dialogType.value = 'add'
  Object.assign(userForm, {
    id: null,
    username: '',
    email: '',
    password: '',
    role: 'user'
  })
  dialogVisible.value = true
  // 在下一个tick重置表单验证
  setTimeout(() => {
    userFormRef.value?.resetFields()
  })
}

const handleEdit = (row) => {
  dialogType.value = 'edit'
  Object.assign(userForm, {
    id: row.id,
    username: row.username,
    email: row.email,
    role: row.role
  })
  dialogVisible.value = true
  // 在下一个tick重置表单验证
  setTimeout(() => {
    userFormRef.value?.resetFields()
  })
}

const handleToggleStatus = async (row) => {
  try {
    const action = row.status ? '禁用' : '启用'
    
    // 实际项目中应调用API更新用户状态
    // await userStore.updateUserStatus(row.id, !row.status)
    
    row.status = !row.status
    ElMessage.success(`已${action}用户：${row.username}`)
  } catch (error) {
    console.error('Failed to update user status:', error)
    ElMessage.error('更新用户状态失败')
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除用户 ${row.username} 吗?`,
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
  .then(async () => {
    try {
      // 实际项目中应调用API删除用户
      // await userStore.deleteUser(row.id)
      
      users.value = users.value.filter(u => u.id !== row.id)
      totalUsers.value = users.value.length
      ElMessage.success('删除成功')
    } catch (error) {
      console.error('Failed to delete user:', error)
      ElMessage.error('删除用户失败')
    }
  })
  .catch(() => {
    ElMessage.info('已取消删除')
  })
}

const handleSaveUser = async () => {
  if (!userFormRef.value) return
  
  userFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (dialogType.value === 'add') {
          // 添加新用户
          // 实际项目中应调用API创建用户
          // await userStore.createUser(userForm)
          
          const newUser = {
            ...userForm,
            id: Date.now(),
            created: new Date().toLocaleString(),
            lastLogin: '-',
            status: true
          }
          users.value.push(newUser)
          totalUsers.value = users.value.length
          ElMessage.success('添加成功')
        } else {
          // 更新现有用户
          // 实际项目中应调用API更新用户
          // await userStore.updateUser(userForm.id, userForm)
          
          const index = users.value.findIndex(u => u.id === userForm.id)
          if (index !== -1) {
            users.value[index] = { ...users.value[index], ...userForm }
            ElMessage.success('更新成功')
          }
        }
        dialogVisible.value = false
      } catch (error) {
        console.error('Failed to save user:', error)
        ElMessage.error('保存用户失败')
      }
    }
  })
}

const handleSearch = () => {
  currentPage.value = 1
}

const handleSizeChange = (val) => {
  pageSize.value = val
  currentPage.value = 1
}

const handleCurrentChange = (val) => {
  currentPage.value = val
}
</script>

<style scoped>
.users-container {
  padding: 20px;
}

.actions {
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style> 