import request from '@/utils/request'

// 获取访问日志
export function getAccessLogs(params) {
  return request({
    url: '/logs',
    method: 'get',
    params
  })
}

// 高级搜索访问日志
export function advancedSearchLogs(data) {
  return request({
    url: '/logs/search',
    method: 'post',
    data
  })
}

// 获取用户访问记录
export function getUserAccessLogs(fingerprint, params) {
  return request({
    url: `/logs/user/${fingerprint}`,
    method: 'get',
    params
  })
}

// 获取最近访问记录
export function getRecentAccessLogs(params) {
  return request({
    url: '/logs/recent',
    method: 'get',
    params
  })
}

// 获取日志统计信息
export function getLogStats(params) {
  return request({
    url: '/logs/stats',
    method: 'get',
    params
  })
}

// 导出日志
export function exportLogs(params) {
  return request({
    url: '/logs/export',
    method: 'get',
    params,
    responseType: 'blob'
  }).then(response => {
    const blob = new Blob([response], { type: 'application/json' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `access_logs_${new Date().getTime()}.json`
    link.click()
    window.URL.revokeObjectURL(url)
  })
}

// 获取实时日志
export function getRealtimeLogs() {
  return request({
    url: '/logs/realtime',
    method: 'get'
  })
}

// 搜索日志
export function searchLogs(params) {
  return request({
    url: '/logs/search',
    method: 'get',
    params
  })
}

// 清理过期日志
export function cleanupLogs(params) {
  return request({
    url: '/logs/cleanup',
    method: 'delete',
    params
  })
}
