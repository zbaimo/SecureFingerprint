import { createRouter, createWebHistory } from 'vue-router'

// 布局组件
import Layout from '@/layout/index.vue'

const routes = [
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { 
          title: '仪表板',
          icon: 'Monitor',
          affix: true
        }
      }
    ]
  },
  {
    path: '/access',
    component: Layout,
    redirect: '/access/logs',
    name: 'Access',
    meta: {
      title: '访问管理',
      icon: 'Connection'
    },
    children: [
      {
        path: 'logs',
        name: 'AccessLogs',
        component: () => import('@/views/access/logs.vue'),
        meta: { 
          title: '访问日志',
          icon: 'Document'
        }
      },
      {
        path: 'realtime',
        name: 'RealtimeLogs',
        component: () => import('@/views/access/realtime.vue'),
        meta: { 
          title: '实时监控',
          icon: 'View'
        }
      }
    ]
  },
  {
    path: '/users',
    component: Layout,
    redirect: '/users/list',
    name: 'Users',
    meta: {
      title: '用户管理',
      icon: 'User'
    },
    children: [
      {
        path: 'list',
        name: 'UserList',
        component: () => import('@/views/users/list.vue'),
        meta: { 
          title: '用户列表',
          icon: 'UserFilled'
        }
      },
      {
        path: 'scores',
        name: 'UserScores',
        component: () => import('@/views/users/scores.vue'),
        meta: { 
          title: '用户评分',
          icon: 'Star'
        }
      },
      {
        path: 'detail/:fingerprint',
        name: 'UserDetail',
        component: () => import('@/views/users/detail.vue'),
        meta: { 
          title: '用户详情',
          icon: 'InfoFilled',
          hidden: true
        }
      }
    ]
  },
  {
    path: '/rules',
    component: Layout,
    redirect: '/rules/ban',
    name: 'Rules',
    meta: {
      title: '风控规则',
      icon: 'Shield'
    },
    children: [
      {
        path: 'ban',
        name: 'BanRules',
        component: () => import('@/views/rules/ban.vue'),
        meta: { 
          title: '封禁管理',
          icon: 'CircleClose'
        }
      },
      {
        path: 'whitelist',
        name: 'Whitelist',
        component: () => import('@/views/rules/whitelist.vue'),
        meta: { 
          title: '白名单',
          icon: 'CircleCheck'
        }
      },
      {
        path: 'analysis',
        name: 'BehaviorAnalysis',
        component: () => import('@/views/rules/analysis.vue'),
        meta: { 
          title: '行为分析',
          icon: 'TrendCharts'
        }
      }
    ]
  },
  {
    path: '/config',
    component: Layout,
    redirect: '/config/system',
    name: 'Config',
    meta: {
      title: '系统配置',
      icon: 'Setting'
    },
    children: [
      {
        path: 'system',
        name: 'SystemConfig',
        component: () => import('@/views/config/system.vue'),
        meta: { 
          title: '系统设置',
          icon: 'Tools'
        }
      },
      {
        path: 'scoring',
        name: 'ScoringConfig',
        component: () => import('@/views/config/scoring.vue'),
        meta: { 
          title: '评分规则',
          icon: 'Opportunity'
        }
      },
      {
        path: 'security',
        name: 'SecurityConfig',
        component: () => import('@/views/config/security.vue'),
        meta: { 
          title: '安全配置',
          icon: 'Lock'
        }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { 
      title: '登录',
      hidden: true
    }
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@/views/error/404.vue'),
    meta: { 
      title: '页面不存在',
      hidden: true
    }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - 防火墙控制器`
  }
  
  // 这里可以添加权限验证逻辑
  next()
})

export default router
