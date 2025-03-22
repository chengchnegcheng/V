<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色权限管理</span>
          <el-button type="primary" @click="handleAdd">添加角色</el-button>
        </div>
      </template>

      <el-table :data="roles" v-loading="loading" border>
        <el-table-column prop="name" label="角色名称" />
        <el-table-column prop="code" label="角色标识" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="权限" min-width="300">
          <template #default="{ row }">
            <el-tag
              v-for="permission in row.permissions"
              :key="permission"
              class="permission-tag"
              size="small"
            >
              {{ permission }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button 
                size="small" 
                type="danger"
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 角色表单对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="角色标识" prop="code">
          <el-input v-model="form.code" :disabled="!!form.id" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="权限" prop="permissions">
          <el-checkbox-group v-model="form.permissions">
            <el-checkbox label="user:view">查看用户</el-checkbox>
            <el-checkbox label="user:create">创建用户</el-checkbox>
            <el-checkbox label="user:edit">编辑用户</el-checkbox>
            <el-checkbox label="user:delete">删除用户</el-checkbox>
            <el-checkbox label="proxy:view">查看代理</el-checkbox>
            <el-checkbox label="proxy:create">创建代理</el-checkbox>
            <el-checkbox label="proxy:edit">编辑代理</el-checkbox>
            <el-checkbox label="proxy:delete">删除代理</el-checkbox>
            <el-checkbox label="stats:view">查看统计</el-checkbox>
            <el-checkbox label="system:view">查看系统</el-checkbox>
            <el-checkbox label="system:edit">编辑系统</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { roles } from '@/api'

export default {
  name: 'Roles',
  setup() {
    const loading = ref(false)
    const roles = ref([])
    const dialogVisible = ref(false)
    const dialogTitle = ref('')
    const formRef = ref(null)

    const form = reactive({
      id: null,
      name: '',
      code: '',
      description: '',
      permissions: []
    })

    const rules = {
      name: [
        { required: true, message: '请输入角色名称', trigger: 'blur' },
        { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
      ],
      code: [
        { required: true, message: '请输入角色标识', trigger: 'blur' },
        { pattern: /^[a-z][a-z0-9:]*$/, message: '只能包含小写字母、数字和冒号，且必须以字母开头', trigger: 'blur' }
      ],
      permissions: [
        { type: 'array', required: true, message: '请至少选择一个权限', trigger: 'change' }
      ]
    }

    const fetchRoles = async () => {
      loading.value = true
      try {
        // 模拟获取角色列表
        // const response = await roles.list()
        // roles.value = response.data
        
        // 使用模拟数据
        setTimeout(() => {
          roles.value = [
            {
              id: 1,
              name: '管理员',
              code: 'admin',
              description: '系统管理员，拥有所有权限',
              permissions: ['user:view', 'user:create', 'user:edit', 'user:delete', 'proxy:view', 'proxy:create', 'proxy:edit', 'proxy:delete', 'stats:view', 'system:view', 'system:edit']
            },
            {
              id: 2,
              name: '操作员',
              code: 'operator',
              description: '系统操作员，拥有部分管理权限',
              permissions: ['user:view', 'proxy:view', 'proxy:create', 'proxy:edit', 'stats:view', 'system:view']
            },
            {
              id: 3,
              name: '访客',
              code: 'guest',
              description: '访客用户，只有查看权限',
              permissions: ['user:view', 'proxy:view', 'stats:view']
            }
          ]
          loading.value = false
        }, 500)
      } catch (error) {
        ElMessage.error('获取角色列表失败')
        loading.value = false
      }
    }

    const handleAdd = () => {
      dialogTitle.value = '添加角色'
      Object.assign(form, {
        id: null,
        name: '',
        code: '',
        description: '',
        permissions: []
      })
      dialogVisible.value = true
    }

    const handleEdit = (row) => {
      dialogTitle.value = '编辑角色'
      Object.assign(form, row)
      dialogVisible.value = true
    }

    const handleSubmit = async () => {
      if (!formRef.value) return
      
      await formRef.value.validate(async (valid) => {
        if (valid) {
          try {
            if (form.id) {
              await roles.update(form.id, form)
              ElMessage.success('更新成功')
            } else {
              await roles.create(form)
              ElMessage.success('创建成功')
            }
            dialogVisible.value = false
            fetchRoles()
          } catch (error) {
            ElMessage.error('操作失败')
          }
        }
      })
    }

    const handleDelete = async (row) => {
      try {
        await ElMessageBox.confirm(
          `确定要删除角色 ${row.name} 吗？`,
          '警告',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        await roles.delete(row.id)
        ElMessage.success('删除成功')
        fetchRoles()
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage.error('删除失败')
        }
      }
    }

    onMounted(() => {
      fetchRoles()
    })

    return {
      loading,
      roles,
      dialogVisible,
      dialogTitle,
      form,
      formRef,
      rules,
      handleAdd,
      handleEdit,
      handleSubmit,
      handleDelete
    }
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.permission-tag {
  margin-right: 5px;
  margin-bottom: 5px;
}

.el-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
</style> 