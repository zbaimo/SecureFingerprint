package api

import (
	"net/http"

	"firewall-controller/internal/collector"

	"github.com/gin-gonic/gin"
)

type ProxyAPI struct {
	collector *collector.Collector
	detector  *collector.ProxyDetector
}

func NewProxyAPI(collector *collector.Collector) *ProxyAPI {
	// 使用默认代理配置创建检测器
	detector, _ := collector.NewProxyDetector(collector.DefaultProxyConfig)
	
	return &ProxyAPI{
		collector: collector,
		detector:  detector,
	}
}

// 获取当前请求的代理信息
func (api *ProxyAPI) GetProxyInfo(c *gin.Context) {
	// 采集访问信息
	accessInfo := api.collector.CollectFromRequest(c.Request)
	
	// 获取代理检测报告
	proxyReport := api.detector.GetProxyReport(c.Request)
	
	// 获取代理摘要
	proxySummary := api.collector.GetProxySummary(accessInfo)

	response := map[string]interface{}{
		"client_info": map[string]interface{}{
			"ip":             accessInfo.IP,
			"original_ip":    accessInfo.OriginalIP,
			"user_agent":     accessInfo.UserAgent,
			"is_behind_proxy": accessInfo.IsBehindProxy,
		},
		"proxy_detection": proxyReport,
		"proxy_summary":   proxySummary,
		"access_info":     accessInfo,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 获取代理配置信息
func (api *ProxyAPI) GetProxyConfig(c *gin.Context) {
	config := collector.DefaultProxyConfig

	response := map[string]interface{}{
		"trusted_proxies":    config.TrustedProxies,
		"trusted_headers":    config.TrustedHeaders,
		"header_priority":    config.HeaderPriority,
		"skip_private_ranges": config.SkipPrivateRanges,
		"max_proxy_depth":    config.MaxProxyDepth,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 测试代理检测
func (api *ProxyAPI) TestProxyDetection(c *gin.Context) {
	var testReq struct {
		Headers map[string]string `json:"headers"`
		RemoteAddr string         `json:"remote_addr"`
	}

	if err := c.ShouldBindJSON(&testReq); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的测试参数: " + err.Error(),
		})
		return
	}

	// 创建模拟请求
	req := c.Request.Clone(c.Request.Context())
	
	// 设置测试头
	for key, value := range testReq.Headers {
		req.Header.Set(key, value)
	}
	
	// 设置远程地址
	if testReq.RemoteAddr != "" {
		req.RemoteAddr = testReq.RemoteAddr
	}

	// 进行检测
	accessInfo := api.collector.CollectFromRequest(req)
	proxyReport := api.detector.GetProxyReport(req)

	response := map[string]interface{}{
		"test_input": testReq,
		"detection_result": map[string]interface{}{
			"extracted_ip":    accessInfo.IP,
			"original_ip":     accessInfo.OriginalIP,
			"proxy_chain":     accessInfo.ProxyChain,
			"is_behind_proxy": accessInfo.IsBehindProxy,
			"proxy_headers":   accessInfo.ProxyHeaders,
		},
		"detailed_report": proxyReport,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 获取代理统计信息
func (api *ProxyAPI) GetProxyStats(c *gin.Context) {
	// 这里应该从数据库获取实际统计数据
	// 为演示目的，返回模拟数据
	stats := map[string]interface{}{
		"total_requests": 10000,
		"proxy_requests": 7500,
		"direct_requests": 2500,
		"proxy_types": map[string]int{
			"nginx":      4000,
			"cloudflare": 2000,
			"other_cdn":  1000,
			"unknown":    500,
		},
		"top_proxy_ips": []map[string]interface{}{
			{"ip": "192.168.1.100", "count": 2000, "type": "nginx"},
			{"ip": "10.0.0.50", "count": 1500, "type": "load_balancer"},
			{"ip": "172.16.0.10", "count": 1000, "type": "reverse_proxy"},
		},
		"header_usage": map[string]int{
			"X-Forwarded-For": 6000,
			"X-Real-IP":       4000,
			"CF-Connecting-IP": 2000,
			"True-Client-IP":   500,
		},
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    stats,
	})
}

// 验证代理配置
func (api *ProxyAPI) ValidateProxyConfig(c *gin.Context) {
	var configReq collector.ProxyConfig

	if err := c.ShouldBindJSON(&configReq); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的配置格式: " + err.Error(),
		})
		return
	}

	// 尝试创建检测器以验证配置
	detector, err := collector.NewProxyDetector(configReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "配置验证失败: " + err.Error(),
		})
		return
	}

	// 验证成功，返回配置摘要
	validation := map[string]interface{}{
		"valid": true,
		"summary": map[string]interface{}{
			"trusted_proxy_count": len(configReq.TrustedProxies),
			"trusted_header_count": len(configReq.TrustedHeaders),
			"max_proxy_depth":     configReq.MaxProxyDepth,
			"skip_private_ranges": configReq.SkipPrivateRanges,
		},
		"warnings": api.validateConfigWarnings(&configReq),
	}

	// 更新检测器（在实际应用中可能需要持久化配置）
	api.detector = detector

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    validation,
	})
}

// 获取配置警告
func (api *ProxyAPI) validateConfigWarnings(config *collector.ProxyConfig) []string {
	var warnings []string

	// 检查是否包含基本的可信代理
	hasLocalhost := false
	hasPrivateRanges := false
	
	for _, proxy := range config.TrustedProxies {
		if proxy == "127.0.0.1/32" || proxy == "::1/128" {
			hasLocalhost = true
		}
		if proxy == "10.0.0.0/8" || proxy == "172.16.0.0/12" || proxy == "192.168.0.0/16" {
			hasPrivateRanges = true
		}
	}

	if !hasLocalhost {
		warnings = append(warnings, "建议添加localhost到可信代理列表")
	}

	if !hasPrivateRanges {
		warnings = append(warnings, "建议添加内网IP段到可信代理列表")
	}

	// 检查头优先级配置
	if len(config.HeaderPriority) == 0 {
		warnings = append(warnings, "未配置头优先级，将使用默认优先级")
	}

	// 检查代理深度
	if config.MaxProxyDepth > 20 {
		warnings = append(warnings, "代理深度过大可能影响性能")
	}

	// 检查可信头列表
	if len(config.TrustedHeaders) == 0 {
		warnings = append(warnings, "未配置可信头列表，可能无法正确检测代理")
	}

	return warnings
}

// 注册代理API路由
func (api *ProxyAPI) RegisterRoutes(router *gin.RouterGroup) {
	proxy := router.Group("/proxy")
	{
		proxy.GET("/info", api.GetProxyInfo)
		proxy.GET("/config", api.GetProxyConfig)
		proxy.POST("/test", api.TestProxyDetection)
		proxy.GET("/stats", api.GetProxyStats)
		proxy.POST("/validate", api.ValidateProxyConfig)
	}
}
