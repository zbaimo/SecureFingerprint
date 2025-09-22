<template>
  <div class="security-config">
    <div class="page-title">
      <el-icon class="title-icon"><Lock /></el-icon>
      安全配置
    </div>

    <!-- 限制器配置 -->
    <el-card class="config-card">
      <template #header>
        <span>访问限制配置</span>
      </template>

      <el-form :model="limiterConfig" label-width="150px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="限速时间窗口(秒)">
              <el-input-number
                v-model="limiterConfig.rateLimitWindow"
                :min="10"
                :max="3600"
                style="width: 100%"
              />
              <div class="form-help">限速统计的时间窗口</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="窗口最大请求数">
              <el-input-number
                v-model="limiterConfig.maxRequestsPerWindow"
                :min="10"
                :max="1000"
                style="width: 100%"
              />
              <div class="form-help">时间窗口内允许的最大请求数</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="封禁时长(秒)">
              <el-input-number
                v-model="limiterConfig.banDuration"
                :min="60"
                :max="86400"
                style="width: 100%"
              />
              <div class="form-help">自动封禁的持续时间</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="延迟响应(毫秒)">
              <el-input-number
                v-model="limiterConfig.delayResponseMs"
                :min="0"
                :max="10000"
                style="width: 100%"
              />
              <div class="form-help">限速时的响应延迟</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="警告阈值">
              <el-input-number
                v-model="limiterConfig.warningThreshold"
                :min="0"
                :max="100"
                style="width: 100%"
              />
              <div class="form-help">分数低于此值时警告</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="严重阈值">
              <el-input-number
                v-model="limiterConfig.criticalThreshold"
                :min="0"
                :max="50"
                style="width: 100%"
              />
              <div class="form-help">分数低于此值时严格限制</div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 行为分析配置 -->
    <el-card class="config-card">
      <template #header>
        <span>行为分析配置</span>
      </template>

      <el-form :model="analyzerConfig" label-width="150px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="可疑请求阈值">
              <el-input-number
                v-model="analyzerConfig.suspiciousRequestThreshold"
                :min="10"
                :max="200"
                style="width: 100%"
              />
              <div class="form-help">判定为可疑行为的请求数阈值</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="路径重复阈值">
              <el-input-number
                v-model="analyzerConfig.pathRepeatThreshold"
                :min="5"
                :max="100"
                style="width: 100%"
              />
              <div class="form-help">同一路径重复访问的阈值</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="启用机器人检测">
              <el-switch v-model="analyzerConfig.botDetectionEnabled" />
              <div class="form-help">是否启用自动机器人检测</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="启用模式检测">
              <el-switch v-model="analyzerConfig.patternDetectionEnabled" />
              <div class="form-help">是否启用访问模式检测</div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="分析时间窗口(秒)">
              <el-input-number
                v-model="analyzerConfig.analysisWindow"
                :min="300"
                :max="7200"
                style="width: 100%"
              />
              <div class="form-help">行为分析的时间窗口</div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 代理配置 -->
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>代理信任配置</span>
          <el-button @click="addTrustedProxy">
            <el-icon><Plus /></el-icon>
            添加可信代理
          </el-button>
        </div>
      </template>

      <el-form :model="proxyConfig" label-width="150px">
        <el-form-item label="可信代理列表">
          <div class="proxy-list">
            <div
              v-for="(proxy, index) in proxyConfig.trustedProxies"
              :key="index"
              class="proxy-item"
            >
              <el-input v-model="proxyConfig.trustedProxies[index]" placeholder="IP/CIDR" />
              <el-button
                type="danger"
                text
                @click="removeTrustedProxy(index)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="可信头列表">
          <el-select
            v-model="proxyConfig.trustedHeaders"
            multiple
            style="width: 100%"
            placeholder="选择可信的代理头"
          >
            <el-option label="X-Real-IP" value="X-Real-IP" />
            <el-option label="X-Forwarded-For" value="X-Forwarded-For" />
            <el-option label="CF-Connecting-IP" value="CF-Connecting-IP" />
            <el-option label="True-Client-IP" value="True-Client-IP" />
            <el-option label="X-Client-IP" value="X-Client-IP" />
          </el-select>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="跳过内网IP">
              <el-switch v-model="proxyConfig.skipPrivateRanges" />
              <div class="form-help">是否跳过内网IP段</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大代理深度">
              <el-input-number
                v-model="proxyConfig.maxProxyDepth"
                :min="1"
                :max="20"
                style="width: 100%"
              />
              <div class="form-help">代理链的最大深度</div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 保存按钮 -->
    <div class="save-actions">
      <el-button size="large" @click="resetAllConfig">
        <el-icon><RefreshRight /></el-icon>
        重置所有配置
      </el-button>
      <el-button type="primary" size="large" @click="saveAllConfig" :loading="saving">
        <el-icon><Check /></el-icon>
        保存所有配置
      </el-button>
    </div>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock, Check, RefreshRight, Plus, Delete } from '@element-plus/icons-vue'

export default {
  name: 'SecurityConfig',
  components: {
    Lock,
    Check,
    RefreshRight,
    Plus,
    Delete
  },
  setup() {
    const saving = ref(false)

    // 限制器配置
    const limiterConfig = reactive({
      rateLimitWindow: 60,
      maxRequestsPerWindow: 100,
      banDuration: 3600,
      delayResponseMs: 1000,
      warningThreshold: 30,
      criticalThreshold: 10
    })

    // 分析器配置
    const analyzerConfig = reactive({
      suspiciousRequestThreshold: 50,
      pathRepeatThreshold: 10,
      botDetectionEnabled: true,
      patternDetectionEnabled: true,
      analysisWindow: 3600
    })

    // 代理配置
    const proxyConfig = reactive({
      trustedProxies: [
        '127.0.0.1/32',
        '10.0.0.0/8',
        '172.16.0.0/12',
        '192.168.0.0/16'
      ],
      trustedHeaders: [
        'X-Real-IP',
        'X-Forwarded-For',
        'CF-Connecting-IP'
      ],
      skipPrivateRanges: true,
      maxProxyDepth: 10
    })

    // 添加可信代理
    const addTrustedProxy = () => {
      proxyConfig.trustedProxies.push('')
    }

    // 删除可信代理
    const removeTrustedProxy = (index) => {
      proxyConfig.trustedProxies.splice(index, 1)
    }

    // 保存所有配置
    const saveAllConfig = async () => {
      saving.value = true
      try {
        // 这里应该调用API保存所有配置
        await new Promise(resolve => setTimeout(resolve, 1500))
        
        ElMessage.success('所有配置保存成功')
      } catch (error) {
        ElMessage.error('保存失败: ' + error.message)
      } finally {
        saving.value = false
      }
    }

    // 重置所有配置
    const resetAllConfig = () => {
      ElMessageBox.confirm(
        '确定要重置所有安全配置吗？这将恢复到系统默认设置。',
        '重置确认',
        {
          type: 'warning',
          confirmButtonText: '确定重置',
          cancelButtonText: '取消'
        }
      ).then(() => {
        // 重置限制器配置
        Object.assign(limiterConfig, {
          rateLimitWindow: 60,
          maxRequestsPerWindow: 100,
          banDuration: 3600,
          delayResponseMs: 1000,
          warningThreshold: 30,
          criticalThreshold: 10
        })

        // 重置分析器配置
        Object.assign(analyzerConfig, {
          suspiciousRequestThreshold: 50,
          pathRepeatThreshold: 10,
          botDetectionEnabled: true,
          patternDetectionEnabled: true,
          analysisWindow: 3600
        })

        // 重置代理配置
        proxyConfig.trustedProxies.splice(0)
        proxyConfig.trustedProxies.push(
          '127.0.0.1/32',
          '10.0.0.0/8',
          '172.16.0.0/12',
          '192.168.0.0/16'
        )
        proxyConfig.trustedHeaders.splice(0)
        proxyConfig.trustedHeaders.push(
          'X-Real-IP',
          'X-Forwarded-For',
          'CF-Connecting-IP'
        )
        proxyConfig.skipPrivateRanges = true
        proxyConfig.maxProxyDepth = 10

        ElMessage.success('所有配置已重置为默认值')
      })
    }

    return {
      saving,
      limiterConfig,
      analyzerConfig,
      proxyConfig,
      addTrustedProxy,
      removeTrustedProxy,
      saveAllConfig,
      resetAllConfig
    }
  }
}
</script>

<style scoped>
.security-config {
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

.proxy-list {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 12px;
  background: #fafbfc;
}

.proxy-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.proxy-item:last-child {
  margin-bottom: 0;
}

.save-actions {
  text-align: center;
  padding: 20px 0;
  border-top: 1px solid #ebeef5;
  margin-top: 20px;
}

.save-actions .el-button {
  margin: 0 10px;
}
</style>
