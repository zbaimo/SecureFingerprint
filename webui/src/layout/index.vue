<template>
  <div class="layout-container">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside :width="sidebarWidth" class="sidebar-container">
        <div class="logo-container">
          <img src="/favicon.ico" alt="Logo" class="logo" />
          <h1 v-if="!isCollapse" class="logo-title">防火墙控制器</h1>
        </div>
        
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :unique-opened="true"
          class="sidebar-menu"
          router
        >
          <sidebar-item
            v-for="route in routes"
            :key="route.path"
            :item="route"
            :base-path="route.path"
          />
        </el-menu>
      </el-aside>

      <!-- 主要区域 -->
      <el-container>
        <!-- 顶部导航栏 -->
        <el-header class="navbar">
          <div class="navbar-left">
            <el-button
              :icon="isCollapse ? 'Expand' : 'Fold'"
              @click="toggleSidebar"
              text
              size="large"
            />
            <breadcrumb />
          </div>
          
          <div class="navbar-right">
            <el-dropdown @command="handleCommand">
              <span class="user-info">
                <el-avatar :size="32" :icon="UserFilled" />
                <span class="username">管理员</span>
                <el-icon><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                  <el-dropdown-item command="settings">系统设置</el-dropdown-item>
                  <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <!-- 主内容区 -->
        <el-main class="main-content">
          <router-view v-slot="{ Component }">
            <transition name="fade-transform" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { UserFilled, ArrowDown } from '@element-plus/icons-vue'
import SidebarItem from './components/SidebarItem.vue'
import Breadcrumb from './components/Breadcrumb.vue'

export default {
  name: 'Layout',
  components: {
    SidebarItem,
    Breadcrumb
  },
  setup() {
    const route = useRoute()
    const isCollapse = ref(false)

    // 计算侧边栏宽度
    const sidebarWidth = computed(() => {
      return isCollapse.value ? '64px' : '220px'
    })

    // 当前激活的菜单项
    const activeMenu = computed(() => {
      const { meta, path } = route
      if (meta.activeMenu) {
        return meta.activeMenu
      }
      return path
    })

    // 路由配置（过滤隐藏的路由）
    const routes = computed(() => {
      return [
        {
          path: '/dashboard',
          name: 'Dashboard',
          meta: { title: '仪表板', icon: 'Monitor' }
        },
        {
          path: '/access',
          name: 'Access',
          meta: { title: '访问管理', icon: 'Connection' },
          children: [
            {
              path: '/access/logs',
              name: 'AccessLogs',
              meta: { title: '访问日志', icon: 'Document' }
            },
            {
              path: '/access/realtime',
              name: 'RealtimeLogs',
              meta: { title: '实时监控', icon: 'View' }
            }
          ]
        },
        {
          path: '/users',
          name: 'Users',
          meta: { title: '用户管理', icon: 'User' },
          children: [
            {
              path: '/users/list',
              name: 'UserList',
              meta: { title: '用户列表', icon: 'UserFilled' }
            },
            {
              path: '/users/scores',
              name: 'UserScores',
              meta: { title: '用户评分', icon: 'Star' }
            }
          ]
        },
        {
          path: '/rules',
          name: 'Rules',
          meta: { title: '风控规则', icon: 'Shield' },
          children: [
            {
              path: '/rules/ban',
              name: 'BanRules',
              meta: { title: '封禁管理', icon: 'CircleClose' }
            },
            {
              path: '/rules/whitelist',
              name: 'Whitelist',
              meta: { title: '白名单', icon: 'CircleCheck' }
            },
            {
              path: '/rules/analysis',
              name: 'BehaviorAnalysis',
              meta: { title: '行为分析', icon: 'TrendCharts' }
            }
          ]
        },
        {
          path: '/config',
          name: 'Config',
          meta: { title: '系统配置', icon: 'Setting' },
          children: [
            {
              path: '/config/system',
              name: 'SystemConfig',
              meta: { title: '系统设置', icon: 'Tools' }
            },
            {
              path: '/config/scoring',
              name: 'ScoringConfig',
              meta: { title: '评分规则', icon: 'Opportunity' }
            },
            {
              path: '/config/security',
              name: 'SecurityConfig',
              meta: { title: '安全配置', icon: 'Lock' }
            }
          ]
        }
      ]
    })

    // 切换侧边栏
    const toggleSidebar = () => {
      isCollapse.value = !isCollapse.value
    }

    // 处理用户菜单命令
    const handleCommand = (command) => {
      switch (command) {
        case 'profile':
          console.log('个人中心')
          break
        case 'settings':
          console.log('系统设置')
          break
        case 'logout':
          console.log('退出登录')
          break
      }
    }

    return {
      isCollapse,
      sidebarWidth,
      activeMenu,
      routes,
      toggleSidebar,
      handleCommand,
      UserFilled,
      ArrowDown
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar-container {
  background: #304156;
  transition: width 0.28s;
  box-shadow: 2px 0 6px rgba(0, 21, 41, 0.35);
}

.logo-container {
  display: flex;
  align-items: center;
  padding: 20px;
  background: #2b2f3a;
  color: white;
  height: 60px;
}

.logo {
  width: 32px;
  height: 32px;
  margin-right: 12px;
}

.logo-title {
  font-size: 18px;
  font-weight: 600;
  white-space: nowrap;
}

.sidebar-menu {
  border: none;
  height: calc(100vh - 60px);
  overflow-y: auto;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 220px;
}

.navbar {
  background: white;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px !important;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.navbar-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 6px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #606266;
}

.main-content {
  background: #f0f2f5;
  overflow-y: auto;
}

/* 页面切换动画 */
.fade-transform-leave-active,
.fade-transform-enter-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
