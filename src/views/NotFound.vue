<template>
  <div class="not-found">
    <el-result
      icon="error"
      title="404"
      sub-title="抱歉，您请求的页面不存在"
    >
      <template #extra>
        <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup>
// 404页面组件
</script>

<style scoped>
.not-found {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 80vh;
}
</style> 