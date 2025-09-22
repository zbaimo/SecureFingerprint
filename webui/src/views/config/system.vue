<template>
  <div class="system-config">
    <div class="page-title">
      <el-icon class="title-icon"><Tools /></el-icon>
      系统设置
    </div>

    <!-- 基础配置 -->
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>基础配置</span>
          <el-button type="primary" @click="saveConfig" :loading="saving">
            <el-icon><Check /></el-icon>
            保存配置
          </el-button>
        </div>
      </template>

      <el-form :model="systemConfig" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="服务端口">
              <el-input-number
                v-model="systemConfig.server.port"
                :min="1"
                :max="65535"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="调试模式">
              <el-switch v-model="systemConfig.server.debug" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Redis地址">
              <el-input v-model="systemConfig.redis.addr" placeholder="redis:6379" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Redis密码">
              <el-input
                v-model="systemConfig.redis.password"
                type="password"
                placeholder="留空表示无密码"
                show-password
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="MySQL DSN">
              <el-input
                v-model="systemConfig.mysql.dsn"
                type="textarea"
                :rows="2"
                placeholder="user:password@tcp(host:port)/database"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大连接数">
              <el-input-number
                v-model="systemConfig.mysql.maxOpenConns"
                :min="1"
                :max="1000"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="日志级别">
              <el-select v-model="systemConfig.logging.level" style="width: 100%">
                <el-option label="Debug" value="debug" />
                <el-option label="Info" value="info" />
                <el-option label="Warn" value="warn" />
                <el-option label="Error" value="error" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="日志文件">
              <el-input v-model="systemConfig.logging.file" placeholder="logs/app.log" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="允许用户注册">
              <el-switch v-model="systemConfig.user.allowRegistration" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="WebUI启用">
              <el-switch v-model="systemConfig.webui.enabled" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 性能配置 -->
    <el-card class="config-card">
      <template #header>
        <span>性能配置</span>
      </template>

      <el-form :model="performanceConfig" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="并发请求数">
              <el-input-number
                v-model="performanceConfig.maxConcurrentRequests"
                :min="100"
                :max="10000"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="缓存TTL(秒)">
              <el-input-number
                v-model="performanceConfig.cacheTTL"
                :min="60"
                :max="86400"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="批处理大小">
              <el-input-number
                v-model="performanceConfig.batchSize"
                :min="10"
                :max="1000"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 配置历史 -->
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>配置历史</span>
          <el-button @click="loadConfigHistory">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table :data="configHistory" size="small">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="type" label="配置类型" width="120">
          <template #default="scope">
            <el-tag :type="getConfigTypeColor(scope.row.type)" size="small">
              {{ getConfigTypeName(scope.row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="changes" label="变更内容" show-overflow-tooltip />
        <el-table-column prop="operator" label="操作员" width="100" />
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button type="text" size="small" @click="viewConfigDetails(scope.row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 配置详情对话框 -->
    <el-dialog
      v-model="detailDialog.visible"
      title="配置变更详情"
      width="800px"
    >
      <el-descriptions v-if="detailDialog.data" :column="1" border>
        <el-descriptions-item label="配置类型">{{ getConfigTypeName(detailDialog.data.type) }}</el-descriptions-item>
        <el-descriptions-item label="变更内容">{{ detailDialog.data.changes }}</el-descriptions-item>
        <el-descriptions-item label="操作员">{{ detailDialog.data.operator }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ formatTime(detailDialog.data.timestamp) }}</el-descriptions-item>
      </el-descriptions>
      
      <template #footer>
        <el-button @click="detailDialog.visible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Tools, Check, Refresh } from '@element-plus/icons-vue'

export default {
  name: 'SystemConfig',
  components: {
    Tools,
    Check,
    Refresh
  },
  setup() {
    const saving = ref(false)
    const loading = ref(false)

    // 系统配置
    const systemConfig = reactive({
      server: {
        port: 8080,
        debug: false
      },
      redis: {
        addr: 'redis:6379',
        password: '',
        db: 0,
        poolSize: 20
      },
      mysql: {
        dsn: 'user:password@tcp(mysql:3306)/securefingerprint?charset=utf8mb4&parseTime=True&loc=Local',
        maxOpenConns: 100,
        maxIdleConns: 20
      },
      logging: {
        level: 'info',
        file: 'logs/app.log',
        maxSize: 100,
        maxBackups: 7,
        maxAge: 30
      },
      webui: {
        enabled: true,
        staticPath: './webui/build',
        apiPrefix: '/api/v1'
      },
      user: {
        allowRegistration: false
      }
    })

    // 性能配置
    const performanceConfig = reactive({
      maxConcurrentRequests: 1000,
      cacheTTL: 3600,
      batchSize: 100,
      dbPoolSize: 20,
      redisPoolSize: 20
    })

    // 配置历史
    const configHistory = ref([
      {
        id: 1,
        type: 'system',
        changes: '更新服务端口: 8080 -> 8081',
        operator: 'admin',
        timestamp: '2024-01-15T10:30:00Z'
      },
      {
        id: 2,
        type: 'performance',
        changes: '调整并发请求数: 500 -> 1000',
        operator: 'admin',
        timestamp: '2024-01-14T15:20:00Z'
      },
      {
        id: 3,
        type: 'logging',
        changes: '修改日志级别: debug -> info',
        operator: 'admin',
        timestamp: '2024-01-13T09:15:00Z'
      }
    ])

    // 详情对话框
    const detailDialog = reactive({
      visible: false,
      data: null
    })

    // 保存配置
    const saveConfig = async () => {
      saving.value = true
      try {
        // 这里应该调用API保存配置
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        ElMessage.success('配置保存成功')
        
        // 添加到历史记录
        configHistory.value.unshift({
          id: Date.now(),
          type: 'system',
          changes: '更新系统配置',
          operator: 'admin',
          timestamp: new Date().toISOString()
        })
      } catch (error) {
        ElMessage.error('配置保存失败')
      } finally {
        saving.value = false
      }
    }

    // 加载配置历史
    const loadConfigHistory = async () => {
      loading.value = true
      try {
        // 这里应该调用API获取配置历史
        await new Promise(resolve => setTimeout(resolve, 500))
        ElMessage.success('配置历史已刷新')
      } catch (error) {
        ElMessage.error('加载配置历史失败')
      } finally {
        loading.value = false
      }
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

    onMounted(() => {
      // 初始化加载配置
    })

    return {
      saving,
      loading,
      systemConfig,
      performanceConfig,
      configHistory,
      detailDialog,
      saveConfig,
      loadConfigHistory,
      viewConfigDetails,
      getConfigTypeColor,
      getConfigTypeName,
      formatTime
    }
  }
}
</script>

<style scoped>
.system-config {
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
</style>
