import request from '@/utils/request'

// 获取系统信息
export function getSystemInfo() {
  return request({
    url: '/system/info',
    method: 'get'
  })
}

// 获取健康检查
export function getHealthCheck() {
  return request({
    url: '/system/health',
    method: 'get'
  })
}

// 获取代理信息
export function getProxyInfo() {
  return request({
    url: '/proxy/info',
    method: 'get'
  })
}

// 获取代理配置
export function getProxyConfig() {
  return request({
    url: '/proxy/config',
    method: 'get'
  })
}

// 测试代理检测
export function testProxyDetection(data) {
  return request({
    url: '/proxy/test',
    method: 'post',
    data
  })
}

// 获取代理统计
export function getProxyStats() {
  return request({
    url: '/proxy/stats',
    method: 'get'
  })
}

// 验证代理配置
export function validateProxyConfig(data) {
  return request({
    url: '/proxy/validate',
    method: 'post',
    data
  })
}
