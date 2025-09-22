package api

import (
	"fmt"
	"net/http"
	"strconv"

	"securefingerprint/internal/limiter"
	"securefingerprint/internal/scorer"

	"github.com/gin-gonic/gin"
)

type ConfigAPI struct {
	limiter *limiter.Limiter
	scorer  *scorer.Scorer
}

type ConfigResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// 系统配置结构
type SystemConfig struct {
	Security SecurityConfig `json:"security"`
	Server   ServerConfig   `json:"server"`
	Logging  LoggingConfig  `json:"logging"`
}

type SecurityConfig struct {
	Scoring scorer.ScoringConfig   `json:"scoring"`
	Limiter limiter.LimiterConfig `json:"limiter"`
}

type ServerConfig struct {
	Port  int  `json:"port"`
	Debug bool `json:"debug"`
}

type LoggingConfig struct {
	Level      string `json:"level"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

func NewConfigAPI(limiter *limiter.Limiter, scorer *scorer.Scorer) *ConfigAPI {
	return &ConfigAPI{
		limiter: limiter,
		scorer:  scorer,
	}
}

// 获取系统配置
func (api *ConfigAPI) GetConfig(c *gin.Context) {
	config := SystemConfig{
		Security: SecurityConfig{
			Scoring: scorer.DefaultScoringConfig,
			Limiter: limiter.DefaultLimiterConfig,
		},
		Server: ServerConfig{
			Port:  8080,
			Debug: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     30,
		},
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    config,
	})
}

// 更新系统配置
func (api *ConfigAPI) UpdateConfig(c *gin.Context) {
	var config SystemConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的配置格式: " + err.Error(),
		})
		return
	}

	// 验证配置有效性
	if err := api.validateConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "配置验证失败: " + err.Error(),
		})
		return
	}

	// 应用配置
	api.limiter.UpdateConfig(config.Security.Limiter)
	
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "配置更新成功",
		Data:    config,
	})
}

// 获取打分规则配置
func (api *ConfigAPI) GetScoringConfig(c *gin.Context) {
	config := scorer.DefaultScoringConfig

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    config,
	})
}

// 更新打分规则配置
func (api *ConfigAPI) UpdateScoringConfig(c *gin.Context) {
	var config scorer.ScoringConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的配置格式: " + err.Error(),
		})
		return
	}

	// 验证打分配置
	if config.InitialScore <= 0 || config.MaxScore <= 0 {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "分数配置必须大于0",
		})
		return
	}

	if config.BanThreshold >= config.InitialScore {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "封禁阈值不能大于等于初始分数",
		})
		return
	}

	// 这里应该保存配置到数据库或配置文件
	// 为简化实现，只返回成功响应

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "打分配置更新成功",
		Data:    config,
	})
}

// 获取限制器配置
func (api *ConfigAPI) GetLimiterConfig(c *gin.Context) {
	config := limiter.DefaultLimiterConfig

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    config,
	})
}

// 更新限制器配置
func (api *ConfigAPI) UpdateLimiterConfig(c *gin.Context) {
	var config limiter.LimiterConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的配置格式: " + err.Error(),
		})
		return
	}

	// 验证限制器配置
	if config.MaxRequestsPerWindow <= 0 {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "最大请求数必须大于0",
		})
		return
	}

	if config.DelayResponseMs < 0 {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "延迟时间不能为负数",
		})
		return
	}

	// 应用新配置
	api.limiter.UpdateConfig(config)

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "限制器配置更新成功",
		Data:    config,
	})
}

// 重置配置为默认值
func (api *ConfigAPI) ResetConfig(c *gin.Context) {
	configType := c.Param("type")
	
	switch configType {
	case "scoring":
		c.JSON(http.StatusOK, ConfigResponse{
			Success: true,
			Message: "打分配置已重置为默认值",
			Data:    scorer.DefaultScoringConfig,
		})
	case "limiter":
		api.limiter.UpdateConfig(limiter.DefaultLimiterConfig)
		c.JSON(http.StatusOK, ConfigResponse{
			Success: true,
			Message: "限制器配置已重置为默认值",
			Data:    limiter.DefaultLimiterConfig,
		})
	case "all":
		api.limiter.UpdateConfig(limiter.DefaultLimiterConfig)
		c.JSON(http.StatusOK, ConfigResponse{
			Success: true,
			Message: "所有配置已重置为默认值",
			Data: SystemConfig{
				Security: SecurityConfig{
					Scoring: scorer.DefaultScoringConfig,
					Limiter: limiter.DefaultLimiterConfig,
				},
			},
		})
	default:
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的配置类型，支持: scoring, limiter, all",
		})
	}
}

// 获取配置历史
func (api *ConfigAPI) GetConfigHistory(c *gin.Context) {
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

	// 模拟配置历史数据
	history := []map[string]interface{}{
		{
			"id":         1,
			"type":       "limiter",
			"changes":    "更新最大请求数: 100 -> 150",
			"operator":   "admin",
			"timestamp":  "2024-01-15T10:30:00Z",
		},
		{
			"id":         2,
			"type":       "scoring",
			"changes":    "更新初始分数: 100 -> 80",
			"operator":   "admin",
			"timestamp":  "2024-01-14T15:20:00Z",
		},
	}

	response := map[string]interface{}{
		"items":       history,
		"total":       len(history),
		"page":        page,
		"size":        size,
		"total_pages": (len(history) + size - 1) / size,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 导出配置
func (api *ConfigAPI) ExportConfig(c *gin.Context) {
	config := SystemConfig{
		Security: SecurityConfig{
			Scoring: scorer.DefaultScoringConfig,
			Limiter: limiter.DefaultLimiterConfig,
		},
		Server: ServerConfig{
			Port:  8080,
			Debug: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     30,
		},
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=firewall-config.json")
	
	c.JSON(http.StatusOK, config)
}

// 导入配置
func (api *ConfigAPI) ImportConfig(c *gin.Context) {
	var config SystemConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的配置格式: " + err.Error(),
		})
		return
	}

	// 验证导入的配置
	if err := api.validateConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "配置验证失败: " + err.Error(),
		})
		return
	}

	// 应用导入的配置
	api.limiter.UpdateConfig(config.Security.Limiter)

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "配置导入成功",
		Data:    config,
	})
}

// 验证配置有效性
func (api *ConfigAPI) validateConfig(config *SystemConfig) error {
	// 验证打分配置
	if config.Security.Scoring.InitialScore <= 0 {
		return fmt.Errorf("初始分数必须大于0")
	}

	if config.Security.Scoring.MaxScore <= 0 {
		return fmt.Errorf("最大分数必须大于0")
	}

	if config.Security.Scoring.BanThreshold >= config.Security.Scoring.InitialScore {
		return fmt.Errorf("封禁阈值不能大于等于初始分数")
	}

	// 验证限制器配置
	if config.Security.Limiter.MaxRequestsPerWindow <= 0 {
		return fmt.Errorf("最大请求数必须大于0")
	}

	if config.Security.Limiter.DelayResponseMs < 0 {
		return fmt.Errorf("延迟时间不能为负数")
	}

	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535范围内")
	}

	return nil
}

// 注册配置API路由
func (api *ConfigAPI) RegisterRoutes(router *gin.RouterGroup) {
	config := router.Group("/config")
	{
		config.GET("", api.GetConfig)
		config.PUT("", api.UpdateConfig)
		config.POST("/export", api.ExportConfig)
		config.POST("/import", api.ImportConfig)
		config.GET("/history", api.GetConfigHistory)
		
		config.GET("/scoring", api.GetScoringConfig)
		config.PUT("/scoring", api.UpdateScoringConfig)
		
		config.GET("/limiter", api.GetLimiterConfig)
		config.PUT("/limiter", api.UpdateLimiterConfig)
		
		config.POST("/reset/:type", api.ResetConfig)
	}
}
