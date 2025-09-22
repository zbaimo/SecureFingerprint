<template>
  <div class="login-container">
    <div class="login-form">
      <div class="logo-section">
        <div class="logo">🛡️</div>
        <h1 class="title">SecureFingerprint</h1>
        <p class="subtitle">智能访问控制系统</p>
      </div>

      <el-form
        :model="loginForm"
        :rules="loginRules"
        ref="loginFormRef"
        class="login-form-content"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>
        
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>
        
        <el-form-item>
          <el-checkbox v-model="loginForm.remember">记住登录状态</el-checkbox>
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            class="login-button"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <p>SecureFingerprint v1.0.0</p>
        <p>基于用户指纹的智能访问控制系统</p>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'

export default {
  name: 'Login',
  components: {
    User,
    Lock
  },
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const loginFormRef = ref(null)

    // 登录表单
    const loginForm = reactive({
      username: '',
      password: '',
      remember: false
    })

    // 表单验证规则
    const loginRules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ]
    }

    // 处理登录
    const handleLogin = async () => {
      if (!loginFormRef.value) return

      try {
        await loginFormRef.value.validate()
        
        loading.value = true
        
        // 模拟登录请求
        await new Promise(resolve => setTimeout(resolve, 1500))
        
        // 模拟登录验证
        if (loginForm.username === 'admin' && loginForm.password === 'admin123') {
          ElMessage.success('登录成功')
          
          // 保存登录状态
          if (loginForm.remember) {
            localStorage.setItem('remember_login', 'true')
          }
          
          // 跳转到仪表板
          router.push('/dashboard')
        } else {
          ElMessage.error('用户名或密码错误')
        }
      } catch (error) {
        console.error('登录失败:', error)
      } finally {
        loading.value = false
      }
    }

    return {
      loading,
      loginForm,
      loginRules,
      loginFormRef,
      handleLogin,
      User,
      Lock
    }
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-form {
  width: 400px;
  background: white;
  border-radius: 20px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  padding: 40px;
}

.logo-section {
  text-align: center;
  margin-bottom: 40px;
}

.logo {
  font-size: 64px;
  margin-bottom: 16px;
}

.title {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.subtitle {
  color: #909399;
  font-size: 14px;
  margin-bottom: 0;
}

.login-form-content {
  margin-bottom: 30px;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
}

.login-footer {
  text-align: center;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}

.login-footer p {
  margin: 4px 0;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-form {
    width: 100%;
    max-width: 360px;
    padding: 30px 20px;
  }
  
  .title {
    font-size: 24px;
  }
  
  .logo {
    font-size: 48px;
  }
}
</style>
