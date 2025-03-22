<template>
  <div class="settings-container">
    <h1>系统设置</h1>
    
    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <h3>常规设置</h3>
        </div>
      </template>
      
      <el-form :model="generalSettings" label-position="top">
        <el-form-item label="系统名称">
          <el-input v-model="generalSettings.siteName"></el-input>
        </el-form-item>
        
        <el-form-item label="管理员邮箱">
          <el-input v-model="generalSettings.adminEmail"></el-input>
        </el-form-item>
        
        <el-form-item label="允许注册">
          <el-switch v-model="generalSettings.allowRegistration"></el-switch>
        </el-form-item>
        
        <el-form-item label="默认流量限制 (GB)">
          <el-input-number v-model="generalSettings.defaultTrafficLimit" :min="1" :max="10000"></el-input-number>
        </el-form-item>
        
        <el-form-item label="面板端口">
          <el-input-number v-model="generalSettings.panelPort" :min="1" :max="65535"></el-input-number>
        </el-form-item>
      </el-form>
      
      <el-button type="primary" @click="saveGeneralSettings">保存设置</el-button>
    </el-card>
    
    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <h3>安全设置</h3>
        </div>
      </template>
      
      <el-form :model="securitySettings" label-position="top">
        <el-form-item label="会话超时时间 (分钟)">
          <el-input-number v-model="securitySettings.sessionTimeout" :min="1" :max="1440"></el-input-number>
        </el-form-item>
        
        <el-form-item label="允许的登录失败次数">
          <el-input-number v-model="securitySettings.maxLoginAttempts" :min="1" :max="10"></el-input-number>
        </el-form-item>
        
        <el-form-item label="启用两步验证">
          <el-switch v-model="securitySettings.enable2FA"></el-switch>
        </el-form-item>
        
        <el-form-item label="允许的IP地址">
          <el-input v-model="securitySettings.allowedIPs" type="textarea" :rows="3" placeholder="每行一个IP地址，留空表示允许所有"></el-input>
        </el-form-item>
      </el-form>
      
      <el-button type="primary" @click="saveSecuritySettings">保存设置</el-button>
    </el-card>
    
    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <h3>备份设置</h3>
        </div>
      </template>
      
      <el-form :model="backupSettings" label-position="top">
        <el-form-item label="自动备份">
          <el-switch v-model="backupSettings.autoBackup"></el-switch>
        </el-form-item>
        
        <el-form-item label="备份频率" v-if="backupSettings.autoBackup">
          <el-select v-model="backupSettings.backupFrequency" style="width: 100%">
            <el-option label="每天" value="daily"></el-option>
            <el-option label="每周" value="weekly"></el-option>
            <el-option label="每月" value="monthly"></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="备份保留时间 (天)" v-if="backupSettings.autoBackup">
          <el-input-number v-model="backupSettings.backupRetention" :min="1" :max="365"></el-input-number>
        </el-form-item>
        
        <el-form-item label="备份存储路径" v-if="backupSettings.autoBackup">
          <el-input v-model="backupSettings.backupPath"></el-input>
        </el-form-item>
      </el-form>
      
      <div class="form-actions">
        <el-button type="primary" @click="saveBackupSettings">保存设置</el-button>
        <el-button type="success" @click="createBackup">立即备份</el-button>
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Settings',
  data() {
    return {
      generalSettings: {
        siteName: 'V 多协议代理面板',
        adminEmail: 'admin@example.com',
        allowRegistration: false,
        defaultTrafficLimit: 50,
        panelPort: 8080
      },
      securitySettings: {
        sessionTimeout: 60,
        maxLoginAttempts: 5,
        enable2FA: false,
        allowedIPs: ''
      },
      backupSettings: {
        autoBackup: true,
        backupFrequency: 'daily',
        backupRetention: 7,
        backupPath: './data/backups'
      }
    }
  },
  methods: {
    saveGeneralSettings() {
      // 实现保存常规设置
      this.$message.success('常规设置已保存')
    },
    saveSecuritySettings() {
      // 实现保存安全设置
      this.$message.success('安全设置已保存')
    },
    saveBackupSettings() {
      // 实现保存备份设置
      this.$message.success('备份设置已保存')
    },
    createBackup() {
      // 实现立即备份功能
      this.$message.success('备份已创建')
    }
  }
}
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.setting-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-actions {
  display: flex;
  gap: 10px;
}
</style> 