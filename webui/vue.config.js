const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: true,
  
  // 构建输出目录
  outputDir: 'build',
  
  // 静态资源目录
  assetsDir: 'static',
  
  // 生产环境去除console
  configureWebpack: {
    optimization: {
      minimize: true
    }
  },
  
  // 开发服务器配置
  devServer: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true
      }
    }
  },
  
  // PWA配置
  pwa: {
    name: '防火墙控制器',
    themeColor: '#409EFF',
    manifestOptions: {
      background_color: '#ffffff'
    }
  }
})
