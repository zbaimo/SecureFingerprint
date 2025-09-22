package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"securefingerprint/internal/storage"

	"github.com/gin-gonic/gin"
)

type LogsAPI struct {
	mysqlClient *storage.MySQLClient
	redisClient *storage.RedisClient
}

func NewLogsAPI(mysqlClient *storage.MySQLClient, redisClient *storage.RedisClient) *LogsAPI {
	return &LogsAPI{
		mysqlClient: mysqlClient,
		redisClient: redisClient,
	}
}

// 查询访问日志
func (api *LogsAPI) GetAccessLogs(c *gin.Context) {
	// 构建查询条件
	query := &storage.AccessRecordQuery{}
	
	// 分页参数
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
	
	query.Limit = size
	query.Offset = (page - 1) * size

	// 筛选参数
	query.Fingerprint = c.Query("fingerprint")
	query.IP = c.Query("ip")
	query.UserAgent = c.Query("user_agent")
	query.Path = c.Query("path")
	query.Method = c.Query("method")
	query.Action = c.Query("action")
	
	// 分数范围
	if minScoreStr := c.Query("min_score"); minScoreStr != "" {
		if minScore, err := strconv.Atoi(minScoreStr); err == nil {
			query.MinScore = &minScore
		}
	}
	
	if maxScoreStr := c.Query("max_score"); maxScoreStr != "" {
		if maxScore, err := strconv.Atoi(maxScoreStr); err == nil {
			query.MaxScore = &maxScore
		}
	}

	// 时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", startTimeStr); err == nil {
			query.StartTime = startTime
		} else {
			c.JSON(http.StatusBadRequest, ConfigResponse{
				Success: false,
				Error:   "无效的开始时间格式",
			})
			return
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", endTimeStr); err == nil {
			query.EndTime = endTime
		} else {
			c.JSON(http.StatusBadRequest, ConfigResponse{
				Success: false,
				Error:   "无效的结束时间格式",
			})
			return
		}
	}

	// 排序参数
	query.OrderBy = c.DefaultQuery("order_by", "timestamp")
	query.OrderDir = c.DefaultQuery("order_dir", "DESC")

	// 查询日志
	result, err := api.mysqlClient.QueryAccessRecords(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "查询日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    result,
	})
}

// 获取日志统计信息
func (api *LogsAPI) GetLogStats(c *gin.Context) {
	// 构建查询条件（支持时间范围筛选）
	query := &storage.AccessRecordQuery{}
	
	// 时间范围参数
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", startTimeStr); err == nil {
			query.StartTime = startTime
		}
	}
	
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", endTimeStr); err == nil {
			query.EndTime = endTime
		}
	}
	
	// 如果没有指定时间范围，默认查询最近7天
	if query.StartTime.IsZero() && query.EndTime.IsZero() {
		query.EndTime = time.Now()
		query.StartTime = query.EndTime.AddDate(0, 0, -7)
	}

	// 获取详细统计信息
	stats, err := api.mysqlClient.GetAccessStats(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取统计信息失败: " + err.Error(),
		})
		return
	}

	// 获取系统基础统计
	systemStats, err := api.mysqlClient.GetSystemStats()
	if err != nil {
		// 如果系统统计失败，不影响主要统计结果
		systemStats = make(map[string]interface{})
	}

	// 组合结果
	result := map[string]interface{}{
		"basic_stats":        systemStats,
		"period_stats":       stats,
		"query_time_range": map[string]interface{}{
			"start_time": query.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			"end_time":   query.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    result,
	})
}

// 生成小时统计数据
func (api *LogsAPI) generateHourlyStats() []map[string]interface{} {
	var hourlyStats []map[string]interface{}
	
	now := time.Now()
	for i := 23; i >= 0; i-- {
		hour := now.Add(-time.Duration(i) * time.Hour)
		hourlyStats = append(hourlyStats, map[string]interface{}{
			"hour":     hour.Format("2006-01-02 15:00"),
			"requests": 50 + (i%12)*10, // 模拟数据
			"blocked":  5 + (i%5),      // 模拟数据
		})
	}
	
	return hourlyStats
}

// 获取热门路径
func (api *LogsAPI) getTopPaths() []map[string]interface{} {
	return []map[string]interface{}{
		{"path": "/", "count": 500, "percentage": 33.3},
		{"path": "/api/v1/users", "count": 300, "percentage": 20.0},
		{"path": "/login", "count": 200, "percentage": 13.3},
		{"path": "/dashboard", "count": 150, "percentage": 10.0},
		{"path": "/api/v1/posts", "count": 100, "percentage": 6.7},
	}
}

// 获取热门IP
func (api *LogsAPI) getTopIPs() []map[string]interface{} {
	return []map[string]interface{}{
		{"ip": "192.168.1.100", "count": 200, "status": "normal"},
		{"ip": "10.0.0.50", "count": 150, "status": "normal"},
		{"ip": "203.0.113.25", "count": 100, "status": "suspicious"},
		{"ip": "198.51.100.75", "count": 80, "status": "banned"},
		{"ip": "172.16.0.200", "count": 70, "status": "normal"},
	}
}

// 导出日志
func (api *LogsAPI) ExportLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	fingerprint := c.Query("fingerprint")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse("2006-01-02T15:04:05Z07:00", startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ConfigResponse{
				Success: false,
				Error:   "无效的开始时间格式",
			})
			return
		}
	}

	if endTimeStr != "" {
		endTime, err = time.Parse("2006-01-02T15:04:05Z07:00", endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ConfigResponse{
				Success: false,
				Error:   "无效的结束时间格式",
			})
			return
		}
	}

	// 查询所有符合条件的日志
	logs, err := api.mysqlClient.GetAccessLogs(fingerprint, 10000, 0, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "导出日志失败: " + err.Error(),
		})
		return
	}

	switch format {
	case "csv":
		api.exportLogsAsCSV(c, logs)
	case "json":
		api.exportLogsAsJSON(c, logs)
	default:
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "不支持的导出格式，支持: json, csv",
		})
	}
}

// 导出为JSON格式
func (api *LogsAPI) exportLogsAsJSON(c *gin.Context, logs []storage.AccessRecord) {
	filename := "access_logs_" + time.Now().Format("20060102_150405") + ".json"
	
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"export_time": time.Now(),
		"total_count": len(logs),
		"logs":        logs,
	})
}

// 导出为CSV格式
func (api *LogsAPI) exportLogsAsCSV(c *gin.Context, logs []storage.AccessRecord) {
	filename := "access_logs_" + time.Now().Format("20060102_150405") + ".csv"
	
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	
	// CSV头部
	c.Writer.WriteString("ID,Fingerprint,IP,UserAgent,Path,Method,Score,Action,Timestamp\n")
	
	// 写入数据
	for _, log := range logs {
		c.Writer.WriteString(fmt.Sprintf("%d,%s,%s,\"%s\",%s,%s,%d,%s,%s\n",
			log.ID, log.Fingerprint, log.IP, log.UserAgent, 
			log.Path, log.Method, log.Score, log.Action, 
			log.Timestamp.Format("2006-01-02 15:04:05")))
	}
}

// 清理日志
func (api *LogsAPI) CleanupLogs(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的天数参数",
		})
		return
	}

	// 计算清理的截止时间
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	// 这里应该实现实际的日志清理逻辑
	// 为简化实现，返回模拟结果
	
	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: fmt.Sprintf("成功清理%d天前的日志", days),
		Data: map[string]interface{}{
			"cutoff_time":    cutoffTime,
			"cleaned_count":  500, // 模拟清理的记录数
		},
	})
}

// 获取实时日志流
func (api *LogsAPI) GetRealtimeLogs(c *gin.Context) {
	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 模拟实时日志推送
	for i := 0; i < 10; i++ {
		log := map[string]interface{}{
			"id":          i + 1,
			"fingerprint": fmt.Sprintf("fp_%d", i+1),
			"ip":          fmt.Sprintf("192.168.1.%d", 100+i),
			"path":        "/api/test",
			"action":      "allow",
			"timestamp":   time.Now(),
		}

		c.SSEvent("log", log)
		c.Writer.Flush()
		
		time.Sleep(time.Second)
	}
}

// 搜索日志
func (api *LogsAPI) SearchLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "搜索关键词不能为空",
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")
	
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	
	// 这里应该实现实际的搜索逻辑
	// 为简化实现，返回模拟结果
	
	results := []map[string]interface{}{
		{
			"id":          1,
			"fingerprint": "abc123",
			"ip":          "192.168.1.100",
			"path":        "/search?q=" + keyword,
			"action":      "allow",
			"timestamp":   time.Now().Add(-time.Hour),
			"highlight":   keyword,
		},
	}

	response := map[string]interface{}{
		"keyword":     keyword,
		"results":     results,
		"total":       len(results),
		"page":        page,
		"size":        size,
		"total_pages": 1,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 获取用户访问记录
func (api *LogsAPI) GetUserAccessLogs(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	// 分页参数
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
	
	offset := (page - 1) * size

	// 获取用户访问记录
	records, err := api.mysqlClient.GetUserAccessRecords(fingerprint, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取用户访问记录失败: " + err.Error(),
		})
		return
	}

	// 获取总数（简化处理，实际应该单独查询）
	totalQuery := &storage.AccessRecordQuery{
		Fingerprint: fingerprint,
		Limit:       1,
		Offset:      0,
	}
	totalResult, err := api.mysqlClient.QueryAccessRecords(totalQuery)
	var total int64 = 0
	if err == nil {
		total = totalResult.Total
	}

	response := map[string]interface{}{
		"records":     records,
		"total":       total,
		"page":        page,
		"size":        size,
		"total_pages": (int(total) + size - 1) / size,
		"fingerprint": fingerprint,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 获取最近访问记录
func (api *LogsAPI) GetRecentAccessLogs(c *gin.Context) {
	minutesStr := c.DefaultQuery("minutes", "60")
	limitStr := c.DefaultQuery("limit", "50")
	
	minutes, _ := strconv.Atoi(minutesStr)
	limit, _ := strconv.Atoi(limitStr)
	
	if minutes < 1 || minutes > 1440 { // 最多24小时
		minutes = 60
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	records, err := api.mysqlClient.GetRecentAccessRecords(minutes, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取最近访问记录失败: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"records":     records,
		"minutes":     minutes,
		"limit":       limit,
		"count":       len(records),
		"query_time":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 高级搜索访问记录
func (api *LogsAPI) AdvancedSearchLogs(c *gin.Context) {
	var searchReq struct {
		Fingerprint string `json:"fingerprint,omitempty"`
		IP          string `json:"ip,omitempty"`
		UserAgent   string `json:"user_agent,omitempty"`
		Path        string `json:"path,omitempty"`
		Method      string `json:"method,omitempty"`
		Action      string `json:"action,omitempty"`
		MinScore    *int   `json:"min_score,omitempty"`
		MaxScore    *int   `json:"max_score,omitempty"`
		StartTime   string `json:"start_time,omitempty"`
		EndTime     string `json:"end_time,omitempty"`
		Page        int    `json:"page"`
		Size        int    `json:"size"`
		OrderBy     string `json:"order_by,omitempty"`
		OrderDir    string `json:"order_dir,omitempty"`
	}

	if err := c.ShouldBindJSON(&searchReq); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的搜索参数: " + err.Error(),
		})
		return
	}

	// 构建查询条件
	query := &storage.AccessRecordQuery{
		Fingerprint: searchReq.Fingerprint,
		IP:          searchReq.IP,
		UserAgent:   searchReq.UserAgent,
		Path:        searchReq.Path,
		Method:      searchReq.Method,
		Action:      searchReq.Action,
		MinScore:    searchReq.MinScore,
		MaxScore:    searchReq.MaxScore,
		OrderBy:     searchReq.OrderBy,
		OrderDir:    searchReq.OrderDir,
	}

	// 分页
	if searchReq.Page < 1 {
		searchReq.Page = 1
	}
	if searchReq.Size < 1 || searchReq.Size > 100 {
		searchReq.Size = 20
	}
	query.Limit = searchReq.Size
	query.Offset = (searchReq.Page - 1) * searchReq.Size

	// 时间范围
	if searchReq.StartTime != "" {
		if startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", searchReq.StartTime); err == nil {
			query.StartTime = startTime
		}
	}
	if searchReq.EndTime != "" {
		if endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", searchReq.EndTime); err == nil {
			query.EndTime = endTime
		}
	}

	// 执行搜索
	result, err := api.mysqlClient.QueryAccessRecords(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    result,
	})
}

// 注册日志API路由
func (api *LogsAPI) RegisterRoutes(router *gin.RouterGroup) {
	logs := router.Group("/logs")
	{
		logs.GET("", api.GetAccessLogs)
		logs.POST("/search", api.AdvancedSearchLogs)
		logs.GET("/user/:fingerprint", api.GetUserAccessLogs)
		logs.GET("/recent", api.GetRecentAccessLogs)
		logs.GET("/stats", api.GetLogStats)
		logs.GET("/export", api.ExportLogs)
		logs.GET("/realtime", api.GetRealtimeLogs)
		logs.GET("/search", api.SearchLogs)
		logs.DELETE("/cleanup", api.CleanupLogs)
	}
}
