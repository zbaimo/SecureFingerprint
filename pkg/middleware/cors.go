package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		
		if origin != "" {
			// 允许的源
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		
		// 允许的请求头
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-Fingerprint")
		
		// 允许的请求方法
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		
		// 允许凭证
		c.Header("Access-Control-Allow-Credentials", "true")
		
		// 预检请求的缓存时间
		c.Header("Access-Control-Max-Age", "3600")

		// 处理预检请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
