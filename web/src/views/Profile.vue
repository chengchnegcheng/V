<template>
  <div class="profile-container">
    <div class="profile-header">
      <h2>个人资料</h2>
      <p>查看和管理您的个人信息</p>
    </div>
    
    <div class="profile-content">
      <el-card class="profile-card">
        <div class="avatar-section">
          <el-avatar :size="100" class="avatar">{{ avatarText }}</el-avatar>
          <div class="avatar-actions">
            <el-button type="primary" size="small">更换头像</el-button>
          </div>
        </div>
        
        <el-form 
          :model="userForm" 
          label-width="100px" 
          class="profile-form"
          :rules="rules"
          ref="profileForm"
        >
          <el-form-item label="用户名" prop="username">
            <el-input v-model="userForm.username" disabled></el-input>
          </el-form-item>
          
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="userForm.email" placeholder="请输入邮箱"></el-input>
          </el-form-item>
          
          <el-form-item label="真实姓名" prop="realName">
            <el-input v-model="userForm.realName" placeholder="请输入真实姓名"></el-input>
          </el-form-item>
          
          <el-form-item label="手机号码" prop="phone">
            <el-input v-model="userForm.phone" placeholder="请输入手机号码"></el-input>
          </el-form-item>
          
          <el-form-item label="上次登录">
            <div>{{ formatDate(userForm.lastLogin) }}</div>
          </el-form-item>
          
          <el-form-item label="注册时间">
            <div>{{ formatDate(userForm.createdAt) }}</div>
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="saveProfile">保存修改</el-button>
            <el-button @click="resetForm">重置</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const profileForm = ref(null)

const userForm = ref({
  username: 'admin',
  email: 'admin@example.com',
  realName: '',
  phone: '',
  lastLogin: new Date(),
  createdAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
})

const avatarText = computed(() => {
  return userForm.value.username.charAt(0).toUpperCase()
})

const rules = {
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号码', trigger: 'blur' }
  ]
}

const formatDate = (date) => {
  if (!date) return '--'
  const d = new Date(date)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const saveProfile = async () => {
  if (!profileForm.value) return
  
  await profileForm.value.validate(async (valid) => {
    if (valid) {
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 500))
        
        // 更新用户信息
        userStore.updateUserProfile({
          email: userForm.value.email,
          realName: userForm.value.realName,
          phone: userForm.value.phone
        })
        
        ElMessage.success('个人资料已更新')
      } catch (error) {
        console.error('更新个人资料失败:', error)
        ElMessage.error('更新个人资料失败')
      }
    }
  })
}

const resetForm = () => {
  if (profileForm.value) {
    profileForm.value.resetFields()
  }
}

onMounted(async () => {
  try {
    // 模拟从API获取用户资料
    // 这里可以替换为实际的API调用
    await new Promise(resolve => setTimeout(resolve, 300))
    
    // 如果userStore中有用户信息，则使用它
    if (userStore.user) {
      userForm.value.username = userStore.user.username || 'admin'
      userForm.value.email = userStore.user.email || ''
      userForm.value.realName = userStore.user.realName || ''
      userForm.value.phone = userStore.user.phone || ''
      userForm.value.lastLogin = userStore.user.lastLogin || new Date()
      userForm.value.createdAt = userStore.user.createdAt || new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
    }
  } catch (error) {
    console.error('获取用户资料失败:', error)
    ElMessage.error('获取用户资料失败')
  }
})
</script>

<style scoped>
.profile-container {
  padding: 20px;
}

.profile-header {
  margin-bottom: 20px;
}

.profile-header h2 {
  font-size: 24px;
  margin: 0 0 8px;
  color: #303133;
}

.profile-header p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.profile-content {
  display: flex;
  justify-content: center;
}

.profile-card {
  width: 100%;
  max-width: 700px;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 30px;
}

.avatar {
  margin-bottom: 15px;
  background-color: #1890ff;
  font-size: 40px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.profile-form {
  max-width: 500px;
  margin: 0 auto;
}

.avatar-actions {
  margin-top: 10px;
}
</style> 