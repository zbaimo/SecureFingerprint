<template>
  <div class="whitelist">
    <div class="page-title">
      <el-icon class="title-icon"><CircleCheck /></el-icon>
      白名单管理
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon>
          添加白名单
        </el-button>
        <el-button @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
      
      <div class="toolbar-right">
        <el-text type="info">
          共 {{ whitelistUsers.length }} 个白名单用户
        </el-text>
      </div>
    </div>

    <!-- 白名单列表 -->
    <el-card class="data-table">
      <el-table :data="whitelistUsers" v-loading="loading" stripe>
        <el-table-column prop="fingerprint" label="用户指纹" width="140">
          <template #default="scope">
            <el-tooltip :content="scope.row.fingerprint" placement="top">
              <span class="fingerprint-text">{{ scope.row.fingerprint.slice(0, 8) }}...</span>
            </el-tooltip>
          </template>
        </el-table-column>
        
        <el-table-column prop="ip" label="IP地址" width="140" />
        
        <el-table-column prop="reason" label="添加原因" min-width="150" show-overflow-tooltip />
        
        <el-table-column prop="addedAt" label="添加时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.addedAt) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="expiresAt" label="到期时间" width="180">
          <template #default="scope">
            <span v-if="scope.row.duration === 'permanent'" class="permanent">永久</span>
            <span v-else :class="getExpiryClass(scope.row.expiresAt)">
              {{ formatTime(scope.row.expiresAt) }}
            </span>
          </template>
        </el-table-column>
        
        <el-table-column prop="duration" label="有效期" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.duration === 'permanent'" type="success" size="small">
              永久
            </el-tag>
            <el-tag v-else type="info" size="small">
              {{ scope.row.duration }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button
              type="text"
              size="small"
              @click="handleRemoveWhitelist(scope.row)"
              class="danger-button"
            >
              <el-icon><Delete /></el-icon>
              移除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加白名单对话框 -->
    <el-dialog v-model="showAddDialog" title="添加白名单" width="500px">
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="用户指纹" required>
          <el-input
            v-model="addForm.fingerprint"
            placeholder="请输入用户指纹"
          />
        </el-form-item>
        <el-form-item label="有效期" required>
          <el-select v-model="addForm.duration" style="width: 100%">
            <el-option label="1小时" value="1h" />
            <el-option label="24小时" value="24h" />
            <el-option label="7天" value="7d" />
            <el-option label="30天" value="30d" />
            <el-option label="永久" value="permanent" />
          </el-select>
        </el-form-item>
        <el-form-item label="添加原因" required>
          <el-input
            v-model="addForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入添加到白名单的原因"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAddWhitelist" :loading="saving">
          确定添加
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CircleCheck, Plus, Refresh, Delete } from '@element-plus/icons-vue'

export default {
  name: 'Whitelist',
  components: {
    CircleCheck,
    Plus,
    Refresh,
    Delete
  },
  setup() {
    const loading = ref(false)
    const saving = ref(false)
    const showAddDialog = ref(false)

    // 白名单用户列表
    const whitelistUsers = ref([
      {
        fingerprint: 'trusted123def456ghi789jkl',
        ip: '192.168.1.10',
        reason: '管理员用户',
        addedAt: '2024-01-10T09:00:00Z',
        expiresAt: '2025-01-10T09:00:00Z',
        duration: 'permanent'
      },
      {
        fingerprint: 'vip456ghi789jkl012mno345',
        ip: '10.0.0.100',
        reason: 'VIP客户',
        addedAt: '2024-01-12T14:30:00Z',
        expiresAt: '2024-02-12T14:30:00Z',
        duration: '30d'
      },
      {
        fingerprint: 'partner789jkl012mno345pqr',
        ip: '203.0.113.50',
        reason: '合作伙伴API',
        addedAt: '2024-01-14T16:15:00Z',
        expiresAt: '2024-01-21T16:15:00Z',
        duration: '7d'
      }
    ])

    // 添加表单
    const addForm = reactive({
      fingerprint: '',
      duration: '7d',
      reason: ''
    })

    // 刷新数据
    const handleRefresh = () => {
      loading.value = true
      setTimeout(() => {
        loading.value = false
        ElMessage.success('白名单已刷新')
      }, 1000)
    }

    // 确认添加白名单
    const confirmAddWhitelist = async () => {
      if (!addForm.fingerprint || !addForm.reason) {
        ElMessage.warning('请填写完整信息')
        return
      }

      saving.value = true
      try {
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        // 添加到列表
        const newItem = {
          fingerprint: addForm.fingerprint,
          ip: '未知',
          reason: addForm.reason,
          addedAt: new Date().toISOString(),
          expiresAt: addForm.duration === 'permanent' 
            ? new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString()
            : new Date(Date.now() + parseDuration(addForm.duration)).toISOString(),
          duration: addForm.duration
        }
        
        whitelistUsers.value.unshift(newItem)
        
        ElMessage.success('白名单添加成功')
        showAddDialog.value = false
        
        // 重置表单
        Object.assign(addForm, {
          fingerprint: '',
          duration: '7d',
          reason: ''
        })
      } catch (error) {
        ElMessage.error('添加失败')
      } finally {
        saving.value = false
      }
    }

    // 移除白名单
    const handleRemoveWhitelist = (user) => {
      ElMessageBox.confirm(
        `确定要将用户 ${user.fingerprint.slice(0, 8)}... 从白名单中移除吗？`,
        '移除确认',
        {
          type: 'warning',
          confirmButtonText: '确定移除',
          cancelButtonText: '取消'
        }
      ).then(() => {
        const index = whitelistUsers.value.findIndex(u => u.fingerprint === user.fingerprint)
        if (index > -1) {
          whitelistUsers.value.splice(index, 1)
        }
        ElMessage.success('用户已从白名单移除')
      })
    }

    // 解析持续时间
    const parseDuration = (duration) => {
      const units = {
        'h': 60 * 60 * 1000,
        'd': 24 * 60 * 60 * 1000
      }
      
      const match = duration.match(/^(\d+)([hd])$/)
      if (match) {
        const value = parseInt(match[1])
        const unit = match[2]
        return value * units[unit]
      }
      
      return 7 * 24 * 60 * 60 * 1000 // 默认7天
    }

    // 获取到期时间样式
    const getExpiryClass = (expiresAt) => {
      const now = new Date()
      const expiry = new Date(expiresAt)
      
      if (expiry <= now) {
        return 'expired'
      }
      
      const diff = expiry - now
      const days = diff / (1000 * 60 * 60 * 24)
      
      if (days < 1) {
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
      showAddDialog,
      whitelistUsers,
      addForm,
      handleRefresh,
      confirmAddWhitelist,
      handleRemoveWhitelist,
      getExpiryClass,
      formatTime
    }
  }
}
</script>

<style scoped>
.whitelist {
  padding: 0;
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

.fingerprint-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.danger-button {
  color: #f56c6c !important;
}

.permanent {
  color: #67c23a;
  font-weight: bold;
}

.expired {
  color: #909399;
}

.expiring-soon {
  color: #e6a23c;
  font-weight: bold;
}

.active {
  color: #67c23a;
}
</style>
