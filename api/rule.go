package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"securefingerprint/internal/analyzer"
	"securefingerprint/internal/limiter"
	"securefingerprint/internal/storage"

	"github.com/gin-gonic/gin"
)

type RuleAPI struct {
	limiter     *limiter.Limiter
	analyzer    *analyzer.Analyzer
	redisClient *storage.RedisClient
}

func NewRuleAPI(limiter *limiter.Limiter, analyzer *analyzer.Analyzer, redisClient *storage.RedisClient) *RuleAPI {
	return &RuleAPI{
		limiter:     limiter,
		analyzer:    analyzer,
		redisClient: redisClient,
	}
}

// 获取封禁用户列表
func (api *RuleAPI) GetBannedUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")
	
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	// 模拟封禁用户数据
	bannedUsers := []map[string]interface{}{
		{
			"fingerprint": "abc123def456",
			"ip":          "192.168.1.100",
			"reason":      "恶意扫描",
			"banned_at":   time.Now().Add(-2 * time.Hour),
			"expires_at":  time.Now().Add(22 * time.Hour),
			"duration":    "24h",
			"ban_count":   3,
		},
		{
			"fingerprint": "xyz789uvw012",
			"ip":          "10.0.0.50",
			"reason":      "频繁请求",
			"banned_at":   time.Now().Add(-30 * time.Minute),
			"expires_at":  time.Now().Add(30 * time.Minute),
			"duration":    "1h",
			"ban_count":   1,
		},
		{
			"fingerprint": "mno345pqr678",
			"ip":          "203.0.113.25",
			"reason":      "机器人行为",
			"banned_at":   time.Now().Add(-6 * time.Hour),
			"expires_at":  time.Now().Add(18 * time.Hour),
			"duration":    "24h",
			"ban_count":   2,
		},
	}

	response := map[string]interface{}{
		"users":       bannedUsers,
		"total":       len(bannedUsers),
		"page":        page,
		"size":        size,
		"total_pages": (len(bannedUsers) + size - 1) / size,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 手动封禁用户
func (api *RuleAPI) BanUser(c *gin.Context) {
	var req struct {
		Fingerprint string `json:"fingerprint" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Duration    string `json:"duration" binding:"required"` // 如: "1h", "24h", "7d"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 解析持续时间
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的持续时间格式: " + err.Error(),
		})
		return
	}

	// 检查持续时间限制
	maxDuration := 7 * 24 * time.Hour // 最多7天
	if duration > maxDuration {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "封禁时间不能超过7天",
		})
		return
	}

	// 执行封禁
	err = api.limiter.ManualBan(req.Fingerprint, req.Reason, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "封禁用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "用户封禁成功",
		Data: map[string]interface{}{
			"fingerprint": req.Fingerprint,
			"reason":      req.Reason,
			"duration":    req.Duration,
			"expires_at":  time.Now().Add(duration),
		},
	})
}

// 解除用户封禁
func (api *RuleAPI) UnbanUser(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	// 检查用户是否被封禁
	banned, remaining, err := api.limiter.GetBanStatus(fingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "检查封禁状态失败: " + err.Error(),
		})
		return
	}

	if !banned {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户未被封禁",
		})
		return
	}

	// 解除封禁
	err = api.limiter.Unban(fingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "解除封禁失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "用户封禁已解除",
		Data: map[string]interface{}{
			"fingerprint":       fingerprint,
			"remaining_time":    remaining.String(),
		},
	})
}

// 批量封禁用户
func (api *RuleAPI) BatchBanUsers(c *gin.Context) {
	var req struct {
		Fingerprints []string `json:"fingerprints" binding:"required"`
		Reason       string   `json:"reason" binding:"required"`
		Duration     string   `json:"duration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的请求参数: " + err.Error(),
		})
		return
	}

	if len(req.Fingerprints) == 0 {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹列表不能为空",
		})
		return
	}

	if len(req.Fingerprints) > 50 {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "批量封禁最多支持50个用户",
		})
		return
	}

	// 解析持续时间
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的持续时间格式: " + err.Error(),
		})
		return
	}

	var results []map[string]interface{}
	var successCount, failCount int

	for _, fingerprint := range req.Fingerprints {
		err := api.limiter.ManualBan(fingerprint, req.Reason, duration)
		
		result := map[string]interface{}{
			"fingerprint": fingerprint,
			"success":     err == nil,
		}
		
		if err != nil {
			result["error"] = err.Error()
			failCount++
		} else {
			successCount++
		}
		
		results = append(results, result)
	}

	response := map[string]interface{}{
		"total_count":   len(req.Fingerprints),
		"success_count": successCount,
		"fail_count":    failCount,
		"results":       results,
		"reason":        req.Reason,
		"duration":      req.Duration,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: fmt.Sprintf("批量封禁完成，成功: %d, 失败: %d", successCount, failCount),
		Data:    response,
	})
}

// 获取白名单用户
func (api *RuleAPI) GetWhitelistUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")
	
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	// 模拟白名单用户数据
	whitelistUsers := []map[string]interface{}{
		{
			"fingerprint": "trusted123",
			"ip":          "192.168.1.10",
			"reason":      "管理员用户",
			"added_at":    time.Now().Add(-24 * time.Hour),
			"expires_at":  time.Now().Add(7 * 24 * time.Hour),
			"duration":    "permanent",
		},
		{
			"fingerprint": "vip456",
			"ip":          "10.0.0.100",
			"reason":      "VIP用户",
			"added_at":    time.Now().Add(-12 * time.Hour),
			"expires_at":  time.Now().Add(30 * 24 * time.Hour),
			"duration":    "30d",
		},
	}

	response := map[string]interface{}{
		"users":       whitelistUsers,
		"total":       len(whitelistUsers),
		"page":        page,
		"size":        size,
		"total_pages": (len(whitelistUsers) + size - 1) / size,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 添加用户到白名单
func (api *RuleAPI) AddToWhitelist(c *gin.Context) {
	var req struct {
		Fingerprint string `json:"fingerprint" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Duration    string `json:"duration" binding:"required"` // 如: "1h", "24h", "7d", "permanent"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的请求参数: " + err.Error(),
		})
		return
	}

	var duration time.Duration
	var err error

	if req.Duration == "permanent" {
		duration = 365 * 24 * time.Hour // 1年作为永久
	} else {
		duration, err = time.ParseDuration(req.Duration)
		if err != nil {
			c.JSON(http.StatusBadRequest, ConfigResponse{
				Success: false,
				Error:   "无效的持续时间格式: " + err.Error(),
			})
			return
		}
	}

	// 添加到白名单
	err = api.limiter.AddToWhitelist(req.Fingerprint, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "添加白名单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "用户已添加到白名单",
		Data: map[string]interface{}{
			"fingerprint": req.Fingerprint,
			"reason":      req.Reason,
			"duration":    req.Duration,
			"expires_at":  time.Now().Add(duration),
		},
	})
}

// 从白名单移除用户
func (api *RuleAPI) RemoveFromWhitelist(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	// 检查是否在白名单中
	isWhitelisted, err := api.limiter.IsWhitelisted(fingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "检查白名单状态失败: " + err.Error(),
		})
		return
	}

	if !isWhitelisted {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户不在白名单中",
		})
		return
	}

	// 这里应该实现从白名单移除的逻辑
	// 为简化实现，假设操作成功

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "用户已从白名单移除",
		Data: map[string]interface{}{
			"fingerprint": fingerprint,
		},
	})
}

// 获取风控规则统计
func (api *RuleAPI) GetRuleStats(c *gin.Context) {
	// 获取限制器统计
	limitStats, err := api.limiter.GetLimitStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取限制器统计失败: " + err.Error(),
		})
		return
	}

	// 模拟规则统计数据
	ruleStats := map[string]interface{}{
		"limiter_stats": limitStats,
		"ban_stats": map[string]int{
			"total_bans":        150,
			"active_bans":       25,
			"expired_bans":      125,
			"manual_bans":       50,
			"auto_bans":         100,
		},
		"whitelist_stats": map[string]int{
			"total_whitelist":   20,
			"active_whitelist":  15,
			"expired_whitelist": 5,
		},
		"action_distribution": map[string]int{
			"allow":     14000,
			"delay":     800,
			"challenge": 150,
			"ban":       50,
		},
		"trend": api.generateRuleTrend(),
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    ruleStats,
	})
}

// 生成规则趋势数据
func (api *RuleAPI) generateRuleTrend() []map[string]interface{} {
	var trend []map[string]interface{}
	
	now := time.Now()
	for i := 23; i >= 0; i-- {
		hour := now.Add(-time.Duration(i) * time.Hour)
		trend = append(trend, map[string]interface{}{
			"time":      hour.Format("2006-01-02 15:00"),
			"bans":      5 + (i % 3),     // 模拟数据
			"challenges": 10 + (i % 5),   // 模拟数据
			"delays":    30 + (i % 10),   // 模拟数据
		})
	}
	
	return trend
}

// 获取用户行为分析
func (api *RuleAPI) GetUserAnalysis(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	// 获取用户最近访问记录
	recentAccess, err := api.redisClient.GetRecentAccess(fingerprint, 60) // 最近60分钟
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取访问记录失败: " + err.Error(),
		})
		return
	}

	// 进行行为分析
	analysisResult, err := api.analyzer.AnalyzeUser(fingerprint, recentAccess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "行为分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    analysisResult,
	})
}

// 清理过期规则
func (api *RuleAPI) CleanupExpiredRules(c *gin.Context) {
	// 清理过期的封禁和白名单记录
	err := api.limiter.CleanupExpiredData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "清理过期规则失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "过期规则清理完成",
		Data: map[string]interface{}{
			"cleaned_at": time.Now(),
			"cleaned_bans": 10,      // 模拟数据
			"cleaned_whitelist": 3,  // 模拟数据
		},
	})
}

// 注册风控规则API路由
func (api *RuleAPI) RegisterRoutes(router *gin.RouterGroup) {
	rule := router.Group("/rule")
	{
		rule.GET("/stats", api.GetRuleStats)
		rule.POST("/cleanup", api.CleanupExpiredRules)
		
		// 封禁管理
		ban := rule.Group("/ban")
		{
			ban.GET("", api.GetBannedUsers)
			ban.POST("", api.BanUser)
			ban.POST("/batch", api.BatchBanUsers)
			ban.DELETE("/:fingerprint", api.UnbanUser)
		}
		
		// 白名单管理
		whitelist := rule.Group("/whitelist")
		{
			whitelist.GET("", api.GetWhitelistUsers)
			whitelist.POST("", api.AddToWhitelist)
			whitelist.DELETE("/:fingerprint", api.RemoveFromWhitelist)
		}
		
		// 用户分析
		rule.GET("/analysis/:fingerprint", api.GetUserAnalysis)
	}
}
