<template>
  <div class="user-detail">
    <div class="page-title">
      <el-icon class="title-icon"><InfoFilled /></el-icon>
      用户详情
      <el-button @click="$router.go(-1)" style="margin-left: auto;">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
    </div>

    <!-- 用户基础信息 -->
    <el-card class="info-card">
      <template #header>
        <span>基础信息</span>
      </template>
      
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户指纹">
          <span class="fingerprint-text">{{ userInfo.fingerprint }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="当前分数">
          <span :class="getScoreClass(userInfo.currentScore)">
            {{ userInfo.currentScore }}分
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="风险等级">
          <el-tag :type="getRiskType(userInfo.riskLevel)" size="small">
            {{ getRiskText(userInfo.riskLevel) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="总请求数">{{ userInfo.totalRequests }}</el-descriptions-item>
        <el-descriptions-item label="封禁次数">
          <el-tag v-if="userInfo.banCount > 0" type="danger" size="small">
            {{ userInfo.banCount }}次
          </el-tag>
          <span v-else>0次</span>
        </el-descriptions-item>
        <el-descriptions-item label="最后IP">{{ userInfo.lastIp }}</el-descriptions-item>
        <el-descriptions-item label="首次访问">{{ formatTime(userInfo.firstSeen) }}</el-descriptions-item>
        <el-descriptions-item label="最后访问">{{ formatTime(userInfo.lastSeen) }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 分数趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header">
          <span>分数趋势</span>
          <el-button-group size="small">
            <el-button 
              :type="trendRange === '24h' ? 'primary' : ''"
              @click="changeTrendRange('24h')"
            >24小时</el-button>
            <el-button 
              :type="trendRange === '7d' ? 'primary' : ''"
              @click="changeTrendRange('7d')"
            >7天</el-button>
            <el-button 
              :type="trendRange === '30d' ? 'primary' : ''"
              @click="changeTrendRange('30d')"
            >30天</el-button>
          </el-button-group>
        </div>
      </template>
      
      <div class="chart-container">
        <v-chart :option="scoreTrendOption" />
      </div>
    </el-card>

    <!-- 行为分析 -->
    <el-card class="analysis-card">
      <template #header>
        <div class="card-header">
          <span>行为分析</span>
          <el-button @click="refreshAnalysis" :loading="analysisLoading">
            <el-icon><Refresh /></el-icon>
            重新分析
          </el-button>
        </div>
      </template>
      
      <div class="analysis-content">
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="analysis-item">
              <h4>检测到的行为</h4>
              <div v-for="behavior in behaviorAnalysis.behaviors" :key="behavior.type" class="behavior-item">
                <el-tag :type="getBehaviorType(behavior.severity)" size="small">
                  {{ behavior.type }}
                </el-tag>
                <span class="behavior-desc">{{ behavior.description }}</span>
                <div class="confidence">置信度: {{ (behavior.confidence * 100).toFixed(1) }}%</div>
              </div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="analysis-item">
              <h4>建议操作</h4>
              <ul class="recommendations">
                <li v-for="rec in behaviorAnalysis.recommendations" :key="rec">
                  {{ rec }}
                </li>
              </ul>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>

    <!-- 最近访问记录 -->
    <el-card class="logs-card">
      <template #header>
        <div class="card-header">
          <span>最近访问记录</span>
          <el-button @click="viewAllLogs">
            查看全部
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </template>
      
      <el-table :data="recentLogs" size="small">
        <el-table-column prop="ip" label="IP地址" width="140" />
        <el-table-column prop="path" label="访问路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="method" label="方法" width="80">
          <template #default="scope">
            <el-tag :type="getMethodType(scope.row.method)" size="small">
              {{ scope.row.method }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="动作" width="80">
          <template #default="scope">
            <el-tag :type="getActionType(scope.row.action)" size="small">
              {{ getActionText(scope.row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="score" label="分数" width="80">
          <template #default="scope">
            <span :class="getScoreClass(scope.row.score)">{{ scope.row.score }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 操作按钮 -->
    <div class="action-panel">
      <el-button @click="handleAdjustScore">
        <el-icon><Edit /></el-icon>
        调整分数
      </el-button>
      <el-button @click="handleResetScore">
        <el-icon><RefreshRight /></el-icon>
        重置分数
      </el-button>
      <el-button type="warning" @click="handleAddWhitelist">
        <el-icon><CircleCheck /></el-icon>
        加入白名单
      </el-button>
      <el-button type="danger" @click="handleBanUser">
        <el-icon><CircleClose /></el-icon>
        封禁用户
      </el-button>
    </div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { InfoFilled, ArrowLeft, ArrowRight, Refresh, Edit, RefreshRight, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, LineChart, TitleComponent, TooltipComponent, GridComponent])

export default {
  name: 'UserDetail',
  components: {
    InfoFilled,
    ArrowLeft,
    ArrowRight,
    Refresh,
    Edit,
    RefreshRight,
    CircleCheck,
    CircleClose,
    VChart
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const analysisLoading = ref(false)
    const trendRange = ref('24h')

    const fingerprint = route.params.fingerprint

    // 用户信息
    const userInfo = reactive({
      fingerprint: fingerprint,
      currentScore: 75,
      riskLevel: 'medium',
      totalRequests: 1250,
      banCount: 1,
      lastIp: '192.168.1.100',
      firstSeen: '2024-01-10T08:30:00Z',
      lastSeen: '2024-01-15T14:20:00Z'
    })

    // 行为分析结果
    const behaviorAnalysis = reactive({
      behaviors: [
        {
          type: 'frequent_requests',
          severity: 'warning',
          description: '检测到频繁请求行为',
          confidence: 0.8
        },
        {
          type: 'suspicious_path',
          severity: 'info',
          description: '访问了一些可疑路径',
          confidence: 0.6
        }
      ],
      recommendations: [
        '建议增加监控频率',
        '考虑启用请求频率限制',
        '建议人工审核访问模式'
      ]
    })

    // 最近访问记录
    const recentLogs = ref([
      {
        ip: '192.168.1.100',
        path: '/api/users',
        method: 'GET',
        action: 'allow',
        score: 75,
        timestamp: '2024-01-15T14:20:00Z'
      },
      {
        ip: '192.168.1.100',
        path: '/admin/config',
        method: 'POST',
        action: 'limit',
        score: 70,
        timestamp: '2024-01-15T14:15:00Z'
      }
    ])

    // 分数趋势图表配置
    const scoreTrendOption = computed(() => ({
      title: {
        text: '分数变化趋势',
        left: 'center'
      },
      tooltip: {
        trigger: 'axis'
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: generateTimeLabels()
      },
      yAxis: {
        type: 'value',
        name: '分数',
        min: 0,
        max: 100
      },
      series: [{
        name: '用户分数',
        type: 'line',
        smooth: true,
        data: generateScoreData(),
        lineStyle: {
          color: '#409eff'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [{
              offset: 0, color: 'rgba(64, 158, 255, 0.3)'
            }, {
              offset: 1, color: 'rgba(64, 158, 255, 0.1)'
            }]
          }
        }
      }]
    }))

    // 生成时间标签
    const generateTimeLabels = () => {
      const labels = []
      const now = new Date()
      
      if (trendRange.value === '24h') {
        for (let i = 23; i >= 0; i--) {
          const time = new Date(now.getTime() - i * 60 * 60 * 1000)
          labels.push(time.getHours() + ':00')
        }
      } else if (trendRange.value === '7d') {
        for (let i = 6; i >= 0; i--) {
          const time = new Date(now.getTime() - i * 24 * 60 * 60 * 1000)
          labels.push((time.getMonth() + 1) + '/' + time.getDate())
        }
      } else {
        for (let i = 29; i >= 0; i--) {
          const time = new Date(now.getTime() - i * 24 * 60 * 60 * 1000)
          labels.push((time.getMonth() + 1) + '/' + time.getDate())
        }
      }
      
      return labels
    }

    // 生成分数数据
    const generateScoreData = () => {
      const data = []
      const count = trendRange.value === '24h' ? 24 : (trendRange.value === '7d' ? 7 : 30)
      let currentScore = userInfo.currentScore
      
      for (let i = 0; i < count; i++) {
        // 模拟分数变化
        const change = Math.random() * 10 - 5
        currentScore = Math.max(0, Math.min(100, currentScore + change))
        data.push(Math.round(currentScore))
      }
      
      return data.reverse()
    }

    // 切换趋势范围
    const changeTrendRange = (range) => {
      trendRange.value = range
    }

    // 刷新行为分析
    const refreshAnalysis = () => {
      analysisLoading.value = true
      setTimeout(() => {
        analysisLoading.value = false
        ElMessage.success('行为分析已更新')
      }, 2000)
    }

    // 查看所有日志
    const viewAllLogs = () => {
      router.push(`/access/logs?fingerprint=${fingerprint}`)
    }

    // 调整分数
    const handleAdjustScore = () => {
      ElMessageBox.prompt(
        '请输入分数调整值（正数为加分，负数为扣分）',
        '调整分数',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputType: 'number',
          inputValidator: (value) => {
            const num = parseInt(value)
            if (isNaN(num) || num < -50 || num > 50) {
              return '请输入-50到50之间的数字'
            }
            return true
          }
        }
      ).then(({ value }) => {
        ElMessage.success(`分数调整成功，调整值: ${value}`)
        userInfo.currentScore += parseInt(value)
      })
    }

    // 重置分数
    const handleResetScore = () => {
      ElMessageBox.confirm(
        '确定要重置用户分数为100分吗？',
        '重置确认',
        { type: 'warning' }
      ).then(() => {
        userInfo.currentScore = 100
        ElMessage.success('用户分数已重置为100分')
      })
    }

    // 加入白名单
    const handleAddWhitelist = () => {
      ElMessageBox.prompt(
        '请输入白名单原因',
        '加入白名单',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputPlaceholder: '如：VIP用户、管理员等'
        }
      ).then(({ value }) => {
        ElMessage.success(`用户已加入白名单，原因: ${value}`)
      })
    }

    // 封禁用户
    const handleBanUser = () => {
      ElMessageBox.prompt(
        '请输入封禁原因',
        '封禁用户',
        {
          confirmButtonText: '确定封禁',
          cancelButtonText: '取消',
          inputPlaceholder: '如：恶意访问、违规行为等'
        }
      ).then(({ value }) => {
        ElMessage.success(`用户已封禁，原因: ${value}`)
      })
    }

    // 辅助函数
    const getScoreClass = (score) => {
      if (score >= 80) return 'score-excellent'
      if (score >= 60) return 'score-good'
      if (score >= 30) return 'score-warning'
      return 'score-danger'
    }

    const getRiskType = (level) => {
      const types = { low: 'success', medium: 'warning', high: 'danger', critical: 'danger' }
      return types[level] || 'info'
    }

    const getRiskText = (level) => {
      const texts = { low: '低', medium: '中', high: '高', critical: '严重' }
      return texts[level] || level
    }

    const getBehaviorType = (severity) => {
      const types = { info: 'info', warning: 'warning', danger: 'danger' }
      return types[severity] || 'info'
    }

    const getMethodType = (method) => {
      const types = { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger' }
      return types[method] || 'info'
    }

    const getActionType = (action) => {
      const types = { allow: 'success', limit: 'warning', challenge: 'info', ban: 'danger' }
      return types[action] || 'info'
    }

    const getActionText = (action) => {
      const texts = { allow: '通过', limit: '限制', challenge: '验证', ban: '封禁' }
      return texts[action] || action
    }

    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString('zh-CN')
    }

    onMounted(() => {
      // 加载用户详细信息
    })

    return {
      userInfo,
      behaviorAnalysis,
      recentLogs,
      analysisLoading,
      trendRange,
      scoreTrendOption,
      changeTrendRange,
      refreshAnalysis,
      viewAllLogs,
      handleAdjustScore,
      handleResetScore,
      handleAddWhitelist,
      handleBanUser,
      getScoreClass,
      getRiskType,
      getRiskText,
      getBehaviorType,
      getMethodType,
      getActionType,
      getActionText,
      formatTime
    }
  }
}
</script>

<style scoped>
.user-detail {
  padding: 0;
}

.info-card,
.chart-card,
.analysis-card,
.logs-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.fingerprint-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  word-break: break-all;
}

.chart-container {
  height: 300px;
}

.analysis-content {
  padding: 20px 0;
}

.analysis-item h4 {
  margin-bottom: 16px;
  color: #303133;
}

.behavior-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 4px;
}

.behavior-desc {
  flex: 1;
  font-size: 14px;
}

.confidence {
  font-size: 12px;
  color: #909399;
}

.recommendations {
  margin: 0;
  padding-left: 20px;
}

.recommendations li {
  margin-bottom: 8px;
  color: #606266;
}

.action-panel {
  text-align: center;
  padding: 20px;
  border-top: 1px solid #ebeef5;
}

.action-panel .el-button {
  margin: 0 8px;
}

.score-excellent { color: #67c23a; font-weight: bold; }
.score-good { color: #409eff; font-weight: bold; }
.score-warning { color: #e6a23c; font-weight: bold; }
.score-danger { color: #f56c6c; font-weight: bold; }
</style>
