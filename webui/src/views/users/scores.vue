<template>
  <div class="user-scores">
    <div class="page-title">
      <el-icon class="title-icon"><Star /></el-icon>
      用户评分管理
    </div>

    <!-- 分数统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card excellent">
          <div class="stat-number">{{ scoreStats.excellent }}</div>
          <div class="stat-label">优秀用户</div>
          <div class="stat-desc">80-100分</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card good">
          <div class="stat-number">{{ scoreStats.good }}</div>
          <div class="stat-label">良好用户</div>
          <div class="stat-desc">60-79分</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card warning">
          <div class="stat-number">{{ scoreStats.warning }}</div>
          <div class="stat-label">警告用户</div>
          <div class="stat-desc">30-59分</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card danger">
          <div class="stat-number">{{ scoreStats.danger }}</div>
          <div class="stat-label">危险用户</div>
          <div class="stat-desc">0-29分</div>
        </div>
      </el-col>
    </el-row>

    <!-- 分数分布图表 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header">
          <span>分数分布</span>
          <el-button-group size="small">
            <el-button 
              :type="timeRange === '24h' ? 'primary' : ''"
              @click="changeTimeRange('24h')"
            >24小时</el-button>
            <el-button 
              :type="timeRange === '7d' ? 'primary' : ''"
              @click="changeTimeRange('7d')"
            >7天</el-button>
            <el-button 
              :type="timeRange === '30d' ? 'primary' : ''"
              @click="changeTimeRange('30d')"
            >30天</el-button>
          </el-button-group>
        </div>
      </template>
      
      <div class="chart-container">
        <v-chart :option="scoreDistributionOption" />
      </div>
    </el-card>

    <!-- 低分用户列表 -->
    <el-card class="data-table">
      <template #header>
        <div class="card-header">
          <span>低分用户列表</span>
          <div>
            <el-select v-model="scoreThreshold" @change="handleThresholdChange" style="width: 150px">
              <el-option label="分数 < 50" :value="50" />
              <el-option label="分数 < 30" :value="30" />
              <el-option label="分数 < 10" :value="10" />
            </el-select>
            <el-button type="primary" @click="handleRefresh" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="lowScoreUsers" v-loading="loading" stripe>
        <el-table-column prop="fingerprint" label="用户指纹" width="140">
          <template #default="scope">
            <el-tooltip :content="scope.row.fingerprint" placement="top">
              <el-button
                type="text"
                @click="handleViewUser(scope.row.fingerprint)"
                class="fingerprint-link"
              >
                {{ scope.row.fingerprint.slice(0, 8) }}...
              </el-button>
            </el-tooltip>
          </template>
        </el-table-column>
        
        <el-table-column prop="currentScore" label="当前分数" width="100">
          <template #default="scope">
            <span :class="getScoreClass(scope.row.currentScore)">
              {{ scope.row.currentScore }}
            </span>
          </template>
        </el-table-column>
        
        <el-table-column prop="riskLevel" label="风险等级" width="100">
          <template #default="scope">
            <el-tag :type="getRiskType(scope.row.riskLevel)" size="small">
              {{ getRiskText(scope.row.riskLevel) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="totalRequests" label="总请求数" width="100" />
        
        <el-table-column prop="lastSeen" label="最后访问" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.lastSeen) }}
          </template>
        </el-table-column>
        
        <el-table-column label="分数趋势" width="150">
          <template #default="scope">
            <div class="trend-indicator">
              <el-icon v-if="scope.row.trend === 'up'" color="#67c23a"><CaretTop /></el-icon>
              <el-icon v-else-if="scope.row.trend === 'down'" color="#f56c6c"><CaretBottom /></el-icon>
              <el-icon v-else color="#909399"><Minus /></el-icon>
              <span :class="getTrendClass(scope.row.trend)">
                {{ getTrendText(scope.row.trend) }}
              </span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <div class="action-buttons">
              <el-button
                type="text"
                size="small"
                @click="handleViewUser(scope.row.fingerprint)"
              >
                详情
              </el-button>
              <el-button
                type="text"
                size="small"
                @click="handleAdjustScore(scope.row)"
              >
                调分
              </el-button>
              <el-button
                type="text"
                size="small"
                @click="handleResetScore(scope.row)"
              >
                重置
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 调分对话框 -->
    <el-dialog v-model="adjustDialog.visible" title="调整用户分数" width="400px">
      <el-form :model="adjustDialog.form" label-width="100px">
        <el-form-item label="用户指纹">
          <el-text class="fingerprint-text">{{ adjustDialog.form.fingerprint }}</el-text>
        </el-form-item>
        <el-form-item label="当前分数">
          <el-text :class="getScoreClass(adjustDialog.form.currentScore)">
            {{ adjustDialog.form.currentScore }}
          </el-text>
        </el-form-item>
        <el-form-item label="分数调整">
          <el-input-number
            v-model="adjustDialog.form.adjustment"
            :min="-50"
            :max="50"
            style="width: 100%"
          />
          <div class="form-help">正数为加分，负数为扣分</div>
        </el-form-item>
        <el-form-item label="调整原因">
          <el-input
            v-model="adjustDialog.form.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入调整原因"
          />
        </el-form-item>
        <el-form-item label="预期分数">
          <el-text :class="getScoreClass(adjustDialog.form.currentScore + adjustDialog.form.adjustment)">
            {{ adjustDialog.form.currentScore + adjustDialog.form.adjustment }}
          </el-text>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="adjustDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="confirmAdjustScore" :loading="saving">
          确定调整
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Star, Refresh, CaretTop, CaretBottom, Minus } from '@element-plus/icons-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, BarChart, TitleComponent, TooltipComponent, GridComponent])

export default {
  name: 'UserScores',
  components: {
    Star,
    Refresh,
    CaretTop,
    CaretBottom,
    Minus,
    VChart
  },
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const saving = ref(false)
    const timeRange = ref('24h')
    const scoreThreshold = ref(30)

    // 分数统计
    const scoreStats = reactive({
      excellent: 1200,
      good: 800,
      warning: 300,
      danger: 50
    })

    // 低分用户列表
    const lowScoreUsers = ref([
      {
        fingerprint: 'abc123def456ghi789jkl012',
        currentScore: 25,
        riskLevel: 'high',
        totalRequests: 150,
        lastSeen: '2024-01-15T14:20:00Z',
        trend: 'down'
      },
      {
        fingerprint: 'xyz789uvw012rst345mno678',
        currentScore: 15,
        riskLevel: 'critical',
        totalRequests: 200,
        lastSeen: '2024-01-15T13:45:00Z',
        trend: 'down'
      },
      {
        fingerprint: 'def456ghi789jkl012mno345',
        currentScore: 28,
        riskLevel: 'high',
        totalRequests: 89,
        lastSeen: '2024-01-15T12:10:00Z',
        trend: 'up'
      }
    ])

    // 调分对话框
    const adjustDialog = reactive({
      visible: false,
      form: {
        fingerprint: '',
        currentScore: 0,
        adjustment: 0,
        reason: ''
      }
    })

    // 分数分布图表配置
    const scoreDistributionOption = computed(() => ({
      title: {
        text: '用户分数分布',
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
        data: ['0-9', '10-19', '20-29', '30-39', '40-49', '50-59', '60-69', '70-79', '80-89', '90-100']
      },
      yAxis: {
        type: 'value',
        name: '用户数量'
      },
      series: [{
        name: '用户数量',
        type: 'bar',
        data: [15, 20, 35, 80, 120, 200, 300, 500, 600, 600],
        itemStyle: {
          color: function(params) {
            const colors = ['#f56c6c', '#f56c6c', '#f56c6c', '#e6a23c', '#e6a23c', '#e6a23c', '#409eff', '#409eff', '#67c23a', '#67c23a']
            return colors[params.dataIndex]
          }
        }
      }]
    }))

    // 切换时间范围
    const changeTimeRange = (range) => {
      timeRange.value = range
      // 重新加载数据
    }

    // 阈值变化
    const handleThresholdChange = () => {
      handleRefresh()
    }

    // 刷新数据
    const handleRefresh = () => {
      loading.value = true
      setTimeout(() => {
        loading.value = false
      }, 1000)
    }

    // 查看用户详情
    const handleViewUser = (fingerprint) => {
      router.push(`/users/detail/${fingerprint}`)
    }

    // 调整分数
    const handleAdjustScore = (user) => {
      adjustDialog.form.fingerprint = user.fingerprint
      adjustDialog.form.currentScore = user.currentScore
      adjustDialog.form.adjustment = 0
      adjustDialog.form.reason = ''
      adjustDialog.visible = true
    }

    // 确认调整分数
    const confirmAdjustScore = async () => {
      if (!adjustDialog.form.reason) {
        ElMessage.warning('请输入调整原因')
        return
      }

      saving.value = true
      try {
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        ElMessage.success('分数调整成功')
        adjustDialog.visible = false
        handleRefresh()
      } catch (error) {
        ElMessage.error('调整失败')
      } finally {
        saving.value = false
      }
    }

    // 重置分数
    const handleResetScore = (user) => {
      ElMessageBox.confirm(
        `确定要重置用户 ${user.fingerprint.slice(0, 8)}... 的分数为100分吗？`,
        '重置确认',
        {
          type: 'warning'
        }
      ).then(() => {
        ElMessage.success('用户分数已重置为100分')
        handleRefresh()
      })
    }

    // 获取分数样式
    const getScoreClass = (score) => {
      if (score >= 80) return 'score-excellent'
      if (score >= 60) return 'score-good'
      if (score >= 30) return 'score-warning'
      return 'score-danger'
    }

    // 获取风险类型
    const getRiskType = (level) => {
      const types = {
        low: 'success',
        medium: 'warning',
        high: 'danger',
        critical: 'danger'
      }
      return types[level] || 'info'
    }

    // 获取风险文本
    const getRiskText = (level) => {
      const texts = {
        low: '低',
        medium: '中',
        high: '高',
        critical: '严重'
      }
      return texts[level] || level
    }

    // 获取趋势样式
    const getTrendClass = (trend) => {
      return {
        'trend-up': trend === 'up',
        'trend-down': trend === 'down',
        'trend-stable': trend === 'stable'
      }
    }

    // 获取趋势文本
    const getTrendText = (trend) => {
      const texts = {
        up: '上升',
        down: '下降',
        stable: '稳定'
      }
      return texts[trend] || '未知'
    }

    // 格式化时间
    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString('zh-CN')
    }

    onMounted(() => {
      handleRefresh()
    })

    return {
      loading,
      saving,
      timeRange,
      scoreThreshold,
      scoreStats,
      lowScoreUsers,
      adjustDialog,
      scoreDistributionOption,
      changeTimeRange,
      handleThresholdChange,
      handleRefresh,
      handleViewUser,
      handleAdjustScore,
      confirmAdjustScore,
      handleResetScore,
      getScoreClass,
      getRiskType,
      getRiskText,
      getTrendClass,
      getTrendText,
      formatTime
    }
  }
}
</script>

<style scoped>
.user-scores {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.chart-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.fingerprint-link {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.fingerprint-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.form-help {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.trend-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
}

.trend-up {
  color: #67c23a;
}

.trend-down {
  color: #f56c6c;
}

.trend-stable {
  color: #909399;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.score-excellent {
  color: #67c23a;
  font-weight: bold;
}

.score-good {
  color: #409eff;
  font-weight: bold;
}

.score-warning {
  color: #e6a23c;
  font-weight: bold;
}

.score-danger {
  color: #f56c6c;
  font-weight: bold;
}

.stat-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  color: white;
  margin-bottom: 10px;
}

.stat-card.excellent {
  background: linear-gradient(135deg, #67c23a, #85ce61);
}

.stat-card.good {
  background: linear-gradient(135deg, #409eff, #66b1ff);
}

.stat-card.warning {
  background: linear-gradient(135deg, #e6a23c, #ebb563);
}

.stat-card.danger {
  background: linear-gradient(135deg, #f56c6c, #f78989);
}

.stat-number {
  font-size: 2.5em;
  font-weight: bold;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 1.1em;
  margin-bottom: 4px;
}

.stat-desc {
  font-size: 0.9em;
  opacity: 0.9;
}

.chart-container {
  height: 400px;
}
</style>
