<template>
  <div class="ban-rules">
    <div class="page-title">
      <el-icon class="title-icon"><CircleClose /></el-icon>
      封禁管理
    </div>

    <!-- 封禁统计 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card">
          <div class="stat-number">{{ banStats.totalBans }}</div>
          <div class="stat-label">总封禁数</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card active">
          <div class="stat-number">{{ banStats.activeBans }}</div>
          <div class="stat-label">当前封禁</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card auto">
          <div class="stat-number">{{ banStats.autoBans }}</div>
          <div class="stat-label">自动封禁</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6" :lg="6" :xl="6">
        <div class="stat-card manual">
          <div class="stat-number">{{ banStats.manualBans }}</div>
          <div class="stat-label">手动封禁</div>
        </div>
      </el-col>
    </el-row>

    <!-- 操作工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" @click="showBanDialog = true">
          <el-icon><CircleClose /></el-icon>
          手动封禁
        </el-button>
        <el-button type="success" @click="showBatchUnbanDialog = true">
          <el-icon><CircleCheck /></el-icon>
          批量解封
        </el-button>
        <el-button @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
      
      <div class="toolbar-right">
        <el-select v-model="filterStatus" @change="handleFilterChange" style="width: 150px">
          <el-option label="全部状态" value="" />
          <el-option label="活跃封禁" value="active" />
          <el-option label="已过期" value="expired" />
          <el-option label="手动解封" value="manual_unban" />
        </el-select>
      </div>
    </div>

    <!-- 封禁列表 -->
    <el-card class="data-table">
      <el-table
        :data="bannedUsers"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
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
        
        <el-table-column prop="ip" label="IP地址" width="140" />
        
        <el-table-column prop="reason" label="封禁原因" min-width="150" show-overflow-tooltip />
        
        <el-table-column prop="bannedAt" label="封禁时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.bannedAt) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="duration" label="封禁时长" width="100" />
        
        <el-table-column prop="expiresAt" label="到期时间" width="180">
          <template #default="scope">
            <span :class="getExpiryClass(scope.row.expiresAt)">
              {{ formatTime(scope.row.expiresAt) }}
            </span>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getBanStatusType(scope.row)" size="small">
              {{ getBanStatusText(scope.row) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <div class="action-buttons">
              <el-button
                v-if="isActiveBan(scope.row)"
                type="text"
                size="small"
                @click="handleUnban(scope.row)"
              >
                <el-icon><CircleCheck /></el-icon>
                解封
              </el-button>
              <el-button
                type="text"
                size="small"
                @click="handleViewUser(scope.row.fingerprint)"
              >
                <el-icon><View /></el-icon>
                详情
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 手动封禁对话框 -->
    <el-dialog v-model="showBanDialog" title="手动封禁用户" width="500px">
      <el-form :model="banForm" label-width="100px">
        <el-form-item label="用户指纹" required>
          <el-input
            v-model="banForm.fingerprint"
            placeholder="请输入用户指纹"
          />
        </el-form-item>
        <el-form-item label="封禁时长" required>
          <el-select v-model="banForm.duration" style="width: 100%">
            <el-option label="1小时" value="1h" />
            <el-option label="6小时" value="6h" />
            <el-option label="24小时" value="24h" />
            <el-option label="7天" value="168h" />
            <el-option label="30天" value="720h" />
          </el-select>
        </el-form-item>
        <el-form-item label="封禁原因" required>
          <el-input
            v-model="banForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入封禁原因"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showBanDialog = false">取消</el-button>
        <el-button type="danger" @click="confirmBan" :loading="saving">
          确定封禁
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量解封对话框 -->
    <el-dialog v-model="showBatchUnbanDialog" title="批量解封" width="400px">
      <el-form label-width="100px">
        <el-form-item label="选中用户">
          <el-text type="info">已选择 {{ selectedUsers.length }} 个用户</el-text>
        </el-form-item>
        <el-form-item label="解封原因">
          <el-input
            v-model="batchUnbanReason"
            type="textarea"
            :rows="3"
            placeholder="请输入解封原因"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showBatchUnbanDialog = false">取消</el-button>
        <el-button type="success" @click="confirmBatchUnban" :loading="saving">
          确定解封
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CircleClose, CircleCheck, Refresh, View } from '@element-plus/icons-vue'

export default {
  name: 'BanRules',
  components: {
    CircleClose,
    CircleCheck,
    Refresh,
    View
  },
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const saving = ref(false)
    const showBanDialog = ref(false)
    const showBatchUnbanDialog = ref(false)
    const filterStatus = ref('')

    // 封禁统计
    const banStats = reactive({
      totalBans: 150,
      activeBans: 25,
      autoBans: 100,
      manualBans: 50
    })

    // 封禁用户列表
    const bannedUsers = ref([
      {
        fingerprint: 'abc123def456ghi789jkl012',
        ip: '192.168.1.100',
        reason: '恶意扫描',
        bannedAt: '2024-01-15T10:00:00Z',
        duration: '24h',
        expiresAt: '2024-01-16T10:00:00Z',
        banCount: 3
      },
      {
        fingerprint: 'xyz789uvw012rst345mno678',
        ip: '10.0.0.50',
        reason: '频繁请求',
        bannedAt: '2024-01-15T14:30:00Z',
        duration: '1h',
        expiresAt: '2024-01-15T15:30:00Z',
        banCount: 1
      }
    ])

    // 分页信息
    const pagination = reactive({
      page: 1,
      size: 20,
      total: 2
    })

    // 选中的用户
    const selectedUsers = ref([])

    // 封禁表单
    const banForm = reactive({
      fingerprint: '',
      duration: '24h',
      reason: ''
    })

    // 批量解封原因
    const batchUnbanReason = ref('')

    // 刷新数据
    const handleRefresh = () => {
      loading.value = true
      setTimeout(() => {
        loading.value = false
      }, 1000)
    }

    // 筛选变化
    const handleFilterChange = () => {
      handleRefresh()
    }

    // 分页变化
    const handlePageChange = (page) => {
      pagination.page = page
      handleRefresh()
    }

    const handleSizeChange = (size) => {
      pagination.size = size
      pagination.page = 1
      handleRefresh()
    }

    // 选择变化
    const handleSelectionChange = (selection) => {
      selectedUsers.value = selection
    }

    // 查看用户详情
    const handleViewUser = (fingerprint) => {
      router.push(`/users/detail/${fingerprint}`)
    }

    // 解封用户
    const handleUnban = (user) => {
      ElMessageBox.confirm(
        `确定要解封用户 ${user.fingerprint.slice(0, 8)}... 吗？`,
        '解封确认',
        {
          type: 'warning'
        }
      ).then(() => {
        ElMessage.success('用户已解封')
        handleRefresh()
      })
    }

    // 确认封禁
    const confirmBan = async () => {
      if (!banForm.fingerprint || !banForm.reason) {
        ElMessage.warning('请填写完整信息')
        return
      }

      saving.value = true
      try {
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        ElMessage.success('用户封禁成功')
        showBanDialog.value = false
        
        // 重置表单
        Object.assign(banForm, {
          fingerprint: '',
          duration: '24h',
          reason: ''
        })
        
        handleRefresh()
      } catch (error) {
        ElMessage.error('封禁失败')
      } finally {
        saving.value = false
      }
    }

    // 确认批量解封
    const confirmBatchUnban = async () => {
      if (selectedUsers.value.length === 0) {
        ElMessage.warning('请先选择要解封的用户')
        return
      }

      if (!batchUnbanReason.value) {
        ElMessage.warning('请输入解封原因')
        return
      }

      saving.value = true
      try {
        await new Promise(resolve => setTimeout(resolve, 1500))
        
        ElMessage.success(`批量解封成功，共解封${selectedUsers.value.length}个用户`)
        showBatchUnbanDialog.value = false
        batchUnbanReason.value = ''
        selectedUsers.value = []
        
        handleRefresh()
      } catch (error) {
        ElMessage.error('批量解封失败')
      } finally {
        saving.value = false
      }
    }

    // 判断是否为活跃封禁
    const isActiveBan = (ban) => {
      return new Date(ban.expiresAt) > new Date()
    }

    // 获取封禁状态类型
    const getBanStatusType = (ban) => {
      if (isActiveBan(ban)) {
        return 'danger'
      }
      return 'info'
    }

    // 获取封禁状态文本
    const getBanStatusText = (ban) => {
      if (isActiveBan(ban)) {
        return '封禁中'
      }
      return '已过期'
    }

    // 获取到期时间样式
    const getExpiryClass = (expiresAt) => {
      const now = new Date()
      const expiry = new Date(expiresAt)
      
      if (expiry <= now) {
        return 'expired'
      }
      
      const diff = expiry - now
      const hours = diff / (1000 * 60 * 60)
      
      if (hours < 1) {
        return 'expiring-soon'
      }
      
      return 'active'
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
      showBanDialog,
      showBatchUnbanDialog,
      filterStatus,
      banStats,
      bannedUsers,
      pagination,
      selectedUsers,
      banForm,
      batchUnbanReason,
      handleRefresh,
      handleFilterChange,
      handlePageChange,
      handleSizeChange,
      handleSelectionChange,
      handleViewUser,
      handleUnban,
      confirmBan,
      confirmBatchUnban,
      isActiveBan,
      getBanStatusType,
      getBanStatusText,
      getExpiryClass,
      formatTime
    }
  }
}
</script>

<style scoped>
.ban-rules {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.fingerprint-link {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  color: white;
  margin-bottom: 10px;
  background: linear-gradient(135deg, #909399, #a6a9ad);
}

.stat-card.active {
  background: linear-gradient(135deg, #f56c6c, #f78989);
}

.stat-card.auto {
  background: linear-gradient(135deg, #e6a23c, #ebb563);
}

.stat-card.manual {
  background: linear-gradient(135deg, #409eff, #66b1ff);
}

.stat-number {
  font-size: 2.5em;
  font-weight: bold;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 1.1em;
}

.expired {
  color: #909399;
}

.expiring-soon {
  color: #e6a23c;
  font-weight: bold;
}

.active {
  color: #f56c6c;
  font-weight: bold;
}
</style>
