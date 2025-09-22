<template>
  <div class="access-logs">
    <div class="page-title">
      <el-icon class="title-icon"><Document /></el-icon>
      访问日志
    </div>

    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :model="searchForm" :inline="true" @submit.prevent="handleSearch">
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
        
        <el-form-item label="访问路径">
          <el-input
            v-model="searchForm.path"
            placeholder="请输入访问路径"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        
        <el-form-item label="请求方法">
          <el-select v-model="searchForm.method" placeholder="请求方法" clearable style="width: 120px">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="处理动作">
          <el-select v-model="searchForm.action" placeholder="处理动作" clearable style="width: 120px">
            <el-option label="通过" value="allow" />
            <el-option label="限制" value="limit" />
            <el-option label="验证" value="challenge" />
            <el-option label="封禁" value="ban" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="分数范围">
          <el-input-number
            v-model="searchForm.minScore"
            :min="0"
            :max="100"
            placeholder="最低分"
            style="width: 100px"
          />
          <span style="margin: 0 8px">-</span>
          <el-input-number
            v-model="searchForm.maxScore"
            :min="0"
            :max="100"
            placeholder="最高分"
            style="width: 100px"
          />
        </el-form-item>
        
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="searchForm.timeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 350px"
          />
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
          <el-button type="success" @click="handleExport">
            <el-icon><Download /></el-icon>
            导出
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
        <el-button type="info" @click="showRecentLogs = true">
          <el-icon><Clock /></el-icon>
          实时日志
        </el-button>
      </div>
      
      <div class="toolbar-right">
        <el-text type="info">
          共 {{ pagination.total }} 条记录
        </el-text>
      </div>
    </div>

    <!-- 数据表格 -->
    <el-card class="data-table">
      <el-table
        :data="tableData"
        v-loading="loading"
        stripe
        @sort-change="handleSortChange"
      >
        <el-table-column prop="id" label="ID" width="80" sortable="custom" />
        
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
        
        <el-table-column prop="ip" label="IP地址" width="140">
          <template #default="scope">
            <el-button
              type="text"
              @click="handleSearchByIP(scope.row.ip)"
              class="ip-link"
            >
              {{ scope.row.ip }}
            </el-button>
          </template>
        </el-table-column>
        
        <el-table-column prop="path" label="访问路径" min-width="200" show-overflow-tooltip />
        
        <el-table-column prop="method" label="方法" width="80">
          <template #default="scope">
            <el-tag :type="getMethodType(scope.row.method)" size="small">
              {{ scope.row.method }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="score" label="分数" width="80" sortable="custom">
          <template #default="scope">
            <span :class="getScoreClass(scope.row.score)">
              {{ scope.row.score }}
            </span>
          </template>
        </el-table-column>
        
        <el-table-column prop="action" label="动作" width="80">
          <template #default="scope">
            <el-tag :type="getActionType(scope.row.action)" size="small">
              {{ getActionText(scope.row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="user_agent" label="User-Agent" min-width="250" show-overflow-tooltip />
        
        <el-table-column prop="timestamp" label="访问时间" width="180" sortable="custom">
          <template #default="scope">
            {{ formatTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button
              type="text"
              size="small"
              @click="handleViewDetails(scope.row)"
            >
              <el-icon><View /></el-icon>
              详情
            </el-button>
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

    <!-- 访问详情对话框 -->
    <el-dialog
      v-model="detailDialog.visible"
      title="访问详情"
      width="800px"
      destroy-on-close
    >
      <el-descriptions v-if="detailDialog.data" :column="2" border>
        <el-descriptions-item label="ID">{{ detailDialog.data.id }}</el-descriptions-item>
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
            {{ detailDialog.data.score }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="处理动作">
          <el-tag :type="getActionType(detailDialog.data.action)" size="small">
            {{ getActionText(detailDialog.data.action) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="访问时间">{{ formatTime(detailDialog.data.timestamp) }}</el-descriptions-item>
        <el-descriptions-item label="User-Agent" :span="2">
          <el-text class="user-agent-text">{{ detailDialog.data.user_agent }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
      
      <template #footer>
        <el-button @click="detailDialog.visible = false">关闭</el-button>
        <el-button type="primary" @click="handleViewUser(detailDialog.data.fingerprint)">
          查看用户详情
        </el-button>
      </template>
    </el-dialog>

    <!-- 实时日志对话框 -->
    <el-dialog
      v-model="showRecentLogs"
      title="实时访问日志"
      width="1200px"
      destroy-on-close
    >
      <RealtimeLogs @close="showRecentLogs = false" />
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Search, Refresh, Download, Clock, View } from '@element-plus/icons-vue'
import RealtimeLogs from './realtime.vue'
import { getAccessLogs, advancedSearchLogs, exportLogs } from '@/api/logs'

export default {
  name: 'AccessLogs',
  components: {
    Document,
    Search,
    Refresh,
    Download,
    Clock,
    View,
    RealtimeLogs
  },
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const showRecentLogs = ref(false)

    // 搜索表单
    const searchForm = reactive({
      fingerprint: '',
      ip: '',
      path: '',
      method: '',
      action: '',
      minScore: null,
      maxScore: null,
      timeRange: null
    })

    // 表格数据
    const tableData = ref([])

    // 分页信息
    const pagination = reactive({
      page: 1,
      size: 20,
      total: 0,
      totalPages: 0
    })

    // 排序信息
    const sortInfo = reactive({
      orderBy: 'timestamp',
      orderDir: 'DESC'
    })

    // 详情对话框
    const detailDialog = reactive({
      visible: false,
      data: null
    })

    // 加载数据
    const loadData = async () => {
      loading.value = true
      try {
        const params = {
          page: pagination.page,
          size: pagination.size,
          order_by: sortInfo.orderBy,
          order_dir: sortInfo.orderDir
        }

        // 添加搜索条件
        if (searchForm.fingerprint) params.fingerprint = searchForm.fingerprint
        if (searchForm.ip) params.ip = searchForm.ip
        if (searchForm.path) params.path = searchForm.path
        if (searchForm.method) params.method = searchForm.method
        if (searchForm.action) params.action = searchForm.action
        if (searchForm.minScore !== null) params.min_score = searchForm.minScore
        if (searchForm.maxScore !== null) params.max_score = searchForm.maxScore
        
        if (searchForm.timeRange && searchForm.timeRange.length === 2) {
          params.start_time = searchForm.timeRange[0]
          params.end_time = searchForm.timeRange[1]
        }

        const response = await getAccessLogs(params)
        
        if (response.success) {
          tableData.value = response.data.records || []
          pagination.total = response.data.total || 0
          pagination.totalPages = response.data.total_pages || 0
        } else {
          ElMessage.error(response.error || '获取数据失败')
        }
      } catch (error) {
        console.error('加载数据失败:', error)
        ElMessage.error('网络错误，请稍后重试')
      } finally {
        loading.value = false
      }
    }

    // 搜索
    const handleSearch = () => {
      pagination.page = 1
      loadData()
    }

    // 重置搜索
    const handleReset = () => {
      Object.assign(searchForm, {
        fingerprint: '',
        ip: '',
        path: '',
        method: '',
        action: '',
        minScore: null,
        maxScore: null,
        timeRange: null
      })
      pagination.page = 1
      loadData()
    }

    // 刷新数据
    const handleRefresh = () => {
      loadData()
    }

    // 导出数据
    const handleExport = async () => {
      try {
        const params = { format: 'json' }
        
        // 添加搜索条件到导出参数
        if (searchForm.fingerprint) params.fingerprint = searchForm.fingerprint
        if (searchForm.ip) params.ip = searchForm.ip
        if (searchForm.timeRange && searchForm.timeRange.length === 2) {
          params.start_time = searchForm.timeRange[0]
          params.end_time = searchForm.timeRange[1]
        }

        await exportLogs(params)
        ElMessage.success('导出成功')
      } catch (error) {
        console.error('导出失败:', error)
        ElMessage.error('导出失败')
      }
    }

    // 分页变化
    const handlePageChange = (page) => {
      pagination.page = page
      loadData()
    }

    const handleSizeChange = (size) => {
      pagination.size = size
      pagination.page = 1
      loadData()
    }

    // 排序变化
    const handleSortChange = ({ prop, order }) => {
      if (order === 'ascending') {
        sortInfo.orderDir = 'ASC'
      } else if (order === 'descending') {
        sortInfo.orderDir = 'DESC'
      } else {
        sortInfo.orderDir = 'DESC'
      }
      
      sortInfo.orderBy = prop || 'timestamp'
      loadData()
    }

    // 查看用户详情
    const handleViewUser = (fingerprint) => {
      router.push(`/users/detail/${fingerprint}`)
    }

    // 按IP搜索
    const handleSearchByIP = (ip) => {
      searchForm.ip = ip
      handleSearch()
    }

    // 查看详情
    const handleViewDetails = (row) => {
      detailDialog.data = { ...row }
      detailDialog.visible = true
    }

    // 格式化时间
    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString('zh-CN')
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
      loadData()
    })

    return {
      loading,
      showRecentLogs,
      searchForm,
      tableData,
      pagination,
      sortInfo,
      detailDialog,
      loadData,
      handleSearch,
      handleReset,
      handleRefresh,
      handleExport,
      handlePageChange,
      handleSizeChange,
      handleSortChange,
      handleViewUser,
      handleSearchByIP,
      handleViewDetails,
      formatTime,
      getActionType,
      getActionText,
      getMethodType,
      getScoreClass
    }
  }
}
</script>

<style scoped>
.access-logs {
  padding: 0;
}

.search-form {
  margin-bottom: 20px;
}

.search-form .el-form {
  margin-bottom: -18px;
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

.fingerprint-link,
.ip-link {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.user-agent-text {
  word-break: break-all;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
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
  .search-form .el-form {
    display: block;
  }
  
  .search-form .el-form-item {
    display: block;
    margin-bottom: 12px;
  }
  
  .toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .toolbar-left,
  .toolbar-right {
    justify-content: center;
  }
}
</style>
