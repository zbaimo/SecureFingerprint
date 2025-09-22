import request from '@/utils/request'

// 获取用户列表
export function getUserList(params) {
  return request({
    url: '/users',
    method: 'get',
    params
  })
}

// 获取用户详情
export function getUserDetail(fingerprint) {
  return request({
    url: `/users/${fingerprint}`,
    method: 'get'
  })
}

// 获取用户分数
export function getUserScore(fingerprint) {
  return request({
    url: `/score/${fingerprint}`,
    method: 'get'
  })
}

// 重置用户分数
export function resetUserScore(fingerprint) {
  return request({
    url: `/score/${fingerprint}/reset`,
    method: 'post'
  })
}

// 调整用户分数
export function adjustUserScore(fingerprint, data) {
  return request({
    url: `/score/${fingerprint}/adjust`,
    method: 'post',
    data
  })
}

// 获取用户分数历史
export function getUserScoreHistory(fingerprint, params) {
  return request({
    url: `/score/${fingerprint}/history`,
    method: 'get',
    params
  })
}

// 获取分数统计
export function getScoreStats() {
  return request({
    url: '/score/stats',
    method: 'get'
  })
}

// 获取低分用户
export function getLowScoreUsers(params) {
  return request({
    url: '/score/low-score-users',
    method: 'get',
    params
  })
}

// 批量分数操作
export function batchScoreOperation(data) {
  return request({
    url: '/score/batch',
    method: 'post',
    data
  })
}
