import { defineStore } from 'pinia'
import { getSystemInfo, getHealthCheck } from '@/api/system'

export const useSystemStore = defineStore('system', {
  state: () => ({
    systemInfo: {
      name: 'SecureFingerprint',
      version: 'v1.0.0',
      status: 'running',
      uptime: '',
      userRegistrationAllowed: false
    },
    healthStatus: {
      status: 'healthy',
      services: {},
      timestamp: null
    },
    loading: false
  }),

  getters: {
    isHealthy: (state) => state.healthStatus.status === 'healthy',
    systemVersion: (state) => state.systemInfo.version,
    isRegistrationAllowed: (state) => state.systemInfo.userRegistrationAllowed
  },

  actions: {
    async fetchSystemInfo() {
      this.loading = true
      try {
        const response = await getSystemInfo()
        if (response.success) {
          this.systemInfo = response.data
        }
      } catch (error) {
        console.error('获取系统信息失败:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchHealthStatus() {
      try {
        const response = await getHealthCheck()
        if (response.success) {
          this.healthStatus = response.data
        }
      } catch (error) {
        console.error('获取健康状态失败:', error)
        this.healthStatus.status = 'unhealthy'
      }
    },

    async initSystem() {
      await Promise.all([
        this.fetchSystemInfo(),
        this.fetchHealthStatus()
      ])
    }
  }
})
