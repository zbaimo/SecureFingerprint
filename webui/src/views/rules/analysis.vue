<template>
  <div class="behavior-analysis">
    <div class="page-title">
      <el-icon class="title-icon"><TrendCharts /></el-icon>
      行为分析
    </div>

    <!-- 分析概览 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card low">
          <div class="stat-number">{{ analysisStats.lowRisk }}</div>
          <div class="stat-label">低风险用户</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card medium">
          <div class="stat-number">{{ analysisStats.mediumRisk }}</div>
          <div class="stat-label">中风险用户</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card high">
          <div class="stat-number">{{ analysisStats.highRisk }}</div>
          <div class="stat-label">高风险用户</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card critical">
          <div class="stat-number">{{ analysisStats.criticalRisk }}</div>
          <div class="stat-label">严重风险用户</div>
        </div>
      </el-col>
    </el-row>

    <!-- 行为检测图表 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>行为类型分布</span>
          </template>
          <div class="chart-container">
            <v-chart :option="behaviorDistributionOption" />
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>风险趋势</span>
          </template>
          <div class="chart-container">
            <v-chart :option="riskTrendOption" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 实时分析 -->
    <el-card class="analysis-card">
      <template #header>
        <div class="card-header">
          <span>实时行为分析</span>
          <el-button @click="handleAnalyzeUser">
            <el-icon><Search /></el-icon>
            分析指定用户
          </el-button>
        </div>
      </template>

      <el-table :data="realtimeAnalysis" size="small">
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
        
        <el-table-column prop="riskScore" label="风险分数" width="100">
          <template #default="scope">
            <span :class="getRiskScoreClass(scope.row.riskScore)">
              {{ scope.row.riskScore.toFixed(1) }}
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
        
        <el-table-column prop="behaviors" label="检测到的行为" min-width="200">
          <template #default="scope">
            <div class="behaviors-list">
              <el-tag
                v-for="behavior in scope.row.behaviors"
                :key="behavior.type"
                :type="getBehaviorType(behavior.severity)"
                size="small"
                class="behavior-tag"
              >
                {{ getBehaviorText(behavior.type) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="recommendations" label="建议操作" min-width="200">
          <template #default="scope">
            <div class="recommendations-list">
              <div
                v-for="(rec, index) in scope.row.recommendations.slice(0, 2)"
                :key="index"
                class="recommendation"
              >
                {{ rec }}
              </div>
              <el-button
                v-if="scope.row.recommendations.length > 2"
                type="text"
                size="small"
                @click="showAllRecommendations(scope.row)"
              >
                查看更多...
              </el-button>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="timestamp" label="分析时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button
              type="text"
              size="small"
              @click="handleViewAnalysisDetail(scope.row)"
            >
              <el-icon><View /></el-icon>
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 分析详情对话框 -->
    <el-dialog
      v-model="detailDialog.visible"
      title="行为分析详情"
      width="800px"
    >
      <div v-if="detailDialog.data">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户指纹">{{ detailDialog.data.fingerprint }}</el-descriptions-item>
          <el-descriptions-item label="风险分数">
            <span :class="getRiskScoreClass(detailDialog.data.riskScore)">
              {{ detailDialog.data.riskScore.toFixed(1) }}
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <el-tag :type="getRiskType(detailDialog.data.riskLevel)" size="small">
              {{ getRiskText(detailDialog.data.riskLevel) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="分析时间">{{ formatTime(detailDialog.data.timestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">检测到的行为</el-divider>
        <div class="behaviors-detail">
          <div
            v-for="behavior in detailDialog.data.behaviors"
            :key="behavior.type"
            class="behavior-detail-item"
          >
            <div class="behavior-header">
              <el-tag :type="getBehaviorType(behavior.severity)" size="small">
                {{ getBehaviorText(behavior.type) }}
              </el-tag>
              <span class="confidence">置信度: {{ (behavior.confidence * 100).toFixed(1) }}%</span>
            </div>
            <div class="behavior-desc">{{ behavior.description }}</div>
            <div v-if="behavior.evidence && behavior.evidence.length > 0" class="evidence">
              <strong>证据:</strong>
              <ul>
                <li v-for="(evidence, index) in behavior.evidence" :key="index">
                  {{ evidence }}
                </li>
              </ul>
            </div>
          </div>
        </div>

        <el-divider content-position="left">建议操作</el-divider>
        <ul class="recommendations-detail">
          <li v-for="(rec, index) in detailDialog.data.recommendations" :key="index">
            {{ rec }}
          </li>
        </ul>
      </div>
      
      <template #footer>
        <el-button @click="detailDialog.visible = false">关闭</el-button>
        <el-button type="primary" @click="handleViewUser(detailDialog.data.fingerprint)">
          查看用户详情
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { TrendCharts, Search, Refresh, View } from '@element-plus/icons-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, LineChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, PieChart, LineChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])

export default {
  name: 'BehaviorAnalysis',
  components: {
    TrendCharts,
    Search,
    Refresh,
    View,
    VChart
  },
  setup() {
    const router = useRouter()
    const loading = ref(false)

    // 分析统计
    const analysisStats = reactive({
      lowRisk: 1200,
      mediumRisk: 300,
      highRisk: 80,
      criticalRisk: 20
    })

    // 实时分析数据
    const realtimeAnalysis = ref([
      {
        fingerprint: 'abc123def456ghi789jkl012',
        riskScore: 75.5,
        riskLevel: 'medium',
        behaviors: [
          { type: 'frequent_requests', severity: 'warning', confidence: 0.8 },
          { type: 'suspicious_path', severity: 'info', confidence: 0.6 }
        ],
        recommendations: ['建议增加监控频率', '考虑启用请求限制'],
        timestamp: '2024-01-15T14:20:00Z'
      },
      {
        fingerprint: 'xyz789uvw012rst345mno678',
        riskScore: 85.2,
        riskLevel: 'high',
        behaviors: [
          { type: 'bot_behavior', severity: 'danger', confidence: 0.9 },
          { type: 'scanning_behavior', severity: 'danger', confidence: 0.7 }
        ],
        recommendations: ['建议立即封禁', '启用机器人验证'],
        timestamp: '2024-01-15T14:15:00Z'
      }
    ])

    // 详情对话框
    const detailDialog = reactive({
      visible: false,
      data: null
    })

    // 行为分布图表配置
    const behaviorDistributionOption = computed(() => ({
      title: {
        text: '行为类型分布',
        left: 'center'
      },
      tooltip: {
        trigger: 'item'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      series: [{
        name: '行为类型',
        type: 'pie',
        radius: '50%',
        data: [
          { value: 40, name: '频繁请求' },
          { value: 25, name: '机器人行为' },
          { value: 20, name: '路径扫描' },
          { value: 10, name: '可疑UA' },
          { value: 5, name: '其他' }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }]
    }))

    // 风险趋势图表配置
    const riskTrendOption = computed(() => ({
      title: {
        text: '风险趋势',
        left: 'center'
      },
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['高风险', '中风险', '低风险']
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
      },
      yAxis: {
        type: 'value'
      },
      series: [
        {
          name: '高风险',
          type: 'line',
          smooth: true,
          data: [5, 8, 12, 15, 10, 7],
          lineStyle: { color: '#f56c6c' }
        },
        {
          name: '中风险',
          type: 'line',
          smooth: true,
          data: [20, 25, 30, 35, 28, 22],
          lineStyle: { color: '#e6a23c' }
        },
        {
          name: '低风险',
          type: 'line',
          smooth: true,
          data: [100, 120, 150, 180, 160, 140],
          lineStyle: { color: '#67c23a' }
        }
      ]
    }))

    // 分析指定用户
    const handleAnalyzeUser = () => {
      ElMessageBox.prompt(
        '请输入要分析的用户指纹',
        '用户行为分析',
        {
          confirmButtonText: '开始分析',
          cancelButtonText: '取消',
          inputPlaceholder: '请输入用户指纹'
        }
      ).then(({ value }) => {
        loading.value = true
        setTimeout(() => {
          loading.value = false
          ElMessage.success(`用户 ${value.slice(0, 8)}... 分析完成`)
          
          // 添加到实时分析列表
          realtimeAnalysis.value.unshift({
            fingerprint: value,
            riskScore: Math.random() * 100,
            riskLevel: ['low', 'medium', 'high'][Math.floor(Math.random() * 3)],
            behaviors: [
              { type: 'normal_access', severity: 'info', confidence: 0.9 }
            ],
            recommendations: ['用户行为正常，继续监控'],
            timestamp: new Date().toISOString()
          })
        }, 2000)
      })
    }

    // 查看用户详情
    const handleViewUser = (fingerprint) => {
      router.push(`/users/detail/${fingerprint}`)
    }

    // 查看分析详情
    const handleViewAnalysisDetail = (row) => {
      detailDialog.data = { ...row }
      detailDialog.visible = true
    }

    // 显示所有建议
    const showAllRecommendations = (row) => {
      ElMessageBox.alert(
        row.recommendations.join('\n'),
        '完整建议列表',
        {
          confirmButtonText: '确定'
        }
      )
    }

    // 刷新数据
    const handleRefresh = () => {
      loading.value = true
      setTimeout(() => {
        loading.value = false
        ElMessage.success('分析数据已刷新')
      }, 1000)
    }

    // 辅助函数
    const getRiskType = (level) => {
      const types = { low: 'success', medium: 'warning', high: 'danger', critical: 'danger' }
      return types[level] || 'info'
    }

    const getRiskText = (level) => {
      const texts = { low: '低', medium: '中', high: '高', critical: '严重' }
      return texts[level] || level
    }

    const getRiskScoreClass = (score) => {
      if (score >= 80) return 'risk-critical'
      if (score >= 60) return 'risk-high'
      if (score >= 30) return 'risk-medium'
      return 'risk-low'
    }

    const getBehaviorType = (severity) => {
      const types = { info: 'info', warning: 'warning', danger: 'danger' }
      return types[severity] || 'info'
    }

    const getBehaviorText = (type) => {
      const texts = {
        frequent_requests: '频繁请求',
        bot_behavior: '机器人行为',
        suspicious_path: '可疑路径',
        scanning_behavior: '扫描行为',
        normal_access: '正常访问'
      }
      return texts[type] || type
    }

    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString('zh-CN')
    }

    onMounted(() => {
      handleRefresh()
    })

    return {
      loading,
      analysisStats,
      realtimeAnalysis,
      detailDialog,
      behaviorDistributionOption,
      riskTrendOption,
      handleAnalyzeUser,
      handleViewUser,
      handleViewAnalysisDetail,
      showAllRecommendations,
      handleRefresh,
      getRiskType,
      getRiskText,
      getRiskScoreClass,
      getBehaviorType,
      getBehaviorText,
      formatTime
    }
  }
}
</script>

<style scoped>
.behavior-analysis {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.chart-card,
.analysis-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  height: 300px;
}

.fingerprint-link {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.behaviors-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.behavior-tag {
  margin: 0;
}

.recommendations-list {
  font-size: 12px;
}

.recommendation {
  margin-bottom: 4px;
  color: #606266;
}

.behaviors-detail {
  margin: 16px 0;
}

.behavior-detail-item {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

.behavior-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.confidence {
  font-size: 12px;
  color: #909399;
}

.behavior-desc {
  color: #606266;
  margin-bottom: 8px;
}

.evidence {
  font-size: 12px;
  color: #909399;
}

.evidence ul {
  margin: 4px 0 0 16px;
}

.recommendations-detail {
  margin: 0;
  padding-left: 20px;
}

.recommendations-detail li {
  margin-bottom: 8px;
  color: #606266;
}

.stat-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  color: white;
  margin-bottom: 10px;
}

.stat-card.low {
  background: linear-gradient(135deg, #67c23a, #85ce61);
}

.stat-card.medium {
  background: linear-gradient(135deg, #e6a23c, #ebb563);
}

.stat-card.high {
  background: linear-gradient(135deg, #f56c6c, #f78989);
}

.stat-card.critical {
  background: linear-gradient(135deg, #909399, #a6a9ad);
}

.stat-number {
  font-size: 2.5em;
  font-weight: bold;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 1.1em;
}

.risk-low { color: #67c23a; font-weight: bold; }
.risk-medium { color: #e6a23c; font-weight: bold; }
.risk-high { color: #f56c6c; font-weight: bold; }
.risk-critical { color: #909399; font-weight: bold; }
</style>
