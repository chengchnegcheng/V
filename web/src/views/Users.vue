<template>
  <div class="users-container">
    <h1>用户管理</h1>
    
    <div class="actions">
      <el-button type="primary" @click="showAddDialog">添加用户</el-button>
    </div>
    
    <el-table :data="users" border style="width: 100%">
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
    
    <!-- 添加/编辑用户对话框 -->
    <el-dialog 
      :title="dialogType === 'add' ? '添加用户' : '编辑用户'" 
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="userForm" label-width="100px">
        <el-form-item label="用户名">
          <el-input v-model="userForm.username" placeholder="请输入用户名"></el-input>
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="userForm.email" placeholder="请输入邮箱"></el-input>
        </el-form-item>
        <el-form-item label="密码" v-if="dialogType === 'add'">
          <el-input v-model="userForm.password" type="password" placeholder="请输入密码"></el-input>
        </el-form-item>
        <el-form-item label="角色">
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

<script>
export default {
  name: 'Users',
  data() {
    return {
      users: [
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
        }
      ],
      dialogVisible: false,
      dialogType: 'add', // 'add' 或 'edit'
      userForm: {
        id: null,
        username: '',
        email: '',
        password: '',
        role: 'user'
      }
    }
  },
  methods: {
    showAddDialog() {
      this.dialogType = 'add'
      this.userForm = {
        id: null,
        username: '',
        email: '',
        password: '',
        role: 'user'
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogType = 'edit'
      this.userForm = {
        id: row.id,
        username: row.username,
        email: row.email,
        role: row.role
      }
      this.dialogVisible = true
    },
    handleToggleStatus(row) {
      // 实现启用/禁用功能
      const action = row.status ? '禁用' : '启用'
      this.$message.success(`已${action}用户：${row.username}`)
      row.status = !row.status
    },
    handleDelete(row) {
      // 实现删除功能
      this.$confirm(`确定要删除用户 ${row.username} 吗?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.users = this.users.filter(u => u.id !== row.id)
        this.$message.success('删除成功')
      }).catch(() => {
        this.$message.info('已取消删除')
      })
    },
    handleSaveUser() {
      if (this.dialogType === 'add') {
        // 添加新用户
        const newUser = {
          ...this.userForm,
          id: Date.now(),
          created: new Date().toLocaleString(),
          lastLogin: '-',
          status: true
        }
        this.users.push(newUser)
        this.$message.success('添加成功')
      } else {
        // 更新现有用户
        const index = this.users.findIndex(u => u.id === this.userForm.id)
        if (index !== -1) {
          this.users[index] = { ...this.users[index], ...this.userForm }
          this.$message.success('更新成功')
        }
      }
      this.dialogVisible = false
    }
  }
}
</script>

<style scoped>
.users-container {
  padding: 20px;
}

.actions {
  margin-bottom: 20px;
}
</style> 