<template>
  <div class="user-list">
    <div class="page-title">
      <el-icon class="title-icon"><UserFilled /></el-icon>
      用户列表
    </div>

    <!-- 搜索和筛选 -->
    <el-card class="search-form">
      <el-form :model="searchForm" :inline="true">
        <el-form-item label="用户指纹">
          <el-input
            v-model="searchForm.fingerprint"
            placeholder="请输入用户指纹"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        
        <el-form-item label="IP地址">
          <el-input
            v-model="searchForm.ip"
            placeholder="请输入IP地址"
            clearable
            style="width: 160px"
          />
        </el-form-item>
        
        <el-form-item label="分数范围">
          <el-select v-model="searchForm.scoreRange" placeholder="分数范围" clearable style="width: 150px">
            <el-option label="优秀 (80-100)" value="excellent" />
            <el-option label="良好 (60-79)" value="good" />
            <el-option label="警告 (30-59)" value="warning" />
            <el-option label="危险 (0-29)" value="danger" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="风险等级">
          <el-select v-model="searchForm.riskLevel" placeholder="风险等级" clearable style="width: 120px">
            <el-option label="低风险" value="low" />
            <el-option label="中风险" value="medium" />
            <el-option label="高风险" value="high" />
            <el-option label="严重" value="critical" />
          </el-select>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSearch" :loading="loading">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="warning" @click="showBatchDialog = true">
          <el-icon><Operation /></el-icon>
          批量操作
        </el-button>
      </div>
      
      <div class="toolbar-right">
        <el-text type="info">
          共 {{ pagination.total }} 个用户
        </el-text>
      </div>
    </div>

    <!-- 用户表格 -->
    <el-card class="data-table">
      <el-table
        :data="tableData"
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
        
        <el-table-column prop="lastIp" label="最后IP" width="140" />
        
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
        
        <el-table-column prop="banCount" label="封禁次数" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.banCount > 0" type="danger" size="small">
              {{ scope.row.banCount }}
            </el-tag>
            <span v-else>0</span>
          </template>
        </el-table-column>
        
        <el-table-column prop="firstSeen" label="首次访问" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.firstSeen) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="lastSeen" label="最后访问" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.lastSeen) }}
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
                <el-icon><View /></el-icon>
                详情
              </el-button>
              <el-button
                type="text"
                size="small"
                @click="handleResetScore(scope.row)"
              >
                <el-icon><RefreshRight /></el-icon>
                重置分数
              </el-button>
              <el-button
                type="text"
                size="small"
                @click="handleBanUser(scope.row)"
                class="danger-button"
              >
                <el-icon><CircleClose /></el-icon>
                封禁
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

    <!-- 批量操作对话框 -->
    <el-dialog v-model="showBatchDialog" title="批量操作" width="500px">
      <el-form :model="batchForm" label-width="100px">
        <el-form-item label="操作类型">
          <el-select v-model="batchForm.operation" style="width: 100%">
            <el-option label="重置分数" value="reset" />
            <el-option label="调整分数" value="adjust" />
            <el-option label="批量封禁" value="ban" />
            <el-option label="添加白名单" value="whitelist" />
          </el-select>
        </el-form-item>
        
        <el-form-item v-if="batchForm.operation === 'adjust'" label="分数调整">
          <el-input-number v-model="batchForm.adjustment" :min="-50" :max="50" style="width: 100%" />
        </el-form-item>
        
        <el-form-item v-if="batchForm.operation === 'ban'" label="封禁时长">
          <el-select v-model="batchForm.duration" style="width: 100%">
            <el-option label="1小时" value="1h" />
            <el-option label="6小时" value="6h" />
            <el-option label="24小时" value="24h" />
            <el-option label="7天" value="168h" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="操作原因">
          <el-input v-model="batchForm.reason" placeholder="请输入操作原因" />
        </el-form-item>
        
        <el-form-item label="选中用户">
          <el-text type="info">已选择 {{ selectedUsers.length }} 个用户</el-text>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showBatchDialog = false">取消</el-button>
        <el-button type="primary" @click="handleBatchOperation" :loading="saving">
          确定执行
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UserFilled, Search, Refresh, View, RefreshRight, CircleClose, Operation } from '@element-plus/icons-vue'

export default {
  name: 'UserList',
  components: {
    UserFilled,
    Search,
    Refresh,
    View,
    RefreshRight,
    CircleClose,
    Operation
  },
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const saving = ref(false)
    const showBatchDialog = ref(false)

    // 搜索表单
    const searchForm = reactive({
      fingerprint: '',
      ip: '',
      scoreRange: '',
      riskLevel: ''
    })

    // 表格数据
    const tableData = ref([
      {
        fingerprint: 'abc123def456ghi789jkl012',
        lastIp: '192.168.1.100',
        currentScore: 85,
        riskLevel: 'low',
        totalRequests: 1250,
        banCount: 0,
        firstSeen: '2024-01-10T08:30:00Z',
        lastSeen: '2024-01-15T14:20:00Z'
      },
      {
        fingerprint: 'xyz789uvw012rst345mno678',
        lastIp: '10.0.0.50',
        currentScore: 45,
        riskLevel: 'medium',
        totalRequests: 890,
        banCount: 1,
        firstSeen: '2024-01-12T10:15:00Z',
        lastSeen: '2024-01-15T13:45:00Z'
      },
      {
        fingerprint: 'def456ghi789jkl012mno345',
        lastIp: '203.0.113.25',
        currentScore: 15,
        riskLevel: 'high',
        totalRequests: 2100,
        banCount: 3,
        firstSeen: '2024-01-08T16:20:00Z',
        lastSeen: '2024-01-15T12:10:00Z'
      }
    ])

    // 分页信息
    const pagination = reactive({
      page: 1,
      size: 20,
      total: 3
    })

    // 选中的用户
    const selectedUsers = ref([])

    // 批量操作表单
    const batchForm = reactive({
      operation: '',
      adjustment: 0,
      duration: '24h',
      reason: ''
    })

    // 搜索
    const handleSearch = () => {
      loading.value = true
      // 模拟搜索
      setTimeout(() => {
        loading.value = false
        ElMessage.success('搜索完成')
      }, 1000)
    }

    // 重置搜索
    const handleReset = () => {
      Object.assign(searchForm, {
        fingerprint: '',
        ip: '',
        scoreRange: '',
        riskLevel: ''
      })
      handleSearch()
    }

    // 刷新数据
    const handleRefresh = () => {
      loading.value = true
      setTimeout(() => {
        loading.value = false
        ElMessage.success('数据已刷新')
      }, 1000)
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

    // 重置用户分数
    const handleResetScore = (user) => {
      ElMessageBox.confirm(
        `确定要重置用户 ${user.fingerprint.slice(0, 8)}... 的分数吗？`,
        '重置确认',
        {
          type: 'warning'
        }
      ).then(() => {
        ElMessage.success('用户分数已重置')
      })
    }

    // 封禁用户
    const handleBanUser = (user) => {
      ElMessageBox.prompt(
        `封禁用户 ${user.fingerprint.slice(0, 8)}...`,
        '封禁确认',
        {
          confirmButtonText: '确定封禁',
          cancelButtonText: '取消',
          inputPlaceholder: '请输入封禁原因',
          inputValidator: (value) => {
            if (!value) {
              return '请输入封禁原因'
            }
            return true
          }
        }
      ).then(({ value }) => {
        ElMessage.success(`用户已封禁，原因: ${value}`)
      })
    }

    // 批量操作
    const handleBatchOperation = () => {
      if (selectedUsers.value.length === 0) {
        ElMessage.warning('请先选择要操作的用户')
        return
      }

      if (!batchForm.reason) {
        ElMessage.warning('请输入操作原因')
        return
      }

      saving.value = true
      setTimeout(() => {
        saving.value = false
        showBatchDialog.value = false
        ElMessage.success(`批量${batchForm.operation}操作完成，影响${selectedUsers.value.length}个用户`)
        
        // 清空选择
        selectedUsers.value = []
        
        // 重置表单
        Object.assign(batchForm, {
          operation: '',
          adjustment: 0,
          duration: '24h',
          reason: ''
        })
      }, 2000)
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
      showBatchDialog,
      searchForm,
      tableData,
      pagination,
      selectedUsers,
      batchForm,
      handleSearch,
      handleReset,
      handleRefresh,
      handlePageChange,
      handleSizeChange,
      handleSelectionChange,
      handleViewUser,
      handleResetScore,
      handleBanUser,
      handleBatchOperation,
      getScoreClass,
      getRiskType,
      getRiskText,
      formatTime
    }
  }
}
</script>

<style scoped>
.user-list {
  padding: 0;
}

.search-form {
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
  flex-wrap: wrap;
}

.danger-button {
  color: #f56c6c !important;
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
</style>
