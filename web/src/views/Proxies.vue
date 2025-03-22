<template>
  <div class="proxies-container">
    <h1>协议管理</h1>
    
    <div class="actions">
      <el-button type="primary" @click="showAddDialog">添加协议</el-button>
    </div>
    
    <el-table :data="proxies" border style="width: 100%">
      <el-table-column prop="name" label="名称" width="180"></el-table-column>
      <el-table-column prop="type" label="协议类型" width="120"></el-table-column>
      <el-table-column prop="port" label="端口" width="100"></el-table-column>
      <el-table-column prop="users" label="用户数" width="100"></el-table-column>
      <el-table-column prop="created" label="创建时间" width="180"></el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.status ? 'success' : 'danger'">
            {{ scope.row.status ? '运行中' : '已停止' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="scope">
          <el-button-group>
            <el-button size="small" type="primary" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="small" type="success" @click="handleView(scope.row)">查看</el-button>
            <el-button 
              size="small" 
              :type="scope.row.status ? 'warning' : 'success'"
              @click="handleToggleStatus(scope.row)"
            >
              {{ scope.row.status ? '停止' : '启动' }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 添加/编辑协议对话框 -->
    <el-dialog 
      :title="dialogType === 'add' ? '添加协议' : '编辑协议'" 
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="proxyForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="proxyForm.name" placeholder="请输入协议名称"></el-input>
        </el-form-item>
        <el-form-item label="协议类型">
          <el-select v-model="proxyForm.type" placeholder="请选择协议类型" style="width: 100%">
            <el-option label="VMess" value="vmess"></el-option>
            <el-option label="VLess" value="vless"></el-option>
            <el-option label="Trojan" value="trojan"></el-option>
            <el-option label="Shadowsocks" value="shadowsocks"></el-option>
            <el-option label="Socks" value="socks"></el-option>
            <el-option label="HTTP" value="http"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="proxyForm.port" :min="1" :max="65535" style="width: 100%"></el-input-number>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveProxy">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: 'Proxies',
  data() {
    return {
      proxies: [
        {
          id: 1,
          name: '默认VMess协议',
          type: 'vmess',
          port: 10086,
          users: 2,
          created: '2023-01-01 12:00:00',
          status: true
        },
        {
          id: 2,
          name: '默认Trojan协议',
          type: 'trojan',
          port: 443,
          users: 3,
          created: '2023-01-01 12:00:00',
          status: true
        }
      ],
      dialogVisible: false,
      dialogType: 'add', // 'add' 或 'edit'
      proxyForm: {
        id: null,
        name: '',
        type: '',
        port: 10000
      }
    }
  },
  methods: {
    showAddDialog() {
      this.dialogType = 'add'
      this.proxyForm = {
        id: null,
        name: '',
        type: '',
        port: 10000
      }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogType = 'edit'
      this.proxyForm = { ...row }
      this.dialogVisible = true
    },
    handleView(row) {
      // 实现查看详情功能
      this.$message.info('查看协议详情：' + row.name)
    },
    handleToggleStatus(row) {
      // 实现启动/停止功能
      const action = row.status ? '停止' : '启动'
      this.$message.success(`已${action}协议：${row.name}`)
      row.status = !row.status
    },
    handleDelete(row) {
      // 实现删除功能
      this.$confirm(`确定要删除协议 ${row.name} 吗?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.proxies = this.proxies.filter(p => p.id !== row.id)
        this.$message.success('删除成功')
      }).catch(() => {
        this.$message.info('已取消删除')
      })
    },
    handleSaveProxy() {
      if (this.dialogType === 'add') {
        // 添加新协议
        const newProxy = {
          ...this.proxyForm,
          id: Date.now(),
          users: 0,
          created: new Date().toLocaleString(),
          status: true
        }
        this.proxies.push(newProxy)
        this.$message.success('添加成功')
      } else {
        // 更新现有协议
        const index = this.proxies.findIndex(p => p.id === this.proxyForm.id)
        if (index !== -1) {
          this.proxies[index] = { ...this.proxies[index], ...this.proxyForm }
          this.$message.success('更新成功')
        }
      }
      this.dialogVisible = false
    }
  }
}
</script>

<style scoped>
.proxies-container {
  padding: 20px;
}

.actions {
  margin-bottom: 20px;
}
</style> 