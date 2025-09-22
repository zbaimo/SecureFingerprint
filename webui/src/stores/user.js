import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    userInfo: {
      username: 'admin',
      avatar: '',
      roles: ['admin'],
      permissions: []
    },
    token: '',
    isLoggedIn: false
  }),

  getters: {
    username: (state) => state.userInfo.username,
    avatar: (state) => state.userInfo.avatar,
    roles: (state) => state.userInfo.roles,
    hasPermission: (state) => (permission) => {
      return state.userInfo.permissions.includes(permission) || 
             state.userInfo.roles.includes('admin')
    }
  },

  actions: {
    // 登录
    async login(loginForm) {
      try {
        // 模拟登录API调用
        if (loginForm.username === 'admin' && loginForm.password === 'admin123') {
          this.token = 'mock-token-' + Date.now()
          this.isLoggedIn = true
          this.userInfo = {
            username: loginForm.username,
            avatar: '',
            roles: ['admin'],
            permissions: ['*']
          }
          
          // 保存到localStorage
          localStorage.setItem('token', this.token)
          localStorage.setItem('userInfo', JSON.stringify(this.userInfo))
          
          return { success: true }
        } else {
          throw new Error('用户名或密码错误')
        }
      } catch (error) {
        return { success: false, error: error.message }
      }
    },

    // 登出
    logout() {
      this.token = ''
      this.isLoggedIn = false
      this.userInfo = {
        username: '',
        avatar: '',
        roles: [],
        permissions: []
      }
      
      // 清除localStorage
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
    },

    // 恢复登录状态
    restoreLogin() {
      const token = localStorage.getItem('token')
      const userInfo = localStorage.getItem('userInfo')
      
      if (token && userInfo) {
        this.token = token
        this.isLoggedIn = true
        this.userInfo = JSON.parse(userInfo)
      }
    },

    // 更新用户信息
    updateUserInfo(userInfo) {
      this.userInfo = { ...this.userInfo, ...userInfo }
      localStorage.setItem('userInfo', JSON.stringify(this.userInfo))
    }
  }
})
