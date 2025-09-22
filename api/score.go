package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"securefingerprint/internal/scorer"
	"securefingerprint/internal/storage"

	"github.com/gin-gonic/gin"
)

type ScoreAPI struct {
	scorer      *scorer.Scorer
	redisClient *storage.RedisClient
}

func NewScoreAPI(scorer *scorer.Scorer, redisClient *storage.RedisClient) *ScoreAPI {
	return &ScoreAPI{
		scorer:      scorer,
		redisClient: redisClient,
	}
}

// 获取用户分数信息
func (api *ScoreAPI) GetUserScore(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	userScore, err := api.redisClient.GetUserScore(fingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取用户分数失败: " + err.Error(),
		})
		return
	}

	// 获取分数趋势
	trend, _ := api.scorer.GetScoreTrend(fingerprint, 24)

	response := map[string]interface{}{
		"fingerprint":    fingerprint,
		"current_score":  userScore.Score,
		"last_seen":      userScore.LastSeen,
		"request_count":  userScore.RequestCount,
		"score_trend":    trend,
		"status":         api.getScoreStatus(userScore.Score),
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 获取分数状态
func (api *ScoreAPI) getScoreStatus(score int) string {
	if score >= 80 {
		return "excellent"
	} else if score >= 60 {
		return "good"
	} else if score >= 30 {
		return "warning"
	} else if score >= 10 {
		return "danger"
	} else {
		return "banned"
	}
}

// 重置用户分数
func (api *ScoreAPI) ResetUserScore(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	err := api.scorer.ResetUserScore(fingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "重置用户分数失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "用户分数已重置",
		Data: map[string]interface{}{
			"fingerprint": fingerprint,
			"new_score":   scorer.DefaultScoringConfig.InitialScore,
		},
	})
}

// 调整用户分数
func (api *ScoreAPI) AdjustUserScore(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	var req struct {
		Adjustment int    `json:"adjustment" binding:"required"`
		Reason     string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 获取当前分数
	userScore, err := api.redisClient.GetUserScore(fingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取用户分数失败: " + err.Error(),
		})
		return
	}

	oldScore := userScore.Score
	newScore := oldScore + req.Adjustment

	// 确保分数在合理范围内
	if newScore > scorer.DefaultScoringConfig.MaxScore {
		newScore = scorer.DefaultScoringConfig.MaxScore
	}
	if newScore < -50 {
		newScore = -50
	}

	// 更新分数
	userScore.Score = newScore
	userScore.LastSeen = time.Now()
	
	err = api.redisClient.UpdateUserScore(fingerprint, userScore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "更新用户分数失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: "用户分数调整成功",
		Data: map[string]interface{}{
			"fingerprint":  fingerprint,
			"old_score":    oldScore,
			"new_score":    newScore,
			"adjustment":   req.Adjustment,
			"reason":       req.Reason,
		},
	})
}

// 获取分数统计信息
func (api *ScoreAPI) GetScoreStats(c *gin.Context) {
	stats, err := api.scorer.GetScoreStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取分数统计失败: " + err.Error(),
		})
		return
	}

	// 添加分数分布信息
	scoreDistribution := map[string]int{
		"excellent": 800,  // 80-100分
		"good":      400,  // 60-79分
		"warning":   200,  // 30-59分
		"danger":    80,   // 10-29分
		"banned":    20,   // 0-9分
	}

	response := map[string]interface{}{
		"basic_stats":        stats,
		"score_distribution": scoreDistribution,
		"average_score":      75.5,
		"score_trend":        api.generateScoreTrend(),
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 生成分数趋势数据
func (api *ScoreAPI) generateScoreTrend() []map[string]interface{} {
	var trend []map[string]interface{}
	
	now := time.Now()
	for i := 23; i >= 0; i-- {
		hour := now.Add(-time.Duration(i) * time.Hour)
		trend = append(trend, map[string]interface{}{
			"time":          hour.Format("2006-01-02 15:00"),
			"average_score": 75 + (i%10 - 5), // 模拟数据
			"user_count":    100 + i*5,       // 模拟数据
		})
	}
	
	return trend
}

// 获取低分用户列表
func (api *ScoreAPI) GetLowScoreUsers(c *gin.Context) {
	thresholdStr := c.DefaultQuery("threshold", "30")
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		threshold = 30
	}

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

	// 模拟低分用户数据
	users := []map[string]interface{}{
		{
			"fingerprint":   "abc123def456",
			"score":         25,
			"last_seen":     time.Now().Add(-time.Hour),
			"request_count": 150,
			"risk_level":    "warning",
		},
		{
			"fingerprint":   "xyz789uvw012",
			"score":         15,
			"last_seen":     time.Now().Add(-30*time.Minute),
			"request_count": 200,
			"risk_level":    "danger",
		},
		{
			"fingerprint":   "mno345pqr678",
			"score":         5,
			"last_seen":     time.Now().Add(-15*time.Minute),
			"request_count": 300,
			"risk_level":    "critical",
		},
	}

	// 过滤低于阈值的用户
	var filteredUsers []map[string]interface{}
	for _, user := range users {
		if user["score"].(int) <= threshold {
			filteredUsers = append(filteredUsers, user)
		}
	}

	response := map[string]interface{}{
		"users":       filteredUsers,
		"threshold":   threshold,
		"total":       len(filteredUsers),
		"page":        page,
		"size":        size,
		"total_pages": (len(filteredUsers) + size - 1) / size,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 批量操作用户分数
func (api *ScoreAPI) BatchScoreOperation(c *gin.Context) {
	var req struct {
		Operation    string   `json:"operation" binding:"required"` // "reset", "adjust"
		Fingerprints []string `json:"fingerprints" binding:"required"`
		Adjustment   int      `json:"adjustment,omitempty"`
		Reason       string   `json:"reason" binding:"required"`
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

	if len(req.Fingerprints) > 100 {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "批量操作最多支持100个用户",
		})
		return
	}

	var results []map[string]interface{}
	var successCount, failCount int

	for _, fingerprint := range req.Fingerprints {
		var err error
		
		switch req.Operation {
		case "reset":
			err = api.scorer.ResetUserScore(fingerprint)
		case "adjust":
			// 获取当前分数
			userScore, getErr := api.redisClient.GetUserScore(fingerprint)
			if getErr != nil {
				err = getErr
			} else {
				userScore.Score += req.Adjustment
				if userScore.Score > scorer.DefaultScoringConfig.MaxScore {
					userScore.Score = scorer.DefaultScoringConfig.MaxScore
				}
				if userScore.Score < -50 {
					userScore.Score = -50
				}
				err = api.redisClient.UpdateUserScore(fingerprint, userScore)
			}
		default:
			err = fmt.Errorf("不支持的操作类型: %s", req.Operation)
		}

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
		"operation":     req.Operation,
		"total_count":   len(req.Fingerprints),
		"success_count": successCount,
		"fail_count":    failCount,
		"results":       results,
		"reason":        req.Reason,
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Message: fmt.Sprintf("批量操作完成，成功: %d, 失败: %d", successCount, failCount),
		Data:    response,
	})
}

// 获取用户分数历史
func (api *ScoreAPI) GetUserScoreHistory(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, ConfigResponse{
			Success: false,
			Error:   "用户指纹不能为空",
		})
		return
	}

	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 || hours > 168 { // 最多7天
		hours = 24
	}

	trend, err := api.scorer.GetScoreTrend(fingerprint, hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfigResponse{
			Success: false,
			Error:   "获取分数历史失败: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"fingerprint": fingerprint,
		"hours":       hours,
		"trend":       trend,
		"summary": map[string]interface{}{
			"current_score": trend[len(trend)-1].Score,
			"highest_score": api.getMaxScore(trend),
			"lowest_score":  api.getMinScore(trend),
			"average_score": api.getAverageScore(trend),
		},
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Success: true,
		Data:    response,
	})
}

// 辅助函数：获取最高分数
func (api *ScoreAPI) getMaxScore(trend []scorer.ScoreTrendPoint) int {
	if len(trend) == 0 {
		return 0
	}
	
	max := trend[0].Score
	for _, point := range trend {
		if point.Score > max {
			max = point.Score
		}
	}
	return max
}

// 辅助函数：获取最低分数
func (api *ScoreAPI) getMinScore(trend []scorer.ScoreTrendPoint) int {
	if len(trend) == 0 {
		return 0
	}
	
	min := trend[0].Score
	for _, point := range trend {
		if point.Score < min {
			min = point.Score
		}
	}
	return min
}

// 辅助函数：获取平均分数
func (api *ScoreAPI) getAverageScore(trend []scorer.ScoreTrendPoint) float64 {
	if len(trend) == 0 {
		return 0
	}
	
	total := 0
	for _, point := range trend {
		total += point.Score
	}
	return float64(total) / float64(len(trend))
}

// 注册分数API路由
func (api *ScoreAPI) RegisterRoutes(router *gin.RouterGroup) {
	score := router.Group("/score")
	{
		score.GET("/stats", api.GetScoreStats)
		score.GET("/low-score-users", api.GetLowScoreUsers)
		score.POST("/batch", api.BatchScoreOperation)
		
		score.GET("/:fingerprint", api.GetUserScore)
		score.POST("/:fingerprint/reset", api.ResetUserScore)
		score.POST("/:fingerprint/adjust", api.AdjustUserScore)
		score.GET("/:fingerprint/history", api.GetUserScoreHistory)
	}
}
