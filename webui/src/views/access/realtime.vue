<template>
  <div class="realtime-logs">
    <div class="controls">
      <el-button
        :type="isRunning ? 'danger' : 'primary'"
        @click="toggleRealtime"
        :loading="loading"
      >
        <el-icon><VideoPlay v-if="!isRunning" /><VideoPause v-else /></el-icon>
        {{ isRunning ? '暂停' : '开始' }}
      </el-button>
      
      <el-button @click="clearLogs">
        <el-icon><Delete /></el-icon>
        清空
      </el-button>
      
      <el-select v-model="refreshInterval" @change="updateInterval" style="width: 150px">
        <el-option label="1秒刷新" :value="1000" />
        <el-option label="3秒刷新" :value="3000" />
        <el-option label="5秒刷新" :value="5000" />
        <el-option label="10秒刷新" :value="10000" />
      </el-select>
      
      <el-text type="info">
        共 {{ realtimeLogs.length }} 条记录
      </el-text>
    </div>

    <div class="log-container" ref="logContainer">
      <div
        v-for="(log, index) in realtimeLogs"
        :key="`${log.id}-${index}`"
        class="log-item"
        :class="getLogClass(log)"
      >
        <div class="log-time">
          {{ formatTime(log.timestamp) }}
        </div>
        <div class="log-content">
          <el-tag :type="getActionType(log.action)" size="small" class="log-action">
            {{ getActionText(log.action) }}
          </el-tag>
          
          <span class="log-ip">{{ log.ip }}</span>
          
          <el-tag :type="getMethodType(log.method)" size="small" class="log-method">
            {{ log.method }}
          </el-tag>
          
          <span class="log-path">{{ log.path }}</span>
          
          <span class="log-score" :class="getScoreClass(log.score)">
            {{ log.score }}分
          </span>
          
          <el-button
            type="text"
            size="small"
            @click="handleViewDetails(log)"
            class="log-details-btn"
          >
            详情
          </el-button>
        </div>
      </div>
      
      <div v-if="realtimeLogs.length === 0" class="empty-state">
        <el-icon class="empty-icon"><Document /></el-icon>
        <div class="empty-text">暂无实时日志数据</div>
        <div class="empty-hint">点击"开始"按钮开始监控实时访问</div>
      </div>
    </div>

    <!-- 日志详情对话框 -->
    <el-dialog
      v-model="detailDialog.visible"
      title="访问详情"
      width="600px"
      destroy-on-close
    >
      <el-descriptions v-if="detailDialog.data" :column="1" border>
        <el-descriptions-item label="时间">{{ formatTime(detailDialog.data.timestamp) }}</el-descriptions-item>
        <el-descriptions-item label="用户指纹">{{ detailDialog.data.fingerprint }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ detailDialog.data.ip }}</el-descriptions-item>
        <el-descriptions-item label="访问路径">{{ detailDialog.data.path }}</el-descriptions-item>
        <el-descriptions-item label="请求方法">
          <el-tag :type="getMethodType(detailDialog.data.method)" size="small">
            {{ detailDialog.data.method }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="用户分数">
          <span :class="getScoreClass(detailDialog.data.score)">
            {{ detailDialog.data.score }}分
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="处理动作">
          <el-tag :type="getActionType(detailDialog.data.action)" size="small">
            {{ getActionText(detailDialog.data.action) }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
      
      <template #footer>
        <el-button @click="detailDialog.visible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay, VideoPause, Delete, Document } from '@element-plus/icons-vue'
import { getRecentAccessLogs } from '@/api/logs'

export default {
  name: 'RealtimeLogs',
  components: {
    VideoPlay,
    VideoPause,
    Delete,
    Document
  },
  emits: ['close'],
  setup(props, { emit }) {
    const loading = ref(false)
    const isRunning = ref(false)
    const refreshInterval = ref(3000)
    const realtimeLogs = ref([])
    const logContainer = ref(null)
    
    let intervalId = null

    // 详情对话框
    const detailDialog = reactive({
      visible: false,
      data: null
    })

    // 获取实时日志数据
    const fetchRealtimeLogs = async () => {
      try {
        const response = await getRecentAccessLogs({
          minutes: 5, // 获取最近5分钟的数据
          limit: 100
        })
        
        if (response.success && response.data.records) {
          // 合并新数据，避免重复
          const newLogs = response.data.records
          const existingIds = new Set(realtimeLogs.value.map(log => log.id))
          
          const uniqueNewLogs = newLogs.filter(log => !existingIds.has(log.id))
          
          if (uniqueNewLogs.length > 0) {
            realtimeLogs.value = [...uniqueNewLogs, ...realtimeLogs.value]
              .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
              .slice(0, 200) // 最多保留200条记录
            
            // 自动滚动到顶部
            nextTick(() => {
              if (logContainer.value) {
                logContainer.value.scrollTop = 0
              }
            })
          }
        }
      } catch (error) {
        console.error('获取实时日志失败:', error)
        if (isRunning.value) {
          ElMessage.error('获取实时日志失败')
        }
      }
    }

    // 开始/暂停实时监控
    const toggleRealtime = async () => {
      if (isRunning.value) {
        // 停止
        if (intervalId) {
          clearInterval(intervalId)
          intervalId = null
        }
        isRunning.value = false
      } else {
        // 开始
        loading.value = true
        try {
          await fetchRealtimeLogs()
          
          intervalId = setInterval(fetchRealtimeLogs, refreshInterval.value)
          isRunning.value = true
        } catch (error) {
          ElMessage.error('启动实时监控失败')
        } finally {
          loading.value = false
        }
      }
    }

    // 更新刷新间隔
    const updateInterval = () => {
      if (isRunning.value && intervalId) {
        clearInterval(intervalId)
        intervalId = setInterval(fetchRealtimeLogs, refreshInterval.value)
      }
    }

    // 清空日志
    const clearLogs = () => {
      realtimeLogs.value = []
    }

    // 查看详情
    const handleViewDetails = (log) => {
      detailDialog.data = { ...log }
      detailDialog.visible = true
    }

    // 格式化时间
    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString('zh-CN', {
        hour12: false,
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    }

    // 获取日志样式类
    const getLogClass = (log) => {
      return {
        'log-item-allow': log.action === 'allow',
        'log-item-limit': log.action === 'limit',
        'log-item-challenge': log.action === 'challenge',
        'log-item-ban': log.action === 'ban'
      }
    }

    // 获取动作类型
    const getActionType = (action) => {
      const types = {
        allow: 'success',
        limit: 'warning',
        challenge: 'info',
        ban: 'danger'
      }
      return types[action] || 'info'
    }

    // 获取动作文本
    const getActionText = (action) => {
      const texts = {
        allow: '通过',
        limit: '限制',
        challenge: '验证',
        ban: '封禁'
      }
      return texts[action] || action
    }

    // 获取方法类型
    const getMethodType = (method) => {
      const types = {
        GET: 'success',
        POST: 'primary',
        PUT: 'warning',
        DELETE: 'danger'
      }
      return types[method] || 'info'
    }

    // 获取分数样式
    const getScoreClass = (score) => {
      if (score >= 80) return 'score-excellent'
      if (score >= 60) return 'score-good'
      if (score >= 30) return 'score-warning'
      return 'score-danger'
    }

    onMounted(() => {
      // 初始加载一次数据
      fetchRealtimeLogs()
    })

    onUnmounted(() => {
      // 清理定时器
      if (intervalId) {
        clearInterval(intervalId)
      }
    })

    return {
      loading,
      isRunning,
      refreshInterval,
      realtimeLogs,
      logContainer,
      detailDialog,
      toggleRealtime,
      updateInterval,
      clearLogs,
      handleViewDetails,
      formatTime,
      getLogClass,
      getActionType,
      getActionText,
      getMethodType,
      getScoreClass
    }
  }
}
</script>

<style scoped>
.realtime-logs {
  height: 600px;
  display: flex;
  flex-direction: column;
}

.controls {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

.log-container {
  flex: 1;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
}

.log-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.2s;
  animation: slideIn 0.3s ease-out;
}

.log-item:hover {
  background-color: #f5f7fa;
}

.log-item:last-child {
  border-bottom: none;
}

.log-item-allow {
  border-left: 3px solid #67c23a;
}

.log-item-limit {
  border-left: 3px solid #e6a23c;
}

.log-item-challenge {
  border-left: 3px solid #409eff;
}

.log-item-ban {
  border-left: 3px solid #f56c6c;
}

.log-time {
  flex-shrink: 0;
  width: 160px;
  font-size: 12px;
  color: #909399;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.log-content {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.log-action,
.log-method {
  flex-shrink: 0;
}

.log-ip {
  flex-shrink: 0;
  width: 120px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.log-path {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-score {
  flex-shrink: 0;
  width: 50px;
  font-weight: bold;
  font-size: 12px;
}

.log-details-btn {
  flex-shrink: 0;
  font-size: 12px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-text {
  font-size: 16px;
  margin-bottom: 8px;
}

.empty-hint {
  font-size: 14px;
  opacity: 0.7;
}

.score-excellent {
  color: #67c23a;
}

.score-good {
  color: #409eff;
}

.score-warning {
  color: #e6a23c;
}

.score-danger {
  color: #f56c6c;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 滚动条样式 */
.log-container::-webkit-scrollbar {
  width: 6px;
}

.log-container::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.log-container::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.log-container::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>
