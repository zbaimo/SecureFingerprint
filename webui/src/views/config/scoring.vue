<template>
  <div class="scoring-config">
    <div class="page-title">
      <el-icon class="title-icon"><Opportunity /></el-icon>
      评分规则配置
    </div>

    <!-- 评分规则设置 -->
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>评分规则</span>
          <div>
            <el-button @click="resetToDefault">
              <el-icon><RefreshRight /></el-icon>
              重置默认
            </el-button>
            <el-button type="primary" @click="saveScoringConfig" :loading="saving">
              <el-icon><Check /></el-icon>
              保存规则
            </el-button>
          </div>
        </div>
      </template>

      <el-form :model="scoringConfig" label-width="150px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="初始分数">
              <el-input-number
                v-model="scoringConfig.initialScore"
                :min="0"
                :max="200"
                style="width: 100%"
              />
              <div class="form-help">新用户的初始信用分数</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大分数">
              <el-input-number
                v-model="scoringConfig.maxScore"
                :min="100"
                :max="200"
                style="width: 100%"
              />
              <div class="form-help">用户可达到的最高分数</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="正常访问加分">
              <el-input-number
                v-model="scoringConfig.normalAccessBonus"
                :min="0"
                :max="10"
                style="width: 100%"
              />
              <div class="form-help">每次正常访问的加分</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="封禁阈值">
              <el-input-number
                v-model="scoringConfig.banThreshold"
                :min="-50"
                :max="50"
                style="width: 100%"
              />
              <div class="form-help">分数低于此值将被封禁</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">扣分规则</el-divider>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="频繁请求扣分">
              <el-input-number
                v-model="scoringConfig.frequentRequestPenalty"
                :min="-50"
                :max="0"
                style="width: 100%"
              />
              <div class="form-help">短时间内过多请求的扣分</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="可疑UA扣分">
              <el-input-number
                v-model="scoringConfig.suspiciousUAPenalty"
                :min="-50"
                :max="0"
                style="width: 100%"
              />
              <div class="form-help">检测到可疑User-Agent的扣分</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="机器人扣分">
              <el-input-number
                v-model="scoringConfig.botPenalty"
                :min="-50"
                :max="0"
                style="width: 100%"
              />
              <div class="form-help">检测到机器人行为的扣分</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="代理访问扣分">
              <el-input-number
                v-model="scoringConfig.proxyPenalty"
                :min="-20"
                :max="0"
                style="width: 100%"
              />
              <div class="form-help">通过代理访问的扣分</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="路径垃圾扣分">
              <el-input-number
                v-model="scoringConfig.pathSpamPenalty"
                :min="-30"
                :max="0"
                style="width: 100%"
              />
              <div class="form-help">访问可疑路径的扣分</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="无来源扣分">
              <el-input-number
                v-model="scoringConfig.noRefererPenalty"
                :min="-10"
                :max="0"
                style="width: 100%"
              />
              <div class="form-help">缺少Referer头的扣分</div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 评分预览 -->
    <el-card class="config-card">
      <template #header>
        <span>评分规则预览</span>
      </template>

      <div class="score-preview">
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="score-item excellent">
              <div class="score-range">90-100分</div>
              <div class="score-label">优秀用户</div>
              <div class="score-desc">完全信任，无限制</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="score-item good">
              <div class="score-range">70-89分</div>
              <div class="score-label">良好用户</div>
              <div class="score-desc">正常访问，轻度监控</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="score-item warning">
              <div class="score-range">30-69分</div>
              <div class="score-label">警告用户</div>
              <div class="score-desc">限制访问，加强监控</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="score-item danger">
              <div class="score-range">0-29分</div>
              <div class="score-label">危险用户</div>
              <div class="score-desc">严格限制或封禁</div>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Opportunity, Check, Refresh, RefreshRight } from '@element-plus/icons-vue'

export default {
  name: 'ScoringConfig',
  components: {
    Opportunity,
    Check,
    Refresh,
    RefreshRight
  },
  setup() {
    const saving = ref(false)

    // 评分配置
    const scoringConfig = reactive({
      initialScore: 100,
      maxScore: 100,
      normalAccessBonus: 1,
      banThreshold: 0,
      frequentRequestPenalty: -10,
      suspiciousUAPenalty: -20,
      botPenalty: -15,
      proxyPenalty: -5,
      pathSpamPenalty: -8,
      noRefererPenalty: -2
    })

    // 配置历史
    const configHistory = ref([])

    // 详情对话框
    const detailDialog = reactive({
      visible: false,
      data: null
    })

    // 保存评分配置
    const saveScoringConfig = async () => {
      // 验证配置
      if (scoringConfig.banThreshold >= scoringConfig.initialScore) {
        ElMessage.error('封禁阈值不能大于等于初始分数')
        return
      }

      saving.value = true
      try {
        // 这里应该调用API保存配置
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        ElMessage.success('评分规则保存成功')
      } catch (error) {
        ElMessage.error('保存失败: ' + error.message)
      } finally {
        saving.value = false
      }
    }

    // 重置为默认值
    const resetToDefault = () => {
      ElMessageBox.confirm(
        '确定要重置为默认配置吗？这将覆盖当前所有设置。',
        '重置确认',
        {
          type: 'warning',
          confirmButtonText: '确定重置',
          cancelButtonText: '取消'
        }
      ).then(() => {
        Object.assign(scoringConfig, {
          initialScore: 100,
          maxScore: 100,
          normalAccessBonus: 1,
          banThreshold: 0,
          frequentRequestPenalty: -10,
          suspiciousUAPenalty: -20,
          botPenalty: -15,
          proxyPenalty: -5,
          pathSpamPenalty: -8,
          noRefererPenalty: -2
        })
        ElMessage.success('已重置为默认配置')
      })
    }

    // 查看配置详情
    const viewConfigDetails = (row) => {
      detailDialog.data = { ...row }
      detailDialog.visible = true
    }

    // 获取配置类型颜色
    const getConfigTypeColor = (type) => {
      const colors = {
        system: 'primary',
        performance: 'success',
        logging: 'warning',
        security: 'danger'
      }
      return colors[type] || 'info'
    }

    // 获取配置类型名称
    const getConfigTypeName = (type) => {
      const names = {
        system: '系统配置',
        performance: '性能配置',
        logging: '日志配置',
        security: '安全配置'
      }
      return names[type] || type
    }

    // 格式化时间
    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString('zh-CN')
    }

    return {
      saving,
      scoringConfig,
      configHistory,
      detailDialog,
      saveScoringConfig,
      resetToDefault,
      viewConfigDetails,
      getConfigTypeColor,
      getConfigTypeName,
      formatTime
    }
  }
}
</script>

<style scoped>
.scoring-config {
  padding: 0;
}

.config-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-help {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.score-preview {
  padding: 20px 0;
}

.score-item {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  color: white;
  margin-bottom: 10px;
}

.score-item.excellent {
  background: linear-gradient(135deg, #67c23a, #85ce61);
}

.score-item.good {
  background: linear-gradient(135deg, #409eff, #66b1ff);
}

.score-item.warning {
  background: linear-gradient(135deg, #e6a23c, #ebb563);
}

.score-item.danger {
  background: linear-gradient(135deg, #f56c6c, #f78989);
}

.score-range {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 8px;
}

.score-label {
  font-size: 14px;
  margin-bottom: 4px;
}

.score-desc {
  font-size: 12px;
  opacity: 0.9;
}
</style>
