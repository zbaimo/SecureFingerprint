import request from '@/utils/request'

// 获取系统配置
export function getSystemConfig() {
  return request({
    url: '/config',
    method: 'get'
  })
}

// 更新系统配置
export function updateSystemConfig(data) {
  return request({
    url: '/config',
    method: 'put',
    data
  })
}

// 获取评分配置
export function getScoringConfig() {
  return request({
    url: '/config/scoring',
    method: 'get'
  })
}

// 更新评分配置
export function updateScoringConfig(data) {
  return request({
    url: '/config/scoring',
    method: 'put',
    data
  })
}

// 获取限制器配置
export function getLimiterConfig() {
  return request({
    url: '/config/limiter',
    method: 'get'
  })
}

// 更新限制器配置
export function updateLimiterConfig(data) {
  return request({
    url: '/config/limiter',
    method: 'put',
    data
  })
}

// 重置配置
export function resetConfig(type) {
  return request({
    url: `/config/reset/${type}`,
    method: 'post'
  })
}

// 导出配置
export function exportConfig() {
  return request({
    url: '/config/export',
    method: 'post',
    responseType: 'blob'
  })
}

// 导入配置
export function importConfig(data) {
  return request({
    url: '/config/import',
    method: 'post',
    data
  })
}

// 获取配置历史
export function getConfigHistory(params) {
  return request({
    url: '/config/history',
    method: 'get',
    params
  })
}
