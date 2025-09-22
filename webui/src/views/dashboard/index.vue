<template>
  <div class="dashboard">
    <div class="page-title">
      <el-icon class="title-icon"><Monitor /></el-icon>
      系统仪表板
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card">
          <div class="stat-number">{{ formatNumber(stats.totalRequests) }}</div>
          <div class="stat-label">总访问量</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card success">
          <div class="stat-number">{{ formatNumber(stats.todayRequests) }}</div>
          <div class="stat-label">今日访问</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card warning">
          <div class="stat-number">{{ formatNumber(stats.blockedRequests) }}</div>
          <div class="stat-label">拦截请求</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card danger">
          <div class="stat-number">{{ stats.bannedUsers }}</div>
          <div class="stat-label">封禁用户</div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <el-card class="dashboard-card">
          <template #header>
            <div class="card-header">
              <span>访问趋势</span>
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
            <v-chart :option="accessTrendOption" />
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <el-card class="dashboard-card">
          <template #header>
            <span>请求状态分布</span>
          </template>
          <div class="chart-container">
            <v-chart :option="statusDistributionOption" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 数据表格区域 -->
    <el-row :gutter="20" class="tables-row">
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <el-card class="dashboard-card">
          <template #header>
            <div class="card-header">
              <span>最近访问记录</span>
              <el-button type="text" @click="$router.push('/access/logs')">
                查看更多 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table :data="recentLogs" size="small">
            <el-table-column prop="ip" label="IP地址" width="120" />
            <el-table-column prop="path" label="访问路径" show-overflow-tooltip />
            <el-table-column prop="action" label="状态" width="80">
              <template #default="scope">
                <el-tag 
                  :type="getActionType(scope.row.action)" 
                  size="small"
                >
                  {{ getActionText(scope.row.action) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="timestamp" label="时间" width="120">
              <template #default="scope">
                {{ formatTime(scope.row.timestamp) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <el-card class="dashboard-card">
          <template #header>
            <div class="card-header">
              <span>高风险用户</span>
              <el-button type="text" @click="$router.push('/users/scores')">
                查看更多 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table :data="riskUsers" size="small">
            <el-table-column prop="fingerprint" label="用户指纹" width="120">
              <template #default="scope">
                <span class="fingerprint">{{ scope.row.fingerprint.slice(0, 8) }}...</span>
              </template>
            </el-table-column>
            <el-table-column prop="score" label="分数" width="60">
              <template #default="scope">
                <span :class="getScoreClass(scope.row.score)">{{ scope.row.score }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="riskLevel" label="风险等级" width="80">
              <template #default="scope">
                <el-tag 
                  :type="getRiskType(scope.row.riskLevel)" 
                  size="small"
                >
                  {{ getRiskText(scope.row.riskLevel) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="lastSeen" label="最后访问" width="100">
              <template #default="scope">
                {{ formatTime(scope.row.lastSeen) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed } from 'vue'
import { Monitor, ArrowRight } from '@element-plus/icons-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer,
  LineChart,
  PieChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

export default {
  name: 'Dashboard',
  components: {
    VChart,
    Monitor,
    ArrowRight
  },
  setup() {
    const timeRange = ref('24h')
    const loading = ref(false)
    
    // 统计数据
    const stats = reactive({
      totalRequests: 156789,
      todayRequests: 12456,
      blockedRequests: 1234,
      bannedUsers: 56
    })

    // 最近访问记录
    const recentLogs = ref([
      {
        ip: '192.168.1.100',
        path: '/api/users',
        action: 'allow',
        timestamp: new Date(Date.now() - 5 * 60 * 1000)
      },
      {
        ip: '10.0.0.50',
        path: '/admin/config',
        action: 'limit',
        timestamp: new Date(Date.now() - 10 * 60 * 1000)
      },
      {
        ip: '203.0.113.25',
        path: '/wp-admin',
        action: 'ban',
        timestamp: new Date(Date.now() - 15 * 60 * 1000)
      },
      {
        ip: '198.51.100.75',
        path: '/api/login',
        action: 'challenge',
        timestamp: new Date(Date.now() - 20 * 60 * 1000)
      }
    ])

    // 高风险用户
    const riskUsers = ref([
      {
        fingerprint: 'abc123def456ghi789',
        score: 15,
        riskLevel: 'high',
        lastSeen: new Date(Date.now() - 30 * 60 * 1000)
      },
      {
        fingerprint: 'xyz789uvw012rst345',
        score: 25,
        riskLevel: 'medium',
        lastSeen: new Date(Date.now() - 45 * 60 * 1000)
      },
      {
        fingerprint: 'mno345pqr678stu901',
        score: 5,
        riskLevel: 'critical',
        lastSeen: new Date(Date.now() - 60 * 60 * 1000)
      }
    ])

    // 访问趋势图表配置
    const accessTrendOption = computed(() => ({
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross'
        }
      },
      legend: {
        data: ['正常访问', '被拦截', '被封禁']
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: generateTimeLabels()
      },
      yAxis: {
        type: 'value'
      },
      series: [
        {
          name: '正常访问',
          type: 'line',
          stack: 'Total',
          smooth: true,
          data: generateTrendData(800, 1200)
        },
        {
          name: '被拦截',
          type: 'line',
          stack: 'Total',
          smooth: true,
          data: generateTrendData(50, 150)
        },
        {
          name: '被封禁',
          type: 'line',
          stack: 'Total',
          smooth: true,
          data: generateTrendData(5, 25)
        }
      ]
    }))

    // 状态分布饼图配置
    const statusDistributionOption = computed(() => ({
      tooltip: {
        trigger: 'item'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      series: [
        {
          name: '请求状态',
          type: 'pie',
          radius: '50%',
          data: [
            { value: 11200, name: '正常通过' },
            { value: 856, name: '限速处理' },
            { value: 234, name: '人机验证' },
            { value: 166, name: '直接封禁' }
          ],
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }
      ]
    }))

    // 生成时间标签
    const generateTimeLabels = () => {
      const labels = []
      const now = new Date()
      
      if (timeRange.value === '24h') {
        for (let i = 23; i >= 0; i--) {
          const time = new Date(now.getTime() - i * 60 * 60 * 1000)
          labels.push(time.getHours() + ':00')
        }
      } else if (timeRange.value === '7d') {
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

    // 生成趋势数据
    const generateTrendData = (min, max) => {
      const data = []
      const count = timeRange.value === '24h' ? 24 : (timeRange.value === '7d' ? 7 : 30)
      
      for (let i = 0; i < count; i++) {
        data.push(Math.floor(Math.random() * (max - min + 1)) + min)
      }
      
      return data
    }

    // 切换时间范围
    const changeTimeRange = (range) => {
      timeRange.value = range
    }

    // 格式化数字
    const formatNumber = (num) => {
      if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K'
      }
      return num.toString()
    }

    // 格式化时间
    const formatTime = (timestamp) => {
      const now = new Date()
      const time = new Date(timestamp)
      const diff = now - time
      
      if (diff < 60 * 1000) {
        return '刚刚'
      } else if (diff < 60 * 60 * 1000) {
        return Math.floor(diff / (60 * 1000)) + '分钟前'
      } else if (diff < 24 * 60 * 60 * 1000) {
        return Math.floor(diff / (60 * 60 * 1000)) + '小时前'
      } else {
        return time.toLocaleDateString()
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

    // 获取分数样式类
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

    // 加载数据
    const loadData = async () => {
      loading.value = true
      try {
        // 这里应该调用实际的API
        await new Promise(resolve => setTimeout(resolve, 1000))
      } catch (error) {
        console.error('加载数据失败:', error)
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      loadData()
    })

    return {
      timeRange,
      loading,
      stats,
      recentLogs,
      riskUsers,
      accessTrendOption,
      statusDistributionOption,
      changeTimeRange,
      formatNumber,
      formatTime,
      getActionType,
      getActionText,
      getScoreClass,
      getRiskType,
      getRiskText
    }
  }
}
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.charts-row {
  margin-bottom: 20px;
}

.tables-row {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.fingerprint {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  color: #666;
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

@media (max-width: 768px) {
  .stats-row .el-col {
    margin-bottom: 10px;
  }
  
  .charts-row .el-col,
  .tables-row .el-col {
    margin-bottom: 20px;
  }
}
</style>
