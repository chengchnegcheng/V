<template>
  <div class="roles-container">
    <div class="page-header">
      <h1>角色管理</h1>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon> 新建角色
      </el-button>
    </div>

    <!-- 角色列表 -->
    <el-card shadow="never" v-loading="loading">
      <template v-if="roles.length > 0">
        <el-table :data="roles" style="width: 100%" border>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="角色名称" />
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="created_at" label="创建时间" width="180" />
          <el-table-column label="操作" width="200" align="center">
            <template #default="scope">
              <el-button size="small" @click="showEditDialog(scope.row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDelete(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </template>

      <template v-else>
        <div class="empty-data">
          <el-empty description="暂无角色数据" />
        </div>
      </template>
    </el-card>

    <!-- 创建/编辑角色对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="roleForm" label-width="80px" :rules="rules" ref="roleFormRef">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="roleForm.description" 
            placeholder="请输入角色描述"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">确认</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, onMounted, reactive } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import axios from 'axios';

export default {
  name: 'Roles',
  components: {
    Plus
  },
  setup() {
    const roles = ref([]);
    const loading = ref(false);
    const dialogVisible = ref(false);
    const dialogTitle = ref('新建角色');
    const submitting = ref(false);
    const roleFormRef = ref(null);
    
    const roleForm = reactive({
      id: null,
      name: '',
      description: ''
    });
    
    const rules = {
      name: [
        { required: true, message: '请输入角色名称', trigger: 'blur' },
        { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
      ]
    };
    
    // 加载角色列表
    const loadRoles = async () => {
      loading.value = true;
      try {
        // 使用模拟数据
        setTimeout(() => {
          roles.value = [
            { id: 1, name: '超级管理员', description: '拥有所有权限', created_at: '2025-03-24 10:00:00' },
            { id: 2, name: '普通管理员', description: '拥有部分管理权限', created_at: '2025-03-24 10:30:00' },
            { id: 3, name: '用户', description: '普通用户权限', created_at: '2025-03-24 11:00:00' }
          ];
          loading.value = false;
        }, 500);
        
        // 实际API调用
        /* 
        const response = await axios.get('/api/roles');
        roles.value = response.data;
        */
      } catch (error) {
        console.error('获取角色列表失败:', error);
        ElMessage.error('获取角色列表失败');
      } finally {
        loading.value = false;
      }
    };
    
    // 显示创建对话框
    const showCreateDialog = () => {
      roleForm.id = null;
      roleForm.name = '';
      roleForm.description = '';
      dialogTitle.value = '新建角色';
      dialogVisible.value = true;
    };
    
    // 显示编辑对话框
    const showEditDialog = (row) => {
      roleForm.id = row.id;
      roleForm.name = row.name;
      roleForm.description = row.description;
      dialogTitle.value = '编辑角色';
      dialogVisible.value = true;
    };
    
    // 提交表单
    const handleSubmit = async () => {
      if (!roleFormRef.value) return;
      
      await roleFormRef.value.validate(async (valid) => {
        if (!valid) return;
        
        submitting.value = true;
        try {
          if (roleForm.id) {
            // 编辑角色
            /* 
            await axios.put(`/api/roles/${roleForm.id}`, {
              name: roleForm.name,
              description: roleForm.description
            });
            */
            
            // 使用模拟数据
            const index = roles.value.findIndex(r => r.id === roleForm.id);
            if (index !== -1) {
              roles.value[index].name = roleForm.name;
              roles.value[index].description = roleForm.description;
            }
            
            ElMessage.success('角色更新成功');
          } else {
            // 创建角色
            /* 
            const response = await axios.post('/api/roles', {
              name: roleForm.name,
              description: roleForm.description
            });
            roles.value.unshift(response.data);
            */
            
            // 使用模拟数据
            const newRole = {
              id: roles.value.length + 1,
              name: roleForm.name,
              description: roleForm.description,
              created_at: new Date().toISOString().replace('T', ' ').substring(0, 19)
            };
            roles.value.unshift(newRole);
            
            ElMessage.success('角色创建成功');
          }
          
          dialogVisible.value = false;
        } catch (error) {
          console.error('操作失败:', error);
          ElMessage.error('操作失败，请重试');
        } finally {
          submitting.value = false;
        }
      });
    };
    
    // 删除角色
    const handleDelete = (row) => {
      ElMessageBox.confirm(
        `确定要删除角色 "${row.name}" 吗？`,
        '警告',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(async () => {
        try {
          // 实际API调用
          /* 
          await axios.delete(`/api/roles/${row.id}`);
          */
          
          // 使用模拟数据
          const index = roles.value.findIndex(r => r.id === row.id);
          if (index !== -1) {
            roles.value.splice(index, 1);
          }
          
          ElMessage.success('删除成功');
        } catch (error) {
          console.error('删除失败:', error);
          ElMessage.error('删除失败，请重试');
        }
      }).catch(() => {});
    };
    
    onMounted(() => {
      loadRoles();
    });
    
    return {
      roles,
      loading,
      dialogVisible,
      dialogTitle,
      roleForm,
      submitting,
      roleFormRef,
      rules,
      loadRoles,
      showCreateDialog,
      showEditDialog,
      handleSubmit,
      handleDelete
    };
  }
};
</script>

<style scoped>
.roles-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
}

.empty-data {
  padding: 40px 0;
}
</style> 