import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'

// 创建axios实例
const service = axios.create({
  baseURL: '/api/v1', // API基础路径
  timeout: 30000 // 请求超时时间
})

// 请求拦截器
service.interceptors.request.use(
  config => {
    // 在发送请求之前做些什么
    // 可以在这里添加token等认证信息
    // if (store.getters.token) {
    //   config.headers['Authorization'] = `Bearer ${getToken()}`
    // }
    return config
  },
  error => {
    // 对请求错误做些什么
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  response => {
    const res = response.data

    // 如果是文件下载，直接返回
    if (response.config.responseType === 'blob') {
      return response.data
    }

    // 如果响应不是预期的格式，直接返回
    if (typeof res !== 'object' || res === null) {
      return res
    }

    // 检查业务状态码
    if (res.success === false) {
      // 业务错误
      ElMessage({
        message: res.error || res.message || '操作失败',
        type: 'error',
        duration: 5 * 1000
      })

      // 如果是认证错误，可以在这里处理登出逻辑
      if (res.code === 401 || res.code === 403) {
        ElMessageBox.confirm(
          '您的登录状态已过期，请重新登录',
          '系统提示',
          {
            confirmButtonText: '重新登录',
            cancelButtonText: '取消',
            type: 'warning'
          }
        ).then(() => {
          // 重新登录逻辑
          // store.dispatch('user/resetToken').then(() => {
          //   location.reload()
          // })
        })
      }

      return Promise.reject(new Error(res.error || res.message || '操作失败'))
    }

    // 成功响应
    return res
  },
  error => {
    console.error('响应错误:', error)
    
    let message = '网络错误'
    
    if (error.response) {
      // 服务器响应了错误状态码
      const { status, data } = error.response
      
      switch (status) {
        case 400:
          message = data?.error || '请求参数错误'
          break
        case 401:
          message = '未授权，请登录'
          break
        case 403:
          message = '拒绝访问'
          break
        case 404:
          message = '请求资源不存在'
          break
        case 408:
          message = '请求超时'
          break
        case 500:
          message = '服务器内部错误'
          break
        case 501:
          message = '服务未实现'
          break
        case 502:
          message = '网关错误'
          break
        case 503:
          message = '服务不可用'
          break
        case 504:
          message = '网关超时'
          break
        case 505:
          message = 'HTTP版本不受支持'
          break
        default:
          message = data?.error || data?.message || `连接错误${status}`
      }
    } else if (error.request) {
      // 请求已经成功发起，但没有收到响应
      message = '网络连接超时'
    } else {
      // 发送请求时出了点问题
      message = error.message || '请求配置错误'
    }

    ElMessage({
      message,
      type: 'error',
      duration: 5 * 1000
    })

    return Promise.reject(error)
  }
)

export default service
