<template>
  <div class="error-404">
    <div class="error-content">
      <div class="error-icon">🔍</div>
      <h1 class="error-title">404</h1>
      <p class="error-message">抱歉，您访问的页面不存在</p>
      <p class="error-description">
        页面可能已被删除、重命名或暂时不可用
      </p>
      
      <div class="error-actions">
        <el-button type="primary" @click="goHome">
          <el-icon><HomeFilled /></el-icon>
          返回首页
        </el-button>
        <el-button @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回上页
        </el-button>
      </div>
      
      <div class="suggestions">
        <h3>您可以尝试：</h3>
        <ul>
          <li>检查URL是否输入正确</li>
          <li>返回首页重新导航</li>
          <li>使用搜索功能查找内容</li>
          <li>联系管理员获取帮助</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { HomeFilled, ArrowLeft } from '@element-plus/icons-vue'

export default {
  name: 'Error404',
  components: {
    HomeFilled,
    ArrowLeft
  },
  setup() {
    const router = useRouter()

    const goHome = () => {
      router.push('/dashboard')
    }

    const goBack = () => {
      router.go(-1)
    }

    return {
      goHome,
      goBack
    }
  }
}
</script>

<style scoped>
.error-404 {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  padding: 20px;
}

.error-content {
  text-align: center;
  max-width: 500px;
}

.error-icon {
  font-size: 120px;
  margin-bottom: 20px;
  opacity: 0.5;
}

.error-title {
  font-size: 72px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 16px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.error-message {
  font-size: 24px;
  color: #303133;
  margin-bottom: 12px;
}

.error-description {
  font-size: 16px;
  color: #606266;
  margin-bottom: 40px;
  line-height: 1.5;
}

.error-actions {
  margin-bottom: 40px;
}

.error-actions .el-button {
  margin: 0 8px;
  padding: 12px 24px;
}

.suggestions {
  text-align: left;
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.suggestions h3 {
  color: #303133;
  margin-bottom: 16px;
  font-size: 16px;
}

.suggestions ul {
  margin: 0;
  padding-left: 20px;
}

.suggestions li {
  color: #606266;
  margin-bottom: 8px;
  line-height: 1.5;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .error-title {
    font-size: 48px;
  }
  
  .error-message {
    font-size: 20px;
  }
  
  .error-description {
    font-size: 14px;
  }
  
  .error-actions .el-button {
    display: block;
    width: 100%;
    margin: 8px 0;
  }
}
</style>
