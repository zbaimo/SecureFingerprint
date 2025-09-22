import request from '@/utils/request'

// 获取封禁用户列表
export function getBannedUsers(params) {
  return request({
    url: '/rule/ban',
    method: 'get',
    params
  })
}

// 封禁用户
export function banUser(data) {
  return request({
    url: '/rule/ban',
    method: 'post',
    data
  })
}

// 批量封禁用户
export function batchBanUsers(data) {
  return request({
    url: '/rule/ban/batch',
    method: 'post',
    data
  })
}

// 解除封禁
export function unbanUser(fingerprint) {
  return request({
    url: `/rule/ban/${fingerprint}`,
    method: 'delete'
  })
}

// 获取白名单用户
export function getWhitelistUsers(params) {
  return request({
    url: '/rule/whitelist',
    method: 'get',
    params
  })
}

// 添加到白名单
export function addToWhitelist(data) {
  return request({
    url: '/rule/whitelist',
    method: 'post',
    data
  })
}

// 从白名单移除
export function removeFromWhitelist(fingerprint) {
  return request({
    url: `/rule/whitelist/${fingerprint}`,
    method: 'delete'
  })
}

// 获取用户行为分析
export function getUserAnalysis(fingerprint) {
  return request({
    url: `/rule/analysis/${fingerprint}`,
    method: 'get'
  })
}

// 获取风控规则统计
export function getRuleStats() {
  return request({
    url: '/rule/stats',
    method: 'get'
  })
}

// 清理过期规则
export function cleanupExpiredRules() {
  return request({
    url: '/rule/cleanup',
    method: 'post'
  })
}
